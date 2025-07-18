package visitors

import (
	"fmt"
	"go/ast"
	"go/types"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
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
			_, _, err := v.VisitStructType(v.currentSourceFile, currentNode)
			if err != nil {
				v.setLastError(err)
				return v
			}
		}
	}
	return v
}

func (v *RecursiveTypeVisitor) VisitStructType(
	file *ast.File,
	node *ast.TypeSpec,
) (*metadata.StructMeta, graphs.SymbolKey, error) {
	fVersion, err := v.context.MetadataCache.GetFileVersion(file, v.context.ArbitrationProvider.Pkg().FSet())
	if err != nil {
		return nil, graphs.SymbolKey{}, v.frozenError(err)
	}

	// Check cache first
	symKey := graphs.NewSymbolKey(node, fVersion)
	cached := v.context.MetadataCache.GetStruct(symKey)
	if cached != nil {
		return cached, symKey, nil
	}

	pkg, err := v.context.ArbitrationProvider.Pkg().GetPackageForFile(file)
	if err != nil {
		return nil, graphs.SymbolKey{}, v.frozenError(err)
	}

	if pkg == nil {
		return nil, graphs.SymbolKey{}, v.getFrozenError(
			"could not determine package for file %s",
			gast.GetAstFileName(
				v.context.ArbitrationProvider.Pkg().FSet(),
				file,
			),
		)
	}

	holder, err := v.getComments(node.Doc, v.currentGenDecl)
	if err != nil {
		return nil, graphs.SymbolKey{}, v.frozenError(err)
	}

	structType, isStruct := node.Type.(*ast.StructType)
	if !isStruct {
		return nil, graphs.SymbolKey{}, v.getFrozenError("non-struct node '%v' was provided to VisitStructType", node.Name.Name)
	}

	fields, err := v.buildFields(pkg, file, structType)
	if err != nil {
		return nil, graphs.SymbolKey{}, v.frozenError(err)
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

	symNode, err := v.context.GraphBuilder.AddStruct(
		symboldg.CreateStructNode{
			Data:        *structMeta,
			Annotations: structMeta.Annotations,
		},
	)

	return structMeta, symNode.Id, err
}

func (v *RecursiveTypeVisitor) VisitEnumType(
	pkg *packages.Package,
	file *ast.File,
	fVersion *gast.FileVersion,
	node *ast.TypeSpec,
) (metadata.EnumMeta, graphs.SymbolKey, error) {
	typeName, err := gast.GetTypeNameOrError(pkg, node.Name.Name)
	if err != nil {
		return metadata.EnumMeta{}, graphs.SymbolKey{}, err
	}

	meta, err := v.extractEnumAliasType(pkg, fVersion, node, typeName)
	if err != nil {
		return meta, graphs.SymbolKey{}, err
	}

	v.context.MetadataCache.AddEnum(&meta)

	symNode, err := v.context.GraphBuilder.AddEnum(
		symboldg.CreateEnumNode{
			Data:        meta,
			Annotations: meta.Annotations,
		},
	)

	if err != nil {
		return meta, graphs.SymbolKey{}, err
	}

	return meta, symNode.Id, nil
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

		// These results will either be full, for real types, nil for universe ones or an outright error
		resolvedField, err := gast.ResolveTypeSpecFromField(
			pkg,
			file,
			field,
			v.context.ArbitrationProvider.Pkg().GetPackage,
		)

		if err != nil {
			return nil, v.frozenError(err)
		}

		if resolvedField.DeclaringPackage != nil &&
			resolvedField.DeclaringAstFile != nil &&
			resolvedField.TypeSpec != nil {

			fVersion, err := v.context.MetadataCache.GetFileVersion(
				resolvedField.DeclaringAstFile,
				resolvedField.DeclaringPackage.Fset,
			)
			if err != nil {
				return nil, v.getFrozenError(
					"failed to obtain FileVersion for type %s in package %s - %w",
					resolvedField.TypeName,
					resolvedField.DeclaringPackage.Name,
					err,
				)
			}
			// Recurse into any nested entities
			_, err = v.visitTypeSpec(
				resolvedField.DeclaringPackage,
				resolvedField.DeclaringAstFile,
				fVersion,
				resolvedField.TypeSpec,
			)

			if err != nil {
				return nil, v.frozenError(err)
			}
		}

		// If we're looking at a universe type, there's no package and PkgPath is left empty
		typeRefPkgPath := ""
		if !resolvedField.IsUniverse {
			typeRefPkgPath = resolvedField.DeclaringPackage.PkgPath
		}

		// If we're looking at a universe type, there's no AST file nor FVersion. Special case.
		var typeRefKey graphs.SymbolKey
		if resolvedField.DeclaringAstFile != nil {
			if resolvedField.TypeSpec == nil {
				return nil, v.getFrozenError("nil TypeSpec unexpectedly returned from ResolveTypeSpecFromField")
			}

			typeFileFVersion, err := v.context.MetadataCache.GetFileVersion(
				resolvedField.DeclaringAstFile,
				resolvedField.DeclaringPackage.Fset,
			)
			typeRefKey = graphs.NewSymbolKey(resolvedField.TypeSpec, typeFileFVersion)
			if err != nil {
				return nil, v.frozenError(err)
			}
		} else {
			typeRefKey = graphs.NewUniverseSymbolKey(resolvedField.TypeName)

			if primitive, isPrimitive := symboldg.ToPrimitiveType(resolvedField.TypeName); isPrimitive {
				v.context.GraphBuilder.AddPrimitive(primitive)
			} else {
				if special, isSpecial := symboldg.ToSpecialType(resolvedField.TypeName); isSpecial {
					v.context.GraphBuilder.AddSpecial(special)
				} else {
					return nil, v.getFrozenError(
						"encountered an unexpected, non-primitive, non-error typename '%s'",
						resolvedField.TypeName,
					)
				}
			}
		}

		// Create TypeUsageMeta (AST part only)
		typeUsage := metadata.TypeUsageMeta{
			SymNodeMeta: metadata.SymNodeMeta{
				Name:        typeIdentName,
				Node:        field.Type,
				PkgPath:     typeRefPkgPath,
				SymbolKind:  common.SymKindField,
				Annotations: holder,
			},
			TypeRefKey: typeRefKey,
			Import:     importType,
		}

		fieldFVersion, err := v.context.MetadataCache.GetFileVersion(file, pkg.Fset)
		if err != nil {
			return results, v.frozenError(err)
		}

		fieldMeta := metadata.FieldMeta{
			SymNodeMeta: metadata.SymNodeMeta{
				Name:        name,
				Node:        field,
				PkgPath:     pkg.PkgPath,
				SymbolKind:  common.SymKindField,
				Annotations: holder,
				FVersion:    fieldFVersion,
			},
			Type:       typeUsage,
			IsEmbedded: isEmbedded,
		}

		results = append(results, fieldMeta)

		err = v.processTypeUsage(pkg, file, field, typeUsage)
		if err != nil {
			return results, v.frozenError(err)
		}

		_, err = v.context.GraphBuilder.AddField(symboldg.CreateFieldNode{
			Data:        fieldMeta,
			Annotations: holder,
		})

		if err != nil {
			return nil, v.frozenError(err)
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
		// Note that fields may have fields, like:
		// a, b, c int
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

	resolvedField, err := gast.ResolveTypeSpecFromField(
		pkg,
		file,
		field,
		v.context.ArbitrationProvider.Pkg().GetPackage,
	)

	if err != nil {
		return v.frozenError(err)
	}

	if resolvedField.IsUniverse {
		return nil
	}

	if resolvedField.DeclaringPackage == nil || resolvedField.DeclaringAstFile == nil || resolvedField.TypeSpec == nil {
		return v.getFrozenError("result of ResolveTypeSpecFromField for field [%v] is missing values", field.Names)
	}

	fVersion, err := v.context.MetadataCache.GetFileVersion(
		resolvedField.DeclaringAstFile,
		resolvedField.DeclaringPackage.Fset,
	)

	if err != nil {
		return v.frozenError(err)
	}

	key := graphs.NewSymbolKey(resolvedField.TypeSpec, fVersion)
	if v.context.MetadataCache.HasVisited(key) {
		return nil
	}

	// Ok, recurse
	_, err = v.visitTypeSpec(
		resolvedField.DeclaringPackage,
		resolvedField.DeclaringAstFile,
		fVersion,
		resolvedField.TypeSpec,
	)
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
) (graphs.SymbolKey, error) {
	var err error

	switch t := spec.Type.(type) {
	case *ast.StructType:
		_, symKey, err := v.VisitStructType(file, spec)
		if err != nil {
			return graphs.SymbolKey{}, v.frozenError(err)
		}

		return symKey, v.frozenIfError(err)

	case *ast.Ident:
		// Enum-like: alias of primitive (string, int, etc)
		// Optional: Check constants with the same name prefix to confirm it's "really" enum-like
		_, symKey, err := v.VisitEnumType(pkg, file, fVersion, spec)
		return symKey, v.frozenIfError(err)

	case *ast.InterfaceType:
		// Check for special interface: Context
		if spec.Name.Name == "Context" && pkg.PkgPath == "context" {
			// Mark it as a special symbol
			key := v.context.GraphBuilder.AddSpecial(symboldg.SpecialTypeContext).Id
			return key, nil
		}

		// Otherwise, reject
		err = fmt.Errorf("interface type %q not supported (only Context is allowed)", spec.Name.Name)

	default:
		err = fmt.Errorf("unhandled TypeSpec type: %T", t)
	}

	return graphs.SymbolKey{}, err
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
