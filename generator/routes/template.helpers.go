package routes

import (
	"fmt"

	"github.com/aymerick/raymond"
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/iancoleman/strcase"
)

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

	raymond.RegisterHelper("ifEqual", func(a interface{}, b interface{}, options *raymond.Options) string {
		if raymond.Str(a) == raymond.Str(b) {
			return options.Fn()
		}

		return options.Inverse()
	})

	raymond.RegisterHelper("ifAnyParamRequiresConversion", func(params []definitions.FuncParam, options *raymond.Options) string {
		for _, param := range params {
			if param.TypeMeta.Name != "string" && param.TypeMeta.FullyQualifiedPackage != "" && param.TypeMeta.EntityKind != definitions.AstNodeKindAlias {
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

	raymond.RegisterHelper("OrEqual", func(val1, comp1, val2, comp2 interface{}) bool {
		isEqual1 := (val1 == comp1)
		isEqual2 := (val2 == comp2)
		return isEqual1 || isEqual2
	})
}
