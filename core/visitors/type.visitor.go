package visitors

import (
	"fmt"
	"go/ast"
	"go/types"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	"golang.org/x/tools/go/packages"
)

type RecursiveTypeVisitor struct {
	BaseVisitor

	// The file currently being worked on
	currentSourceFile *ast.File

	// The last-encountered GenDecl.
	//
	// Documentation may be placed on TypeDecl or their parent GenDecl so we track these,
	// in case we need to fetch the docs from the TypeDecl's parent.
	currentGenDecl *ast.GenDecl
}

// NewRecursiveTypeVisitor Instantiates a new type visitor.
func NewRecursiveTypeVisitor(context *VisitContext) (*RecursiveTypeVisitor, error) {
	visitor := RecursiveTypeVisitor{}
	err := visitor.initialize(context)
	return &visitor, err
}

// Visit fulfils the ast.Visitor interface
func (v *RecursiveTypeVisitor) Visit(node ast.Node) ast.Visitor {
	v.enterFmt("Visiting node interface at position %d", node.Pos())
	defer v.exit()

	switch currentNode := node.(type) {
	case *ast.File:
		// Update the current file when visiting an *ast.File node
		v.currentSourceFile = currentNode
	case *ast.GenDecl:
		v.currentGenDecl = currentNode
	case *ast.TypeSpec:
		// Check if it's a struct and if it embeds GleeceController
		if _, isStruct := currentNode.Type.(*ast.StructType); isStruct {
			_, _, err := v.VisitStructType(v.currentSourceFile, v.currentGenDecl, currentNode)
			if err != nil {
				v.setLastError(err)
				return v
			}
		}
	}
	return v
}

func (v *RecursiveTypeVisitor) VisitField(
	pkg *packages.Package,
	file *ast.File,
	field *ast.Field,
) ([]metadata.FieldMeta, error) {
	v.enterFmt("Visiting field '%v' in file %s", field.Names, file.Name.Name)
	defer v.exit()

	var results []metadata.FieldMeta

	holder, err := v.getAnnotations(field.Doc, nil)
	if err != nil {
		return nil, v.frozenError(err)
	}

	isEmbedded := len(field.Names) == 0
	names := v.resolveFieldNames(field, isEmbedded)

	for _, nameIdent := range names {
		fieldMeta, err := v.buildFieldMeta(pkg, file, field, nameIdent, isEmbedded, holder)
		if err != nil {
			return nil, err
		}

		results = append(results, fieldMeta)

		if _, err := v.resolveFieldTypeRecursive(pkg, file, field); err != nil {
			return nil, v.frozenError(err)
		}

		if _, err := v.context.GraphBuilder.AddField(symboldg.CreateFieldNode{
			Data:        fieldMeta,
			Annotations: holder,
		}); err != nil {
			return nil, v.frozenError(err)
		}
	}

	return results, nil
}

// VisitStructType dives into a Struct, given as a TypeSpec and returns its metadata.
//
// Note that this method is internally recursive and will dive into the struct's dependencies such
// as nested fields and their types and build/store their entities in the symbol graph.
func (v *RecursiveTypeVisitor) VisitStructType(
	file *ast.File,
	nodeGenDecl *ast.GenDecl,
	node *ast.TypeSpec,
) (metadata.StructMeta, graphs.SymbolKey, error) {
	v.enterFmt("Visiting struct %s", node.Name.Name)
	defer v.exit()

	// Check the cache first
	symKey, fVersion, cached, err := v.checkStructCache(file, node)
	if err != nil || cached != nil {
		return *cached, symKey, err
	}

	// If the type isn't cached, build it
	structMeta, err := v.constructStructMeta(file, fVersion, nodeGenDecl, node)
	if err != nil {
		return structMeta, symKey, err
	}

	// Insert it to the cache for faster lookup next time we encounter it
	v.context.MetadataCache.AddStruct(&structMeta)

	// And finally, add it to the symbol graph
	symNode, err := v.context.GraphBuilder.AddStruct(
		symboldg.CreateStructNode{
			Data:        structMeta,
			Annotations: structMeta.Annotations,
		},
	)

	return structMeta, symNode.Id, err
}

