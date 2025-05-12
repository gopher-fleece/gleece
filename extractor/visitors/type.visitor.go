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

// TypeVisitor handles recursive 'visitation' of types (structs, enums, aliases)
type TypeVisitor struct {
	// A facade for managing packages.Package
	packagesFacade *arbitrators.PackagesFacade

	// An arbitrator providing logic for working with Go's AST
	astArbitrator *arbitrators.AstArbitrator

	// A map of struct names to metadata
	structsByName map[string]*definitions.StructMetadata

	// A map of 'enum' names to metadata
	enumsByName map[string]*definitions.EnumMetadata
}

// StructAttributeHolders serves as a container for tracking AnnotationHolders for a struct and its fields
type StructAttributeHolders struct {
	StructHolder annotations.AnnotationHolder
	FieldHolders map[string]*annotations.AnnotationHolder
}

// NewTypeVisitor Initializes a new visitor
func NewTypeVisitor(packagesFacade *arbitrators.PackagesFacade, astArbitrator *arbitrators.AstArbitrator) *TypeVisitor {
	return &TypeVisitor{
		packagesFacade: packagesFacade,
		astArbitrator:  astArbitrator,
		structsByName:  make(map[string]*definitions.StructMetadata),
		enumsByName:    make(map[string]*definitions.EnumMetadata),
	}
}

// VisitStruct dives into the given struct in the given package and recursively parses it to obtain metadata,
// storing the results in the visitor's internal fields.
func (v *TypeVisitor) VisitStruct(fullPackageName string, structName string, structType *types.Struct) error {
	fullStructName := fmt.Sprintf("%s.%s", fullPackageName, structName)
	if v.structsByName[fullStructName] != nil {
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

	// Iterate the struct's fields
	for i := range structType.NumFields() {
		field := structType.Field(i)
		fieldType := field.Type()
		tag := structType.Tag(i)

		// Skip embedded 'error' fields.
		if field.Name() == "error" && field.Type().String() == "error" {
			continue
		}

		var fieldTypeString string

		switch t := fieldType.(type) {
		case *types.Pointer:
			// Raise error for pointer fields.
			return fmt.Errorf("field %q in struct %q is a pointer, which is not allowed", field.Name(), structName)

		case *types.Slice, *types.Array:
			// Go's rigid typing makes reuse pretty difficult...
			iterable, ok := t.(definitions.Iterable)
			if !ok {
				return fmt.Errorf("expected slice or array to implement Iterable, got %T", t)
			}

			// Field type string is for the parent model's metadata
			fieldTypeString = extractor.GetIterableElementType(t)

			// Dive into the slice and recurse into nested structs, if required
			err := v.processIterableField(field, iterable, fieldTypeString, &structInfo, tag)
			if err != nil {
				return err
			}

		case *types.Named:
			isEnumOrAlias, err := v.processNamedEntity(t, &structInfo, tag)
			if err != nil {
				return err
			}
			if isEnumOrAlias {
				// Enums and aliases have already been processed at this point - continue to the next field.
				continue
			}
			// Add the field as a reference to another struct.
			fieldTypeString = t.Obj().Name()

		default:
			// Primitive field
			fieldTypeString = fieldType.String()
		}

		fieldMeta := definitions.FieldMetadata{
			Name:       field.Name(),
			Type:       fieldTypeString,
			Tag:        tag,
			IsEmbedded: field.Anonymous(),
		}

		fieldAttr := attributeHolders.FieldHolders[field.Name()]
		if fieldAttr != nil {
			fieldMeta.Description = fieldAttr.GetDescription()
			deprecationOpts := getDeprecationOpts(*fieldAttr)
			fieldMeta.Deprecation = &deprecationOpts
		} else {
			fieldMeta.Deprecation = &definitions.DeprecationOptions{}
		}

		structInfo.Fields = append(structInfo.Fields, fieldMeta)
	}

	v.structsByName[fullStructName] = &structInfo
	return nil
}

func (v *TypeVisitor) processIterableField(
	field *types.Var,
	fieldType definitions.Iterable,
	sliceElementTypeNameString string,
	structInfo *definitions.StructMetadata,
	fieldTag string,
) error {
	if extractor.IsBasic(fieldType) {
		return nil
	}

	pkg := extractor.GetPackageOwnerOfType(fieldType)
	if pkg == nil {
		baseTypeName := sliceElementTypeNameString
		if strings.HasPrefix(sliceElementTypeNameString, "[]") {
			baseTypeName = sliceElementTypeNameString[2:]
		}

		if extractor.IsUniverseType(baseTypeName) {
			// If there's no owner package, this might be a universe type. If so, just ignore it
			return nil
		}

		// Otherwise, we've a problem.
		return fmt.Errorf(
			"could not deduce package for the type of field %s of type %s on struct %s",
			field.Name(),
			sliceElementTypeNameString,
			structInfo.Name,
		)
	}

	if named, ok := fieldType.Elem().(*types.Named); ok {
		_, err := v.processNamedEntity(named, structInfo, fieldTag)
		if err != nil {
			return err
		}
	}

	return nil
}

// processNamedEntity receives a Named node and translates it into an enum/alias model.
// Returns a boolean indicating whether the node is considered an enum/alias which indicates
// whether it should be appended to the visitor's struct list or not.
//
// To clarify, an enum/alias is NOT appended to the struct list.
func (v *TypeVisitor) processNamedEntity(
	node *types.Named,
	structInfo *definitions.StructMetadata,
	fieldTag string,
) (bool, error) {
	// If the type's an enum.
	// This is pretty nasty - there needs to be a stricter separation between simple aliases and actual enums.
	if extractor.IsAliasType(node) {
		aliasModel := createAliasModel(node, fieldTag)
		err := v.visitNestedEnum(node, aliasModel)
		if err != nil {
			return true, err
		}

		structInfo.Fields = append(structInfo.Fields, aliasModel)
		return true, nil
	}

	// Check if the named type is a struct.
	if underlying, ok := node.Underlying().(*types.Struct); ok {
		// Recursively process the nested struct.
		nestedPackageName, err := v.packagesFacade.GetPackageNameByNamedEntity(node)
		if err != nil {
			return false, err
		}

		if len(nestedPackageName) == 0 {
			return false, fmt.Errorf("could not determine package for named entity '%s'", node.Obj().Name())
		}

		err = v.VisitStruct(nestedPackageName, node.Obj().Name(), underlying)
		return false, err
	}

	return false, fmt.Errorf(
		"node '%s' is of kind 'Named' but is neither an alias-like nor struct and cannot be used",
		node.Obj().Name(),
	)
}

func (v *TypeVisitor) VisitEnum(enumName string, model definitions.TypeMetadata) error {
	if v.enumsByName[enumName] != nil {
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

	v.enumsByName[enumName] = enumModel
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

			if len(field.Names) == 0 {
				// Embedded/Anonymous field, skip
				continue
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
	for _, value := range v.structsByName {
		models = append(models, *value)
	}
	return models
}

func (v *TypeVisitor) GetEnums() []definitions.EnumMetadata {
	models := []definitions.EnumMetadata{}
	for _, value := range v.enumsByName {
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
	if v.enumsByName[cleanedModelName] != nil {
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

	v.enumsByName[cleanedModelName] = enumModel
	return nil
}
