package visitors

import (
	"fmt"
	"go/ast"
	"go/types"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/extractor/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"golang.org/x/tools/go/packages"
)

type ControllerWithStructMeta struct {
	Controller definitions.ControllerMetadata
	StructMeta metadata.StructMeta
}

type NodeWithComments struct {
	Doc *ast.CommentGroup
}

type RecursiveTypeVisitor struct {
	BaseVisitor

	// The file currently being worked on
	currentSourceFile *ast.File

	// The last-encountered GenDecl.
	//
	// Documentation may be placed on TypeDecl or their parent GenDecl so we track these,
	// in case we need to fetch the docs from the TypeDecl's parent.
	currentGenDecl *ast.GenDecl

	// The current Gleece Controller being processed
	// currentController *definitions.ControllerMetadata

	// currentFVersion gast.FileVersion
}

// NewTypeVisitor Instantiates a new type visitor.
func NewTypeVisitor(context *VisitContext) (*RecursiveTypeVisitor, error) {
	visitor := RecursiveTypeVisitor{}
	err := visitor.initialize((context))
	return &visitor, err
}

func (v *RecursiveTypeVisitor) SetCurrentFile(file *ast.File) {
	// This breaks encapsulation and is a hack...
	v.currentSourceFile = file
}

func (v *RecursiveTypeVisitor) Visit(node ast.Node) ast.Visitor {
	switch currentNode := node.(type) {
	case *ast.File:
		// Update the current file when visiting an *ast.File node
		v.currentSourceFile = currentNode
	case *ast.GenDecl:
		v.currentGenDecl = currentNode
	case *ast.TypeSpec:
		// Check if it's a struct and if it embeds GleeceController
		if _, isStruct := currentNode.Type.(*ast.StructType); isStruct {
			_, err := v.VisitStructType(v.currentSourceFile, currentNode)
			if err != nil {
				v.setLastError(err)
				return v
			}
		}
	}
	return v
}

func (v *RecursiveTypeVisitor) VisitStructType(file *ast.File, node *ast.TypeSpec) (*metadata.StructMeta, error) {
	fVersion, err := v.context.MetadataCache.GetFileVersion(file, v.context.ArbitrationProvider.Pkg().FSet())
	if err != nil {
		return nil, v.frozenError(err)
	}

	pkg, err := v.context.ArbitrationProvider.Pkg().GetPackageForFile(file)
	if err != nil {
		return nil, v.frozenError(err)
	}

	if pkg == nil {
		return nil, v.getFrozenError(
			"could not determine package for file %s",
			gast.GetAstFileName(
				v.context.ArbitrationProvider.Pkg().FSet(),
				file,
			),
		)
	}

	holder, err := v.getComments(node.Doc, v.currentGenDecl)
	if err != nil {
		return nil, v.frozenError(err)
	}

	structType, isStruct := node.Type.(*ast.StructType)
	if !isStruct {
		return nil, v.getFrozenError("non-struct node '%v' was provided to VisitStructType", node.Name.Name)
	}

	fields, err := v.buildFields(pkg, file, structType)
	if err != nil {
		return nil, v.frozenError(err)
	}

	structMeta := &metadata.StructMeta{
		SymNodeMeta: metadata.SymNodeMeta{
			Name:        node.Name.Name,
			Node:        node,
			SymbolKind:  common.SymKindStruct,
			PkgPath:     pkg.PkgPath,
			Annotations: holder,
			FVersion:    fVersion,
		},
		Fields: fields,
	}

	v.context.MetadataCache.AddStruct(structMeta)

	return structMeta, nil
}