func (v *RecursiveTypeVisitor) VisitEnumType(
	pkg *packages.Package,
	file *ast.File,
	fVersion *gast.FileVersion,
	nodeGenDecl *ast.GenDecl,
	node *ast.TypeSpec,
) (metadata.EnumMeta, graphs.SymbolKey, error) {
	v.enterFmt("Visiting enum %s.%s", pkg.PkgPath, node.Name.Name)
	defer v.exit()

	// Check cache first
	symKey := graphs.NewSymbolKey(node, fVersion)
	cached := v.context.MetadataCache.GetEnum(symKey)
	if cached != nil {
		return *cached, symKey, nil
	}

	typeName, err := gast.GetTypeNameOrError(pkg, node.Name.Name)
	if err != nil {
		return metadata.EnumMeta{}, graphs.SymbolKey{}, err
	}

	meta, err := v.extractEnumAliasType(pkg, fVersion, nodeGenDecl, node, typeName)
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

func (v *RecursiveTypeVisitor) constructStructMeta(
	file *ast.File,
	fVersion *gast.FileVersion,
	nodeGenDecl *ast.GenDecl,
	node *ast.TypeSpec,
) (metadata.StructMeta, error) {
	v.enterFmt("Constructing struct metadata for %s", node.Name.Name)
	defer v.exit()

	pkg, err := v.context.ArbitrationProvider.Pkg().GetPackageForFile(file)
	if err != nil {
		return metadata.StructMeta{}, v.frozenError(err)
	}

	if pkg == nil {
		return metadata.StructMeta{}, v.getFrozenError(
			"could not determine package for file %s",
			gast.GetAstFileName(
				v.context.ArbitrationProvider.Pkg().FSet(),
				file,
			),
		)
	}

	holder, err := v.getAnnotations(node.Doc, nodeGenDecl)
	if err != nil {
		return metadata.StructMeta{}, v.frozenError(err)
	}

	structType, isStruct := node.Type.(*ast.StructType)
	if !isStruct {
		return metadata.StructMeta{}, v.getFrozenError("non-struct node '%v' was provided to VisitStructType", node.Name.Name)
	}

	fields, err := v.buildStructFields(pkg, file, structType)
	if err != nil {
		return metadata.StructMeta{}, v.frozenError(err)
	}

	return metadata.StructMeta{
		SymNodeMeta: metadata.SymNodeMeta{
			Name:        node.Name.Name,
			Node:        node,
			SymbolKind:  common.SymKindStruct,
			PkgPath:     pkg.PkgPath,
			Annotations: holder,
			FVersion:    fVersion,
		},
		Fields: fields,
	}, nil
}

// checkStructCache retrieves cached metadata for the given file/node combination.
// Returns the relevant SymbolKey, FileVersion and any cached information, if it exists
func (v *RecursiveTypeVisitor) checkStructCache(
	file *ast.File,
	node *ast.TypeSpec,
) (graphs.SymbolKey, *gast.FileVersion, *metadata.StructMeta, error) {
	v.enterFmt("Checking struct cache for file %s node %s", file.Name.Name, node.Name.Name)
	defer v.exit()

	fVersion, err := v.context.MetadataCache.GetFileVersion(file, v.context.ArbitrationProvider.Pkg().FSet())
	if err != nil {
		return graphs.SymbolKey{}, nil, nil, v.frozenError(err)
	}

	symKey := graphs.NewSymbolKey(node, fVersion)
	cached := v.context.MetadataCache.GetStruct(symKey)
	return symKey, fVersion, cached, nil
}

func (v *RecursiveTypeVisitor) extractEnumAliasType(
	pkg *packages.Package,
	fVersion *gast.FileVersion,
	specGenDecl *ast.GenDecl,
	spec *ast.TypeSpec,
	typeName *types.TypeName,
) (metadata.EnumMeta, error) {
	v.enterFmt("Extracting metadata for enum %s.%s", pkg.PkgPath, typeName.Name())
	defer v.exit()

	basic, isBasicType := typeName.Type().Underlying().(*types.Basic)
	if !isBasicType {
		return metadata.EnumMeta{}, v.getFrozenError("type %s is not a basic type", typeName.Name())
	}

	kind, err := metadata.NewEnumValueKind(basic.Kind())
	if err != nil {
		return metadata.EnumMeta{}, err
	}

	enumAnnotations, err := v.getAnnotations(spec.Doc, specGenDecl)
	if err != nil {
		return metadata.EnumMeta{}, err
	}

	enumValues, err := v.getEnumValueDefinitions(
		fVersion,
		pkg,
		typeName,
		basic,
	)

	if err != nil {
		return metadata.EnumMeta{}, err
	}

	enum := metadata.EnumMeta{
		SymNodeMeta: metadata.SymNodeMeta{
			Name:        typeName.Name(),
			Node:        spec,
			SymbolKind:  common.SymKindEnum,
			PkgPath:     pkg.PkgPath,
			FVersion:    fVersion,
			Annotations: enumAnnotations,
		},
		ValueKind: kind,
		Values:    enumValues,
	}

	return enum, nil
}

func (v *RecursiveTypeVisitor) getEnumValueDefinitions(
	enumFVersion *gast.FileVersion,
	enumPkg *packages.Package,
	enumTypeName *types.TypeName,
	enumBasic *types.Basic,
) ([]metadata.EnumValueDefinition, error) {
	v.enterFmt("Obtaining enum value definitions for %s.%s", enumPkg.PkgPath, enumTypeName.Name())
	defer v.exit()

	enumValues := []metadata.EnumValueDefinition{}

	scope := enumPkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		constVal, isConst := obj.(*types.Const)
		if !isConst || !types.Identical(constVal.Type(), enumTypeName.Type()) {
			continue
		}

		val := gast.ExtractConstValue(enumBasic.Kind(), constVal)
		if val == nil {
			continue
		}

		// Attempt to find the AST node
		var constNode ast.Node
		if v.context != nil && v.context.ArbitrationProvider != nil {
			constNode = gast.FindConstSpecNode(enumPkg, constVal.Name())
		}

		// Attempt to extract annotations
		var holder *annotations.AnnotationHolder
		if constNode != nil {
			comments := gast.GetCommentsFromNode(constNode, v.context.ArbitrationProvider.Pkg().FSet())
			if h, err := annotations.NewAnnotationHolder(comments, annotations.CommentSourceProperty); err == nil {
				holder = &h
			} else {
				return []metadata.EnumValueDefinition{}, err
			}
		}

		enumValues = append(enumValues, metadata.EnumValueDefinition{
			SymNodeMeta: metadata.SymNodeMeta{
				Name:        constVal.Name(),
				Node:        constNode,
				SymbolKind:  common.SymKindEnumValue,
				PkgPath:     enumPkg.PkgPath,
				Annotations: holder,
				FVersion:    enumFVersion,
			},
			Value: val,
		})
	}

	return enumValues, nil
}

