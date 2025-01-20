package extractor

import (
	"fmt"
	"go/types"
)

// Field represents a field in a struct
type Field struct {
	Name string
	Type string
}

// StructInfo represents a struct with its fields
type StructInfo struct {
	Name   string
	Fields []Field
}

type TypeVisitor struct {
	typesByName map[string]*StructInfo
}

func NewTypeVisitor() *TypeVisitor {
	return &TypeVisitor{
		typesByName: make(map[string]*StructInfo),
	}
}

func (v *TypeVisitor) VisitStruct(pkgName, structName string, structType *types.Struct) error {
	fullName := fmt.Sprintf("%s.%s", pkgName, structName)
	if v.typesByName[fullName] != nil {
		return fmt.Errorf("struct %q has already been processed", fullName)
	}
	structInfo := StructInfo{Name: structName}
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		fieldType := field.Type()

		switch t := fieldType.(type) {
		case *types.Pointer:
			// Raise error for pointer fields.
			return fmt.Errorf("field %q in struct %q is a pointer, which is not allowed", field.Name(), structName)
		case *types.Named:
			// Check if the named type is a struct.
			if underlying, ok := t.Underlying().(*types.Struct); ok {
				// Recursively process the nested struct.
				err := v.VisitStruct(pkgName, t.Obj().Name(), underlying)
				if err != nil {
					return err
				}
			}

			// Add the field as a reference to another struct.
			structInfo.Fields = append(structInfo.Fields, Field{
				Name: field.Name(),
				Type: t.Obj().Name(),
			})
		default:
			// Add primitive field types.
			structInfo.Fields = append(structInfo.Fields, Field{
				Name: field.Name(),
				Type: fieldType.String(),
			})
		}
	}

	v.typesByName[fullName] = &structInfo
	return nil
}

// GetStructs returns the list of processed structs.
func (v *TypeVisitor) GetStructs() []*StructInfo {
	structs := []*StructInfo{}
	for _, value := range v.typesByName {
		structs = append(structs, value)
	}
	return structs
}