func (v *RecursiveTypeVisitor) VisitEnumType(
	pkg *packages.Package,
	file *ast.File,
	fVersion *gast.FileVersion,
	node *ast.TypeSpec,
) (metadata.EnumMeta, error) {
	typeName, err := gast.GetTypeNameOrError(pkg, node.Name.Name)
	if err != nil {
		return metadata.EnumMeta{}, err
	}

	meta, err := v.extractEnumAliasType(pkg, fVersion, node, typeName)
	if err != nil {
		return meta, err
	}

	v.context.MetadataCache.AddEnum(&meta)
	return meta, nil
}
func (v *RecursiveTypeVisitor) extractEnumAliasType(
	pkg *packages.Package,
	fVersion *gast.FileVersion,
	node ast.Node,
	typeName *types.TypeName,
) (metadata.EnumMeta, error) {

	basic, isBasicType := typeName.Type().Underlying().(*types.Basic)
	if !isBasicType {
		return metadata.EnumMeta{}, fmt.Errorf("type %s is not a basic type", typeName.Name())
	}

	kind, err := metadata.NewEnumValueKind(basic.Kind())
	if err != nil {
		return metadata.EnumMeta{}, err
	}

	enum := metadata.EnumMeta{
		SymNodeMeta: metadata.SymNodeMeta{
			Name:       typeName.Name(),
			Node:       node,
			SymbolKind: common.SymKindEnum,
			PkgPath:    pkg.PkgPath,
			FVersion:   fVersion,
		},
		ValueKind: kind,
		Values:    []metadata.EnumValueDefinition{},
	}

	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		constVal, isConst := obj.(*types.Const)
		if !isConst || !types.Identical(constVal.Type(), typeName.Type()) {
			continue
		}

		val := gast.ExtractConstValue(basic.Kind(), constVal)
		if val == nil {
			continue
		}

		// Attempt to find the AST node
		var constNode ast.Node
		if v.context != nil && v.context.ArbitrationProvider != nil {
			constNode = gast.FindConstSpecNode(pkg, constVal.Name())
		}

		// Attempt to extract annotations
		var holder *annotations.AnnotationHolder
		if constNode != nil {
			comments := gast.GetCommentsFromNode(constNode)
			if h, err := annotations.NewAnnotationHolder(comments, annotations.CommentSourceProperty); err == nil {
				holder = &h
			} else {
				return metadata.EnumMeta{}, err
			}
		}

		enum.Values = append(enum.Values, metadata.EnumValueDefinition{
			SymNodeMeta: metadata.SymNodeMeta{
				Name:        constVal.Name(),
				Node:        constNode,
				SymbolKind:  common.SymKindEnumValue,
				PkgPath:     pkg.PkgPath,
				Annotations: holder,
				FVersion:    fVersion,
			},
			Value: val,
		})
	}

	return enum, nil
}

func (v *RecursiveTypeVisitor) VisitField(
	pkg *packages.Package,
	file *ast.File,
	field *ast.Field,
) ([]metadata.FieldMeta, error) {
	var results []metadata.FieldMeta

	// Get doc/comments on the field
	holder, err := v.getComments(field.Doc, nil)
	if err != nil {
		return nil, v.frozenError(err)
	}

	// Determine embedded vs named
	isEmbedded := len(field.Names) == 0

	// This supports multiple names: `Name1, Name2 string`
	// or embedded anonymous fields which have no name
	names := field.Names
	if isEmbedded {
		// For embedded fields, Names is nil – we still need to resolve the type name
		ident := gast.GetIdentFromExpr(field.Type)
		if ident != nil {
			names = []*ast.Ident{{Name: ident.Name}}
		}
	}

	for _, nameIdent := range names {
		name := nameIdent.Name

		typeIdentName := ""
		typeIdent := gast.GetIdentFromExpr(field.Type)
		if typeIdent != nil {
			typeIdentName = typeIdent.Name
		}

		importType, err := v.context.ArbitrationProvider.Ast().GetImportType(file, field.Type)
		if err != nil {
			return nil, v.frozenError(err)
		}

		// Create TypeUsageMeta (AST part only)
		typeUsage := metadata.TypeUsageMeta{
			SymNodeMeta: metadata.SymNodeMeta{
				Name:        typeIdentName,
				Node:        field.Type,
				PkgPath:     pkg.PkgPath,
				SymbolKind:  common.SymKindField,
				Annotations: holder,
			},
			TypeRefKey: graphs.SymbolKeyFor(field.Type, nil /*THIS NEEDS TO BE THE AST FILE FOR THE UNDERLYING TYPE?*/),
			Import:     importType,
		}

		fieldMeta := metadata.FieldMeta{
			SymNodeMeta: metadata.SymNodeMeta{
				Name:        name,
				Node:        field,
				PkgPath:     pkg.PkgPath,
				SymbolKind:  common.SymKindField,
				Annotations: holder,
			},
			Type:       typeUsage,
			IsEmbedded: isEmbedded,
		}

		results = append(results, fieldMeta)

		err = v.processTypeUsage(pkg, file, field, typeUsage)
		if err != nil {
			return results, v.frozenError(err)
		}
	}

	return results, nil
}

