package extractor

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"
)

func DoesStructEmbedType(pkg *packages.Package, structName string, embeddedStructFullPackage string, embeddedStructName string) (bool, error) {
	// Look for the struct definition in the package
	obj := pkg.Types.Scope().Lookup(structName)
	if obj == nil {
		return false, fmt.Errorf("struct '%s' not found in package '%s'", structName, pkg.PkgPath)
	}

	// Ensure it's a named type (i.e., a struct)
	named, ok := obj.Type().(*types.Named)
	if !ok {
		return false, fmt.Errorf("type '%s' is not a named type", structName)
	}

	// Get the underlying struct type
	structType, ok := named.Underlying().(*types.Struct)
	if !ok {
		return false, fmt.Errorf("type '%s' is not a struct", structName)
	}

	// Check if any field in the struct embeds the target type
	var expectedFullyQualifiedName string
	if len(embeddedStructFullPackage) > 0 {
		expectedFullyQualifiedName = embeddedStructFullPackage + "." + embeddedStructName
	} else {
		expectedFullyQualifiedName = embeddedStructName
	}

	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		if field.Embedded() {
			// Check if the field type matches the embedded struct we are looking for
			fieldType := field.Type().String()
			// Check if the type matches the embedded struct (full package path + type name)
			if fieldType == expectedFullyQualifiedName {
				return true, nil
			}
		}
	}

	return false, nil
}
