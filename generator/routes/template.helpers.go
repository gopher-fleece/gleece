package routes

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/aymerick/raymond"
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/iancoleman/strcase"
)

var collapsibleSpaceRegex = regexp.MustCompile("[\t\r\n]")

// SplitBracketName splits a string into brackets part and name part
// Example: "[][][]Name" -> "[][][]", "Name"
func splitSliceBracket(input string) (brackets string, name string) {
	// Find the position where brackets end
	pos := 0
	for pos+1 < len(input) {
		if input[pos] == '[' && input[pos+1] == ']' {
			pos += 2
		} else {
			break
		}
	}

	// Split the string
	return input[:pos], input[pos:]
}

func registerHandlebarsHelpers() {
	raymond.RegisterHelper("SlicePrefix", func(arg string) string {
		brackets, _ := splitSliceBracket(arg)
		return brackets
	})

	raymond.RegisterHelper("SliceSlice", func(arg string) string {
		_, name := splitSliceBracket(arg)
		return name
	})

	raymond.RegisterHelper("ToSnakeCase", func(arg string) string {
		return strcase.ToSnake(arg)
	})

	raymond.RegisterHelper("ToLowerCamel", func(arg string) string {
		return strcase.ToLowerCamel(arg)
	})

	raymond.RegisterHelper("ToUpperCamel", func(arg string) string {
		return strcase.ToCamel(arg)
	})

	raymond.RegisterHelper("UnwrapArrayTypeRecursive", func(arg string) string {
		return common.UnwrapArrayTypeString(arg)
	})

	raymond.RegisterHelper("LastTypeNameEquals", func(types []definitions.FuncReturnValue, value string, options *raymond.Options) string {
		if len(types) <= 0 {
			panic("LastTypeNameEquals received a 0-length array")
		}

		if types[len(types)-1].Name == value {
			return options.Fn()
		}

		return options.Inverse()
	})

	raymond.RegisterHelper("ifEqual", func(a any, b any, options *raymond.Options) string {
		if raymond.Str(a) == raymond.Str(b) {
			return options.Fn()
		}

		return options.Inverse()
	})

	raymond.RegisterHelper("ifAnyParamRequiresConversion", func(params []definitions.FuncParam, options *raymond.Options) string {
		for _, param := range params {
			if param.IsContext {
				// The CTX param is not come from the http "input" and should not be converted from anything
				continue
			}
			if param.TypeMeta.Name != "string" && param.TypeMeta.PkgPath != "" && param.TypeMeta.SymbolKind != common.SymKindEnum {
				// Currently, only 'string' parameters don't undergo any validation
				return options.Fn()
			}
		}

		return options.Inverse()
	})

	raymond.RegisterHelper("LastTypeIsByAddress", func(types []definitions.FuncReturnValue, options *raymond.Options) string {
		if len(types) <= 0 {
			panic("LastTypeIsByAddress received a 0-length array")
		}

		if types[len(types)-1].IsByAddress {
			return options.Fn()
		}

		return options.Inverse()
	})

	raymond.RegisterHelper("GetLastTyeFullyQualified", func(types []definitions.FuncReturnValue) string {
		if len(types) <= 0 {
			panic("GetLastTyeFullyQualified received a 0-length array")
		}

		last := types[len(types)-1]
		return fmt.Sprintf("Response%d%s.%s", last.UniqueImportSerial, last.Name, last.Name)
	})

	raymond.RegisterHelper("OrEqual", func(val1, comp1, val2, comp2 any) bool {
		isEqual1 := (val1 == comp1)
		isEqual2 := (val2 == comp2)
		return isEqual1 || isEqual2
	})

	raymond.RegisterHelper("CollapseMultiline", func(options *raymond.Options) string {
		content := options.Fn()
		return string(collapsibleSpaceRegex.ReplaceAllString(content, ""))
	})

	raymond.RegisterHelper("UnpackImportsMap", func(context map[string][]string) string {
		var importsBuilder strings.Builder

		// Step 1: Collect and sort package paths
		pkgPaths := make([]string, 0, len(context))
		for pkgPath := range context {
			pkgPaths = append(pkgPaths, pkgPath)
		}
		sort.Strings(pkgPaths)

		// Step 2: Sort aliases and emit
		for _, pkgPath := range pkgPaths {
			aliases := context[pkgPath]
			if len(aliases) == 0 {
				continue
			}
			sort.Strings(aliases)

			for _, alias := range aliases {
				importsBuilder.WriteString(fmt.Sprintf("%s \"%s\"\n", alias, pkgPath))
			}
		}

		return importsBuilder.String()
	})

}