func (v *RecursiveTypeVisitor) buildFields(
	pkg *packages.Package,
	file *ast.File,
	node *ast.StructType,
) ([]metadata.FieldMeta, error) {
	var results []metadata.FieldMeta

	for _, field := range node.Fields.List {
		fields, err := v.VisitField(pkg, file, field)
		if err != nil {
			return results, err
		}
		results = append(results, fields...)
	}

	return results, nil
}

func (v *RecursiveTypeVisitor) processTypeUsage(
	pkg *packages.Package,
	file *ast.File,
	field *ast.Field,
	typeUsage metadata.TypeUsageMeta,
) error {
	// Don't dive into built-in types
	if typeUsage.IsUniverseType() {
		return nil
	}

	// This is the identifier you need to resolve
	underlyingIdent := gast.GetIdentFromExpr(field.Type)
	if underlyingIdent == nil {
		return nil // No ident to resolve
	}

	if len(field.Names) > 1 {
		return v.getFrozenError("field '%v' is a multi-variable declaration which is currently not supported", field.Names)
	}

	underlyingTypePkg, underlyingAstFile, typeSpec, err := gast.ResolveTypeSpecFromField(
		pkg,
		file,
		field,
		v.context.ArbitrationProvider.Pkg().GetPackage,
	)

	if err != nil {
		return v.frozenError(err)
	}

	fVersion, err := v.context.MetadataCache.GetFileVersion(underlyingAstFile, underlyingTypePkg.Fset)
	if err != nil {
		return v.frozenError(err)
	}

	key := graphs.SymbolKeyFor(typeSpec, fVersion)
	if v.context.MetadataCache.HasVisited(key) {
		return nil
	}

	// Ok, recurse
	err = v.visitTypeSpec(underlyingTypePkg, underlyingAstFile, fVersion, typeSpec)
	if err != nil {
		return v.frozenError(err)
	}

	return nil
}

func (v *RecursiveTypeVisitor) visitTypeSpec(
	pkg *packages.Package,
	file *ast.File,
	fVersion *gast.FileVersion,
	spec *ast.TypeSpec,
) error {
	var err error

	switch t := spec.Type.(type) {
	case *ast.StructType:
		_, err := v.VisitStructType(file, spec)
		if err != nil {
			return v.frozenError(err)
		}
	case *ast.Ident:
		// Enum-like: alias of primitive (string, int, etc)
		// Optional: Check constants with the same name prefix to confirm it's "really" enum-like
		_, err := v.VisitEnumType(pkg, file, fVersion, spec)
		if err != nil {
			return v.frozenError(err)
		}
	default:
		err = fmt.Errorf("unhandled TypeSpec type: %T", t)
	}

	return err
}

func (v *RecursiveTypeVisitor) getComments(onNodeDoc *ast.CommentGroup, nodeGenDecl *ast.GenDecl) (*annotations.AnnotationHolder, error) {
	var commentSource *ast.CommentGroup
	if onNodeDoc != nil {
		commentSource = onNodeDoc
	} else {
		if nodeGenDecl != nil {
			commentSource = nodeGenDecl.Doc
		}
	}

	if commentSource != nil {
		comments := gast.MapDocListToStrings(commentSource.List)
		holder, err := annotations.NewAnnotationHolder(comments, annotations.CommentSourceController)
		return &holder, err
	}

	return nil, nil
}