func (v *RecursiveTypeVisitor) resolveFieldNames(field *ast.Field, isEmbedded bool) []*ast.Ident {
	if !isEmbedded {
		return field.Names
	}
	if ident := gast.GetIdentFromExpr(field.Type); ident != nil {
		return []*ast.Ident{{Name: ident.Name}}
	}
	return nil
}

func (v *RecursiveTypeVisitor) buildFieldMeta(
	pkg *packages.Package,
	file *ast.File,
	field *ast.Field,
	fieldNameIdent *ast.Ident,
	isEmbedded bool,
	holder *annotations.AnnotationHolder,
) (metadata.FieldMeta, error) {
	v.enterFmt("Building field metadata for file %s field %s", file.Name.Name, fieldNameIdent)
	defer v.exit()

	fieldFVersion, err := v.context.MetadataCache.GetFileVersion(file, pkg.Fset)
	if err != nil {
		return metadata.FieldMeta{}, v.getFrozenError(
			"could not obtain a FileVersion for file %v whilst processing field %v - %v",
			file.Name.Name,
			fieldNameIdent.Name,
			err,
		)
	}

	typeUsage, err := v.resolveTypeUsage(pkg, file, field.Type)
	if err != nil {
		return metadata.FieldMeta{}, v.getFrozenError(
			"could not create type usage metadata for field %v - %v",
			fieldNameIdent.Name,
			err,
		)
	}

	return metadata.FieldMeta{
		SymNodeMeta: metadata.SymNodeMeta{
			Name:        fieldNameIdent.Name,
			Node:        field,
			PkgPath:     pkg.PkgPath,
			SymbolKind:  common.SymKindField,
			Annotations: holder,
			FVersion:    fieldFVersion,
		},
		Type:       typeUsage,
		IsEmbedded: isEmbedded,
	}, nil
}

