package visitors

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor"
	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/extractor/arbitrators"
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
	// A facade for managing packages.Package
	packagesFacade *arbitrators.PackagesFacade

	astArbitrator *arbitrators.AstArbitrator

	typesByName map[string]*definitions.StructMetadata
	enumByName  map[string]*definitions.EnumMetadata
}

type StructAttributeHolders struct {
	StructHolder annotations.AnnotationHolder
	FieldHolders map[string]*annotations.AnnotationHolder
}

func NewTypeVisitor(packagesFacade *arbitrators.PackagesFacade, astArbitrator *arbitrators.AstArbitrator) *TypeVisitor {
	return &TypeVisitor{
		packagesFacade: packagesFacade,
		astArbitrator:  astArbitrator,
		typesByName:    make(map[string]*definitions.StructMetadata),
		enumByName:     make(map[string]*definitions.EnumMetadata),
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
				err := v.visitNestedEnum(t, aliasModel)
				if err != nil {
					return err
				}

				structInfo.Fields = append(structInfo.Fields, aliasModel)
				continue // A bit ugly. Need to clean this up...
			} else {
				// Check if the named type is a struct.
				if underlying, ok := t.Underlying().(*types.Struct); ok {
					// Recursively process the nested struct.
					nestedPackageName, err := v.packagesFacade.GetPackageNameByNamedEntity(t)
					if err != nil {
						return err
					}

					if len(nestedPackageName) == 0 {
						return fmt.Errorf("could not determine package for named entity '%s'", t.Obj().Name())
					}

					err = v.VisitStruct(nestedPackageName, t.Obj().Name(), underlying)
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

func (v *TypeVisitor) VisitEnum(enumName string, model definitions.TypeMetadata) error {
	if v.enumByName[enumName] != nil {
		return nil // Already processed, ignore
	}

	enumModel := &definitions.EnumMetadata{
		Name:                  model.Name,
		FullyQualifiedPackage: model.FullyQualifiedPackage,
		Description:           model.Description,
		Values:                model.AliasMetadata.Values,
		Type:                  model.AliasMetadata.AliasType,
		// Deprecation         ?
	}

	v.enumByName[enumName] = enumModel
	return nil
}

func (v *TypeVisitor) getAttributeHolderFromEntityGenDecl(pkg *packages.Package, name string) (annotations.AnnotationHolder, error) {
	genDecl := extractor.FindGenDeclByName(pkg, name)
	if genDecl == nil {
		return annotations.AnnotationHolder{},
			fmt.Errorf("could not find GenDecl node for struct '%s.%s'", pkg.PkgPath, name)
	}

	if genDecl.Doc == nil || genDecl.Doc.List == nil || len(genDecl.Doc.List) <= 0 {
		return annotations.AnnotationHolder{}, nil
	}

	holder, err := annotations.NewAnnotationHolder(
		extractor.MapDocListToStrings(genDecl.Doc.List),
		annotations.CommentSourceSchema,
	)

	if err != nil {
		logger.Error("Could not create an attribute holder for struct '%s.%s' - %v", pkg.PkgPath, name, err)
		return annotations.AnnotationHolder{}, err
	}

	return holder, nil
}

func (v *TypeVisitor) getAttributeHolders(fullPackageName string, structName string) (StructAttributeHolders, error) {
	holders := StructAttributeHolders{FieldHolders: make(map[string]*annotations.AnnotationHolder)}

	relevantPackage, err := v.packagesFacade.GetPackage(fullPackageName)
	if err != nil {
		return holders, err
	}

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

func (v *TypeVisitor) GetEnums() []definitions.EnumMetadata {
	models := []definitions.EnumMetadata{}
	for _, value := range v.enumByName {
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
	name := node.Obj().Name()

	underlying := node.Underlying()

	typeName := extractor.GetUnderlyingTypeName(underlying)

	// In case of an alias to a primitive type, we need to use the alias name.
	_, isBasic := underlying.(*types.Basic)
	if isBasic && name != typeName {
		typeName = name
	}

	return definitions.FieldMetadata{
		Name: name,
		Type: typeName,
		Tag:  tag,
	}
}

func (v *TypeVisitor) visitNestedEnum(t *types.Named, aliasModel definitions.FieldMetadata) error {
	pkg, err := v.packagesFacade.GetPackageByTypeName(t.Obj())
	if err != nil {
		return err
	}

	if pkg == nil {
		return fmt.Errorf("could not obtain package for entity '%v'", t)
	}

	cleanedModelName := common.UnwrapArrayTypeString(aliasModel.Name)
	if v.enumByName[cleanedModelName] != nil {
		return nil // Already processed, ignore
	}

	typeName, err := extractor.GetTypeNameOrError(pkg, aliasModel.Name)
	if err != nil {
		return err
	}

	aliasMetadata, err := v.astArbitrator.ExtractAliasType(pkg, typeName)
	if err != nil {
		return err
	}

	holder, err := v.getAttributeHolderFromEntityGenDecl(pkg, cleanedModelName)
	if err != nil {
		return err
	}

	enumModel := &definitions.EnumMetadata{
		Name:                  cleanedModelName,
		FullyQualifiedPackage: pkg.PkgPath,
		Description:           holder.GetDescription(),
		Values:                aliasMetadata.Values,
		Type:                  aliasMetadata.AliasType,
		// Deprecation         ?
	}

	v.enumByName[cleanedModelName] = enumModel
	return nil
}
