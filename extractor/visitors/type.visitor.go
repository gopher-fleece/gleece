package visitors

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor"
	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"golang.org/x/tools/go/packages"
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
	packages    []*packages.Package
	typesByName map[string]*definitions.StructMetadata
}

type StructAttributeHolders struct {
	StructHolder annotations.AnnotationHolder
	FieldHolders map[string]*annotations.AnnotationHolder
}

func NewTypeVisitor(packages []*packages.Package) *TypeVisitor {
	return &TypeVisitor{
		packages:    packages,
		typesByName: make(map[string]*definitions.StructMetadata),
	}
}

func (v *TypeVisitor) VisitStruct(fullPackageName string, structName string, structType *types.Struct) error {
	fullName := fmt.Sprintf("%s.%s", fullPackageName, structName)
	if v.typesByName[fullName] != nil {
		return nil // Already processed, ignore
	}

	attributeHolders, err := v.getAttributeHolders(fullPackageName, structName)
	if err != nil {
		return err
	}

	structInfo := definitions.StructMetadata{
		Name:                  structName,
		FullyQualifiedPackage: fullPackageName,
		Description:           attributeHolders.StructHolder.GetDescription(),
		Deprecation:           getDeprecationOpts(attributeHolders.StructHolder),
	}

	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		fieldType := field.Type()
		tag := structType.Tag(i)

		// Skip embedded error fields.
		if field.Name() == "error" && field.Type().String() == "error" {
			continue
		}

		var fieldTypeString string

		switch t := fieldType.(type) {
		case *types.Pointer:
			// Raise error for pointer fields.
			return fmt.Errorf("field %q in struct %q is a pointer, which is not allowed", field.Name(), structName)
		case *types.Array, *types.Slice:
			fieldTypeString = extractor.GetIterableElementType(t)
		case *types.Named:
			if extractor.IsAliasType(t) {
				aliasModel := createAliasModel(t, tag)
				structInfo.Fields = append(structInfo.Fields, aliasModel)
				continue // A bit ugly. Need to clean this up...
			} else {
				// Check if the named type is a struct.
				if underlying, ok := t.Underlying().(*types.Struct); ok {
					// Recursively process the nested struct.
					err := v.VisitStruct(fullPackageName, t.Obj().Name(), underlying)
					if err != nil {
						return err
					}
				} else {
					return fmt.Errorf(
						"node '%s' is of kind 'Named' but is neither an alias-like nor struct and cannot be used",
						t.Obj().Name(),
					)
				}
			}

			// Add the field as a reference to another struct.
			fieldTypeString = t.Obj().Name()
		default:
			// Primitive field
			fieldTypeString = fieldType.String()
		}

		fieldMeta := definitions.FieldMetadata{
			Name: field.Name(),
			Type: fieldTypeString,
			Tag:  tag,
		}

		fieldAttr := attributeHolders.FieldHolders[field.Name()]
		if fieldAttr != nil {
			fieldMeta.Description = fieldAttr.GetDescription()
			deprecationOpts := getDeprecationOpts(*fieldAttr)
			fieldMeta.Deprecation = &deprecationOpts
		}

		structInfo.Fields = append(structInfo.Fields, fieldMeta)
	}

	v.typesByName[fullName] = &structInfo
	return nil
}

func (v *TypeVisitor) getAttributeHolders(fullPackageName string, structName string) (StructAttributeHolders, error) {
	holders := StructAttributeHolders{FieldHolders: make(map[string]*annotations.AnnotationHolder)}

	relevantPackage := extractor.FilterPackageByFullName(v.packages, fullPackageName)
	if relevantPackage == nil {
		return holders, fmt.Errorf(
			"could not find package object for '%s' whilst looking for struct '%s'",
			fullPackageName,
			structName,
		)
	}

	genDecl := extractor.FindGenDeclByName(relevantPackage, structName)
	if genDecl == nil {
		return holders, fmt.Errorf("could not find GenDecl node for struct '%s' in package '%s'", structName, fullPackageName)
	}

	structNode := extractor.GetStructFromGenDecl(genDecl)
	if structNode == nil {
		return holders, fmt.Errorf(
			"could not obtain StructType node from the GenDecl of struct '%s' in package '%s'",
			structName,
			fullPackageName,
		)
	}

	if genDecl.Doc != nil && genDecl.Doc.List != nil && len(genDecl.Doc.List) > 0 {
		structAttributes, err := annotations.NewAnnotationHolder(extractor.MapDocListToStrings(genDecl.Doc.List), annotations.CommentSourceSchema)
		if err != nil {
			logger.Error("Could not create an attribute holder for struct '%s' - %v", structName, err)
			return holders, err
		}

		holders.StructHolder = structAttributes
	}

	for _, field := range structNode.Fields.List {
		if field.Doc != nil && field.Doc.List != nil && len(field.Doc.List) > 0 {
			if len(field.Names) > 1 {
				names := []string{}
				for _, nameIdent := range field.Names {
					names = append(names, nameIdent.Name)
				}
				return holders, fmt.Errorf(
					"field/s [%s] on struct %s has more than one name i.e., is a multi-var declaration. This is not currently supported",
					strings.Join(names, ", "),
					structName,
				)
			}

			fieldName := field.Names[0].Name
			fieldHolder, err := annotations.NewAnnotationHolder(extractor.MapDocListToStrings(field.Doc.List), annotations.CommentSourceProperty)
			if err != nil {
				logger.Error("Could not create an attribute holder for field %s on struct '%s' - %v", fieldName, structName, err)
				return holders, err
			}

			holders.FieldHolders[fieldName] = &fieldHolder
		}
	}

	return holders, nil
}

// GetStructs returns the list of processed structs.
func (v *TypeVisitor) GetStructs() []definitions.StructMetadata {
	models := []definitions.StructMetadata{}
	for _, value := range v.typesByName {
		models = append(models, *value)
	}
	return models
}

func getDeprecationOpts(attributes annotations.AnnotationHolder) definitions.DeprecationOptions {
	deprecationAttr := attributes.GetFirst(annotations.AttributeDeprecated)
	if deprecationAttr == nil {
		return definitions.DeprecationOptions{}
	}

	return definitions.DeprecationOptions{
		Deprecated:  true,
		Description: deprecationAttr.Description,
	}
}

func createAliasModel(node *types.Named, tag string) definitions.FieldMetadata {
	return definitions.FieldMetadata{
		Name: node.Obj().Name(),
		Type: extractor.GetUnderlyingTypeName(node.Underlying()),
		Tag:  tag,
	}
}