func (v *RecursiveTypeVisitor) resolveTypeUsage(
	pkg *packages.Package,
	file *ast.File,
	typeExpr ast.Expr,
) (metadata.TypeUsageMeta, error) {
	v.enterFmt("Type usage resolution for file %s expression %v", file.Name.Name, typeExpr)
	defer v.exit()

	// 1. Resolve the type
	resolvedType, err := gast.ResolveTypeSpecFromExpr(pkg, file, typeExpr, v.context.ArbitrationProvider.Pkg().GetPackage)
	if err != nil {
		return metadata.TypeUsageMeta{}, v.frozenError(err)
	}

	importType := common.ImportTypeNone
	var underlyingAnnotations *annotations.AnnotationHolder
	var pkgPath string

	// 2. Gather any auxiliary metadata we need for non-universe types
	if !resolvedType.IsUniverse {
		importType, err = v.context.ArbitrationProvider.Ast().GetImportType(file, typeExpr)
		if err != nil {
			return metadata.TypeUsageMeta{}, v.frozenError(err)
		}

		underlyingAnnotations, err = v.tryGetUnderlyingAnnotations(resolvedType)
		if err != nil {
			return metadata.TypeUsageMeta{}, err
		}

		if !resolvedType.IsUniverse {
			pkgPath = resolvedType.DeclaringPackage.PkgPath
		}
	}

	typeUsageKind := getTypeSymKind(resolvedType.DeclaringPackage, resolvedType)
	typeLayers, err := v.buildTypeLayers(
		resolvedType.DeclaringPackage,
		resolvedType.DeclaringAstFile,
		typeExpr,
		resolvedType.TypeName,
	)

	if err != nil {
		return metadata.TypeUsageMeta{}, v.getFrozenError(
			"failed to build type layers for expression with type name '%v' - %v",
			resolvedType.TypeName,
			err,
		)
	}

	// 3. Make sure that if the type is a built in like "error" or "string" or "any",
	// it's inserted into the symbol graph
	if err := v.ensureBuiltinTypeIsGraph(resolvedType); err != nil {
		return metadata.TypeUsageMeta{}, err
	}

	// Finally, cobble it all together
	return metadata.TypeUsageMeta{
		SymNodeMeta: metadata.SymNodeMeta{
			Name:        resolvedType.TypeName,
			Node:        typeExpr,
			PkgPath:     pkgPath,
			SymbolKind:  typeUsageKind,
			Annotations: underlyingAnnotations,
		},
		Layers: typeLayers,
		Import: importType,
	}, nil

}

// ensureBuiltinTypeIsGraph ensures that, if the given resolved type is a 'builtin' or 'special builtin' type, it's
// inserted to the symbol graph
func (v *RecursiveTypeVisitor) ensureBuiltinTypeIsGraph(resolved gast.TypeSpecResolution) error {
	v.enterFmt("Ensuring built-in resolution for %s", resolved.TypeName)
	defer v.exit()

	if resolved.DeclaringAstFile != nil {
		// This isn't a universe type — nothing to do here
		return nil
	}

	// Handle known universe types like string, int, etc
	if primitive, ok := symboldg.ToPrimitiveType(resolved.TypeName); ok {
		v.context.GraphBuilder.AddPrimitive(primitive)
		return nil
	}

	// Handle known special types like error
	if special, ok := symboldg.ToSpecialType(resolved.TypeName); ok {
		v.context.GraphBuilder.AddSpecial(special)
		return nil
	}

	return v.getFrozenError("encountered unknown universe type '%s'", resolved.TypeName)
}

func (v *RecursiveTypeVisitor) tryGetUnderlyingAnnotations(resolved gast.TypeSpecResolution) (*annotations.AnnotationHolder, error) {
	v.enterFmt("Retrieving annotations for resolved %s", resolved.String())
	defer v.exit()

	if resolved.IsUniverse ||
		resolved.DeclaringPackage == nil ||
		resolved.DeclaringAstFile == nil ||
		resolved.TypeSpec == nil {
		return nil, nil
	}

	return v.getAnnotations(resolved.TypeSpec.Doc, resolved.GenDecl)
}

func (v *RecursiveTypeVisitor) buildStructFields(
	pkg *packages.Package,
	file *ast.File,
	node *ast.StructType,
) ([]metadata.FieldMeta, error) {
	v.enterFmt("Building struct fields for struct at position %d in file %s", node.Struct, file.Name.Name)
	defer v.exit()

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

func (v *RecursiveTypeVisitor) resolveFieldTypeRecursive(
	pkg *packages.Package,
	file *ast.File,
	field *ast.Field,
) (gast.TypeSpecResolution, error) {
	v.enterFmt("Recursively resolving type for field [%v] in file %s", field.Names, file.Name.Name)
	defer v.exit()

	resolved, err := gast.ResolveTypeSpecFromField(pkg, file, field, v.context.ArbitrationProvider.Pkg().GetPackage)
	if err != nil {
		return gast.TypeSpecResolution{}, err
	}

	if err := v.recursivelyResolve(resolved); err != nil {
		return gast.TypeSpecResolution{}, err
	}

	return resolved, err
}

func (v *RecursiveTypeVisitor) recursivelyResolve(resolved gast.TypeSpecResolution) error {
	v.enterFmt("Headless recursive resolution for type %s", resolved.String())
	defer v.exit()

	if resolved.IsUniverse {
		return nil
	}

	if resolved.DeclaringPackage == nil || resolved.DeclaringAstFile == nil || resolved.TypeSpec == nil {
		return v.getFrozenError("resolved type '%s' is missing necessary declaration context", resolved.TypeName)
	}

	fVersion, err := v.context.MetadataCache.GetFileVersion(
		resolved.DeclaringAstFile,
		resolved.DeclaringPackage.Fset,
	)
	if err != nil {
		return v.frozenError(err)
	}

	key := graphs.NewSymbolKey(resolved.TypeSpec, fVersion)
	if v.context.MetadataCache.HasVisited(key) {
		return nil
	}

	_, err = v.visitTypeSpec(
		resolved.DeclaringPackage,
		resolved.DeclaringAstFile,
		fVersion,
		resolved.GenDecl,
		resolved.TypeSpec,
	)

	return v.frozenError(err)
}

func (v *RecursiveTypeVisitor) visitTypeSpec(
	pkg *packages.Package,
	file *ast.File,
	fVersion *gast.FileVersion,
	specGenDecl *ast.GenDecl,
	spec *ast.TypeSpec,
) (graphs.SymbolKey, error) {
	v.enterFmt("Visiting type spec %s.%s", pkg.PkgPath, spec.Name.Name)
	defer v.exit()

	var err error
	switch t := spec.Type.(type) {
	case *ast.StructType:
		_, symKey, err := v.VisitStructType(file, specGenDecl, spec)
		if err != nil {
			return graphs.SymbolKey{}, v.frozenError(err)
		}

		return symKey, v.frozenIfError(err)

	case *ast.Ident:
		// Enum-like: alias of primitive (string, int, etc)
		// Optional: Check constants with the same name prefix to confirm it's "really" enum-like
		_, symKey, err := v.VisitEnumType(pkg, file, fVersion, specGenDecl, spec)
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

func (v *RecursiveTypeVisitor) getAnnotations(
	onNodeDoc *ast.CommentGroup,
	nodeGenDecl *ast.GenDecl,
) (*annotations.AnnotationHolder, error) {
	v.enter("Obtaining annotations for comment group")
	defer v.exit()

	var commentSource *ast.CommentGroup
	if onNodeDoc != nil {
		commentSource = onNodeDoc
	} else {
		if nodeGenDecl != nil {
			commentSource = nodeGenDecl.Doc
		}
	}

	if commentSource != nil {
		comments := gast.MapDocListToCommentBlock(commentSource.List, v.context.ArbitrationProvider.Pkg().FSet())
		holder, err := annotations.NewAnnotationHolder(comments, annotations.CommentSourceController)
		return &holder, err
	}

	return nil, nil
}

func (v *RecursiveTypeVisitor) buildTypeLayers(
	pkg *packages.Package,
	file *ast.File,
	expr ast.Expr,
	exprTypeName string,
) ([]metadata.TypeLayer, error) {
	v.enterFmt("Building type layers for an expression named '%s'", exprTypeName)
	defer v.exit()

	switch t := expr.(type) {
	case *ast.StarExpr:
		inner, err := v.buildTypeLayers(pkg, file, t.X, exprTypeName)
		if err != nil {
			return nil, err
		}
		return append([]metadata.TypeLayer{metadata.NewPointerLayer()}, inner...), nil

	case *ast.ArrayType:
		inner, err := v.buildTypeLayers(pkg, file, t.Elt, exprTypeName)
		if err != nil {
			return nil, err
		}
		return append([]metadata.TypeLayer{metadata.NewArrayLayer()}, inner...), nil

	case *ast.MapType:
		keyLayers, err := v.buildTypeLayers(pkg, file, t.Key, exprTypeName)
		if err != nil {
			return nil, err
		}
		valueLayers, err := v.buildTypeLayers(pkg, file, t.Value, exprTypeName)
		if err != nil {
			return nil, err
		}

		if len(keyLayers) != 1 || keyLayers[0].Kind != metadata.TypeLayerKindBase {
			return nil, fmt.Errorf("map key must be a base type")
		}
		if len(valueLayers) != 1 || valueLayers[0].Kind != metadata.TypeLayerKindBase {
			return nil, fmt.Errorf("map value must be a base type")
		}

		return []metadata.TypeLayer{metadata.NewMapLayer(keyLayers[0].BaseTypeRef, valueLayers[0].BaseTypeRef)}, nil

	case *ast.Ident, *ast.SelectorExpr:
		if pkg == nil || file == nil {
			// E.g., this is a 'universe' type like 'string'
			typeKey := graphs.NewUniverseSymbolKey(exprTypeName)
			return []metadata.TypeLayer{metadata.NewBaseLayer(&typeKey)}, nil
		}
		baseKey, err := v.resolveBaseTypeKey(pkg, file, expr)
		if err != nil {
			return nil, err
		}
		return []metadata.TypeLayer{metadata.NewBaseLayer(baseKey)}, nil

	default:
		return nil, fmt.Errorf("unsupported type expression: %T", t)
	}
}

func (v *RecursiveTypeVisitor) resolveBaseTypeKey(
	pkg *packages.Package,
	file *ast.File,
	expr ast.Expr,
) (*graphs.SymbolKey, error) {
	v.enterFmt("Resolving base type key for expression position %d in file %s", expr.Pos(), file.Name.Name)
	defer v.exit()

	resolved, err := gast.ResolveTypeSpecFromExpr(pkg, file, expr, v.context.ArbitrationProvider.Pkg().GetPackage)
	if err != nil {
		return nil, err
	}

	if resolved.DeclaringAstFile != nil && resolved.TypeSpec != nil {
		fVersion, err := v.context.MetadataCache.GetFileVersion(resolved.DeclaringAstFile, resolved.DeclaringPackage.Fset)
		if err != nil {
			return nil, err
		}
		key := graphs.NewSymbolKey(resolved.TypeSpec, fVersion)
		return &key, nil
	}

	key := graphs.NewUniverseSymbolKey(resolved.TypeName)
	return &key, nil
}

func getTypeSymKind(
	pkg *packages.Package,
	resolvedType gast.TypeSpecResolution,
) common.SymKind {
	if _, isSpecial := symboldg.ToSpecialType(resolvedType.TypeName); isSpecial {
		return common.SymKindSpecialBuiltin
	}

	if resolvedType.IsUniverse {
		return common.SymKindBuiltin
	}

	if resolvedType.TypeSpec == nil {
		return common.SymKindUnknown
	}

	switch resolvedType.TypeSpec.Type.(type) {
	case *ast.StructType:
		return common.SymKindStruct

	case *ast.InterfaceType:
		if resolvedType.TypeSpec.Name.Name == "Context" && pkg.PkgPath == "context" {
			return common.SymKindSpecialBuiltin
		}
		return common.SymKindInterface

	case *ast.Ident:
		// Use your enum detection here — e.g., check constants or alias metadata
		if gast.IsEnumLike(pkg, resolvedType.TypeSpec) {
			return common.SymKindEnum
		}
		return common.SymKindAlias

	default:
		return common.SymKindUnknown
	}
}
