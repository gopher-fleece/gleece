package visitors

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/packages"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/metadata/typeref"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
)

type topLevelMeta struct {
	SymName     string
	PackagePath string
	NodeRange   common.ResolvedRange
	Annotations *annotations.AnnotationHolder
	FileVersion *gast.FileVersion
	SymbolKind  common.SymKind
}

// TypeUsageVisitor builds TypeRef trees for usage sites.
// VisitExpr accepts a typeParamEnv (map[name]index) or nil.
// It may call a TypeDeclVisitor on-demand to materialize declarations.
type TypeUsageVisitor struct {
	BaseVisitor
	declVisitor *TypeDeclVisitor // optional; set via setDeclVisitor after wiring

	materializing bool
}

// NewTypeUsageVisitor constructs a new TypeUsageVisitor.
func NewTypeUsageVisitor(context *VisitContext, materializing bool) (*TypeUsageVisitor, error) {
	v := &TypeUsageVisitor{materializing: materializing}
	if err := v.initialize(context); err != nil {
		return nil, err
	}
	return v, nil
}

func (v *TypeUsageVisitor) setDeclVisitor(visitor *TypeDeclVisitor) {
	v.declVisitor = visitor
}

// VisitExpr is the canonical entrypoint for usage-side type analysis.
// It remains small: helpers do the work.
func (v *TypeUsageVisitor) VisitExpr(
	pkg *packages.Package,
	file *ast.File,
	expr ast.Expr,
	typeParamEnv map[string]int,
) (metadata.TypeUsageMeta, error) {
	fileName := gast.GetAstFileNameOrFallback(file, nil)
	v.enterFmt("TypeUsageVisitor.VisitExpr file=%s expr=%T", fileName, expr)
	defer v.exit()

	if expr == nil {
		return metadata.TypeUsageMeta{}, nil
	}

	// 1) structural: always build the TypeRef tree (may include NamedTypeRef with TypeArgs)
	root, err := v.buildTypeRef(pkg, file, expr, typeParamEnv)
	if err != nil {
		return metadata.TypeUsageMeta{}, err
	}

	// 2) derive top-level meta (name, pkg, annotations, FileVersion, SymKind) and import type
	topMeta, importType, err := v.deriveTopMetaFromRoot(pkg, file, expr, root)
	if err != nil {
		return metadata.TypeUsageMeta{}, err
	}

	// 3) build top-level type-arguments metadata (if any)
	typeArgs, err := v.buildTypeParamsFromExpr(pkg, file, expr, typeParamEnv)
	if err != nil {
		return metadata.TypeUsageMeta{}, err
	}

	// 4) assemble final TypeUsageMeta
	res := metadata.TypeUsageMeta{
		SymNodeMeta: metadata.SymNodeMeta{
			Name:        topMeta.SymName,
			Node:        expr,
			SymbolKind:  topMeta.SymbolKind,
			PkgPath:     topMeta.PackagePath,
			Annotations: topMeta.Annotations,
			Range:       topMeta.NodeRange,
			FVersion:    topMeta.FileVersion,
		},
		TypeParams: typeArgs,
		Import:     importType,
		Root:       root,
	}
	return res, nil
}

// deriveTopMetaFromRoot inspects the structural root and returns a topLevelMeta and import type.
// It collapses pointer/array/slice wrappers first so pointer-to-enum and slice-of-instantiated behave correctly.
func (v *TypeUsageVisitor) deriveTopMetaFromRoot(
	pkg *packages.Package,
	file *ast.File,
	expr ast.Expr,
	root metadata.TypeRef,
) (*topLevelMeta, common.ImportType, error) {
	// Collapse pointer/array/slice to the inner element when possible
	innerRef, collapsed := collapseContainerTypeRef(root)

	// Unwrap common AST wrappers so resolution sees the inner node
	unwrappedExpr := unwrapExpr(expr)

	// Check named-ness against the inner (collapsed) ref
	named, isNamed := isNamedRef(innerRef)
	if isNamed && named != nil {
		// plain named reference without type args -> resolve declaration or universe
		if len(named.TypeArgs) == 0 {
			return v.resolveNamedType(pkg, file, unwrappedExpr)
		}

		// Instantiated named usage T[A,...] -> pass the named inner ref for generic handling
		return v.resolveGenericNamedType(pkg, file, unwrappedExpr, innerRef)
	}

	// Not a named inner type:
	// - if we collapsed a container, prefer resolving the composite/inner using the unwrapped expr
	// - otherwise resolve using the original shape (preserve pointer/composite for composite resolver)
	if collapsed {
		return v.resolveCompositeType(pkg, file, unwrappedExpr, innerRef)
	}
	return v.resolveCompositeType(pkg, file, expr, root)
}

func (v *TypeUsageVisitor) resolveNamedType(
	pkg *packages.Package,
	file *ast.File,
	expr ast.Expr,
) (*topLevelMeta, common.ImportType, error) {
	var importType common.ImportType

	resolution, ok, resolveErr := gast.ResolveNamedType(
		pkg,
		file,
		expr,
		v.context.ArbitrationProvider.Pkg().GetPackage,
	)
	if resolveErr != nil {
		return nil, importType, v.getFrozenError(
			"ResolveNamedType error for %T: %v", expr, resolveErr,
		)
	}
	if !ok {
		return nil, importType, v.getFrozenError(
			"could not resolve named type for expression (%T)", expr,
		)
	}

	// Universe/builtin types
	if resolution.IsUniverse {
		if err := v.ensureUniverseTypeInGraph(resolution.TypeName); err != nil {
			return nil, importType, err
		}

		usageMeta, err := v.createUsageMeta(
			pkg,
			file,
			resolution.TypeName,
			common.SymKindBuiltin,
			expr,
		)

		return usageMeta, importType, err
	}

	// Declared type -> full derivation and import computation
	derived, err := v.deriveTopLevelMeta(resolution, pkg, expr)
	if err != nil {
		return nil, importType, err
	}
	it, err := v.context.ArbitrationProvider.Ast().GetImportType(file, expr)
	if err != nil {
		return nil, importType, err
	}
	importType = it

	// ensure materialized
	if err := v.ensureDeclMaterialized(resolution, derived.FileVersion); err != nil {
		return nil, importType, err
	}

	return derived, importType, nil
}

func (v *TypeUsageVisitor) resolveGenericNamedType(
	pkg *packages.Package,
	file *ast.File,
	expr ast.Expr,
	root metadata.TypeRef,
) (*topLevelMeta, common.ImportType, error) {
	resolution, ok, resolveErr := gast.ResolveNamedType(
		pkg,
		file,
		expr,
		v.context.ArbitrationProvider.Pkg().GetPackage,
	)

	if resolveErr != nil {
		return nil, common.ImportTypeNone, v.getFrozenError(
			"ResolveNamedType error for instantiated base %T: %v", expr, resolveErr,
		)
	}

	// If resolver says it could not find a named base -> strict error.
	if !ok {
		return nil, common.ImportTypeNone, v.getFrozenError(
			"could not resolve base identifier for instantiated type '%s' (%T)",
			canonicalNameForUsage(root),
			expr,
		)
	}

	// If base resolved to universe -> ensure primitive and return alias-kind meta
	if resolution.IsUniverse {
		if err := v.ensureUniverseTypeInGraph(resolution.TypeName); err != nil {
			return nil, common.ImportTypeNone, err
		}

		// usage-centric alias is acceptable here (base is builtin)
		usageMeta, err := v.createUsageMeta(
			pkg,
			file,
			resolution.TypeName,
			common.SymKindBuiltin,
			expr,
		)

		return usageMeta, common.ImportTypeNone, err
	}

	return v.resolveNonUniverseGenericNamed(pkg, file, expr, root, resolution)
}

func (v *TypeUsageVisitor) resolveNonUniverseGenericNamed(
	pkg *packages.Package,
	file *ast.File,
	expr ast.Expr,
	root metadata.TypeRef,
	resolution gast.TypeSpecResolution,
) (*topLevelMeta, common.ImportType, error) {
	usageName := canonicalNameForUsage(root)

	if resolution.DeclaringPackage == nil || resolution.DeclaringAstFile == nil || resolution.TypeSpec == nil {
		return nil, common.ImportTypeNone, v.getFrozenError(
			"incomplete resolution for declared instantiated base %s (canonical name %s)",
			resolution.TypeName,
			usageName,
		)
	}

	// Get the declaring file version (strict)
	fv, fvErr := v.context.MetadataCache.GetFileVersion(
		resolution.DeclaringAstFile,
		resolution.DeclaringPackage.Fset,
	)

	if fvErr != nil {
		return nil, common.ImportTypeNone, fvErr
	}

	if fv == nil {
		return nil, common.ImportTypeNone, v.getFrozenError(
			"file version missing for declaring file of instantiated base %s",
			usageName,
		)
	}

	// Ensure the declared type has been materialized (decl visitor will parse and insert the decl).
	if err := v.ensureDeclMaterialized(resolution, fv); err != nil {
		return nil, common.ImportTypeNone, err
	}

	// Build usage-centric top-level meta for the instantiated usage.
	top := &topLevelMeta{
		SymName:     usageName,
		PackagePath: resolution.DeclaringPackage.PkgPath,     // origin package of the base
		NodeRange:   common.ResolveNodeRange(pkg.Fset, expr), // usage range (not decl range)
		Annotations: nil,                                     // keep annotations exclusive to decl visitor
		FileVersion: fv,                                      // declaring file version (useful for canonicalization)
		SymbolKind:  chooseSymKind(resolution, pkg),          // conservative choice based on declaration
	}

	// Resolve whether the usage imports from another package (import type)
	it, itErr := v.context.ArbitrationProvider.Ast().GetImportType(file, expr)
	if itErr != nil {
		return nil, common.ImportTypeNone, itErr
	}

	return top, it, nil
}

func (v *TypeUsageVisitor) resolveCompositeType(
	pkg *packages.Package,
	file *ast.File,
	expr ast.Expr,
	root metadata.TypeRef,
) (*topLevelMeta, common.ImportType, error) {
	var symKind common.SymKind
	if _, isParam := root.(*typeref.ParamTypeRef); isParam {
		symKind = common.SymKindUnknown
	} else {
		symKind = common.SymKindAlias
	}

	usageMeta, err := v.createUsageMeta(
		pkg,
		file,
		root.CanonicalString(),
		symKind,
		expr,
	)

	return usageMeta, common.ImportTypeNone, err
}

// deriveTopLevelMeta returns name/pkg/annotations/FileVersion/SymKind for a declared resolution.
func (v *TypeUsageVisitor) deriveTopLevelMeta(
	resolution gast.TypeSpecResolution,
	currentPkg *packages.Package,
	expr ast.Expr,
) (*topLevelMeta, error) {

	// Universe handled by caller; here we expect declared resolution.
	if resolution.IsUniverse {
		return nil, v.getFrozenError("deriveTopLevelMeta called for universe type %s", resolution.TypeName)
	}

	if resolution.DeclaringPackage == nil ||
		resolution.DeclaringAstFile == nil ||
		resolution.TypeSpec == nil {
		return nil, v.getFrozenError("incomplete resolution for declared type %s", resolution.TypeName)
	}

	fileVersion, err := v.context.MetadataCache.GetFileVersion(
		resolution.DeclaringAstFile,
		resolution.DeclaringPackage.Fset,
	)
	if err != nil {
		return nil, err
	}
	if fileVersion == nil {
		return nil, v.getFrozenError("file version missing for declaring file of type %s", resolution.TypeName)
	}

	holder, aErr := v.getAnnotations(resolution.TypeSpec.Doc, resolution.GenDecl)
	if aErr != nil {
		return nil, aErr
	}

	return &topLevelMeta{
		SymName:     resolution.TypeName,
		PackagePath: resolution.DeclaringPackage.PkgPath,
		NodeRange:   common.ResolveNodeRange(resolution.DeclaringPackage.Fset, expr),
		Annotations: holder,
		FileVersion: fileVersion,
		SymbolKind:  chooseSymKind(resolution, currentPkg),
	}, nil
}

// ensureDeclMaterialized makes sure declared type has been materialized (cached + graph).
func (v *TypeUsageVisitor) ensureDeclMaterialized(
	resolution gast.TypeSpecResolution,
	fileVersion *gast.FileVersion,
) error {
	// Check if the visitor is meant to materialize nodes - this may be false for type-constraint analysis
	if !v.materializing {
		return nil
	}

	// No-op for universe types
	if resolution.IsUniverse {
		return nil
	}
	if resolution.TypeSpec == nil || resolution.DeclaringAstFile == nil || resolution.DeclaringPackage == nil {
		return v.getFrozenError("cannot materialize incomplete declaration for %s", resolution.TypeName)
	}
	if fileVersion == nil {
		return v.getFrozenError("missing fileVersion when trying to materialize %s", resolution.TypeName)
	}

	key := graphs.NewSymbolKey(resolution.TypeSpec, fileVersion)
	if v.context.MetadataCache.HasVisited(key) {
		return nil
	}

	_, err := v.declVisitor.EnsureDeclMaterialized(
		resolution.DeclaringPackage,
		resolution.DeclaringAstFile,
		resolution.GenDecl,
		resolution.TypeSpec,
	)
	return err
}

// ensureUniverseTypeInGraph inserts primitives/specials into the graph so later edges can reference them.
func (v *TypeUsageVisitor) ensureUniverseTypeInGraph(typeName string) error {
	if !v.materializing {
		return nil
	}

	v.enterFmt("ensureUniverseTypeInGraph %s", typeName)
	defer v.exit()

	// primitives
	if prim, ok := symboldg.ToPrimitiveType(typeName); ok {
		if !v.context.GraphBuilder.IsPrimitivePresent(prim) {
			v.context.GraphBuilder.AddPrimitive(prim)
		}
		return nil
	}

	// special types
	if sp, ok := symboldg.ToSpecialType(typeName); ok {
		if !v.context.GraphBuilder.IsSpecialPresent(sp) {
			v.context.GraphBuilder.AddSpecial(sp)
		}
		return nil
	}

	return v.getFrozenError("unknown universe type '%s'", typeName)
}

// ------------------------- buildTypeRef dispatcher + small helpers -------------------------

func (v *TypeUsageVisitor) buildTypeRef(
	pkg *packages.Package,
	file *ast.File,
	expr ast.Expr,
	typeParamEnv map[string]int,
) (metadata.TypeRef, error) {
	switch t := expr.(type) {
	case *ast.Ident, *ast.SelectorExpr:
		return v.buildIdentOrSelectorRef(pkg, file, expr, typeParamEnv)
	case *ast.StarExpr:
		return v.buildPtrRef(pkg, file, t, typeParamEnv)
	case *ast.ArrayType:
		return v.buildArrayOrSliceRef(pkg, file, t, typeParamEnv)
	case *ast.MapType:
		return v.buildMapRef(pkg, file, t, typeParamEnv)
	case *ast.FuncType:
		return v.buildFuncTypeRef(pkg, file, t, typeParamEnv)
	case *ast.IndexExpr:
		return v.buildIndexExprRef(pkg, file, t, typeParamEnv)
	case *ast.IndexListExpr:
		return v.buildIndexListExprRef(pkg, file, t, typeParamEnv)
	case *ast.StructType:
		return v.buildInlineStructRef(pkg, file, t, typeParamEnv)
	default:
		return nil, fmt.Errorf("unsupported type expression: %T", expr)
	}
}

func (v *TypeUsageVisitor) buildPtrRef(
	pkg *packages.Package,
	file *ast.File,
	ptr *ast.StarExpr,
	typeParamEnv map[string]int,
) (metadata.TypeRef, error) {
	elem, err := v.buildTypeRef(pkg, file, ptr.X, typeParamEnv)
	if err != nil {
		return nil, err
	}
	return &typeref.PtrTypeRef{Elem: elem}, nil
}

func (v *TypeUsageVisitor) buildArrayOrSliceRef(
	pkg *packages.Package,
	file *ast.File,
	array *ast.ArrayType,
	typeParamEnv map[string]int,
) (metadata.TypeRef, error) {
	elem, err := v.buildTypeRef(pkg, file, array.Elt, typeParamEnv)
	if err != nil {
		return nil, err
	}
	if array.Len == nil {
		return &typeref.SliceTypeRef{Elem: elem}, nil
	}
	// preserving Len as nil for now
	return &typeref.ArrayTypeRef{Len: nil, Elem: elem}, nil
}

func (v *TypeUsageVisitor) buildMapRef(
	pkg *packages.Package,
	file *ast.File,
	m *ast.MapType,
	typeParamEnv map[string]int,
) (metadata.TypeRef, error) {
	keyRef, err := v.buildTypeRef(pkg, file, m.Key, typeParamEnv)
	if err != nil {
		return nil, err
	}
	valueRef, err := v.buildTypeRef(pkg, file, m.Value, typeParamEnv)
	if err != nil {
		return nil, err
	}
	return &typeref.MapTypeRef{Key: keyRef, Value: valueRef}, nil
}

func (v *TypeUsageVisitor) buildFuncTypeRef(
	pkg *packages.Package,
	file *ast.File,
	ft *ast.FuncType,
	typeParamEnv map[string]int,
) (metadata.TypeRef, error) {
	params := make([]metadata.TypeRef, 0, 4)
	results := make([]metadata.TypeRef, 0, 2)

	if ft.Params != nil {
		for _, f := range ft.Params.List {
			tr, err := v.buildTypeRef(pkg, file, f.Type, typeParamEnv)
			if err != nil {
				return nil, err
			}
			params = append(params, tr)
		}
	}
	if ft.Results != nil {
		for _, f := range ft.Results.List {
			tr, err := v.buildTypeRef(pkg, file, f.Type, typeParamEnv)
			if err != nil {
				return nil, err
			}
			results = append(results, tr)
		}
	}
	return &typeref.FuncTypeRef{Params: params, Results: results, Variadic: false}, nil
}

func (v *TypeUsageVisitor) buildIndexExprRef(
	pkg *packages.Package,
	file *ast.File,
	idx *ast.IndexExpr,
	typeParamEnv map[string]int,
) (metadata.TypeRef, error) {
	baseRef, err := v.buildTypeRef(pkg, file, idx.X, typeParamEnv)
	if err != nil {
		return nil, err
	}
	argRef, err := v.buildTypeRef(pkg, file, idx.Index, typeParamEnv)
	if err != nil {
		return nil, err
	}
	if named, ok := baseRef.(*typeref.NamedTypeRef); ok {
		named.TypeArgs = append(named.TypeArgs, argRef)
		return named, nil
	}
	return common.Ptr(typeref.NewNamedTypeRef(nil, []metadata.TypeRef{argRef})), nil
}

func (v *TypeUsageVisitor) buildIndexListExprRef(
	pkg *packages.Package,
	file *ast.File,
	ile *ast.IndexListExpr,
	typeParamEnv map[string]int,
) (metadata.TypeRef, error) {
	baseRef, err := v.buildTypeRef(pkg, file, ile.X, typeParamEnv)
	if err != nil {
		return nil, err
	}
	args := make([]metadata.TypeRef, 0, len(ile.Indices))
	for _, idx := range ile.Indices {
		a, err := v.buildTypeRef(pkg, file, idx, typeParamEnv)
		if err != nil {
			return nil, err
		}
		args = append(args, a)
	}
	if named, ok := baseRef.(*typeref.NamedTypeRef); ok {
		named.TypeArgs = append(named.TypeArgs, args...)
		return named, nil
	}
	return common.Ptr(typeref.NewNamedTypeRef(nil, args)), nil
}

// buildInlineStructRef builds an InlineStructTypeRef for inline struct literals.
// It returns a non-graphing inline representation keyed by a rep SymbolKey.
func (v *TypeUsageVisitor) buildInlineStructRef(
	pkg *packages.Package,
	file *ast.File,
	structType *ast.StructType,
	typeParamEnv map[string]int,
) (*typeref.InlineStructTypeRef, error) {
	fv, err := v.context.MetadataCache.GetFileVersion(file, pkg.Fset)
	if err != nil {
		return nil, err
	}
	if fv == nil {
		return nil, fmt.Errorf("file version missing for inline struct in %s", file.Name.Name)
	}

	out := &typeref.InlineStructTypeRef{
		Fields: []metadata.FieldMeta{},
		RepKey: graphs.NewSymbolKey(structType, fv),
	}

	for _, field := range structType.Fields.List {
		typeUsage, err := v.VisitExpr(pkg, file, field.Type, typeParamEnv)
		if err != nil {
			return nil, err
		}
		names := gast.GetFieldNames(field)
		if len(names) == 0 {
			names = []string{""} // anonymous/embedded
		}
		isEmbedded := gast.IsEmbeddedOrAnonymousField(field)
		ann, _ := v.getAnnotations(field.Doc, nil) // inline structs have no genDecl
		for _, nm := range names {
			fMeta := metadata.FieldMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:        nm,
					Node:        field,
					SymbolKind:  common.SymKindField,
					PkgPath:     pkg.PkgPath,
					Annotations: ann,
					FVersion:    fv,
					Range:       common.ResolveNodeRange(pkg.Fset, field),
				},
				Type:       typeUsage,
				IsEmbedded: isEmbedded,
			}
			out.Fields = append(out.Fields, fMeta)
		}
	}
	return out, nil
}

// buildIdentOrSelectorRef resolves ident/selector to a TypeRef; handles type params and declared/universe types.
func (v *TypeUsageVisitor) buildIdentOrSelectorRef(
	pkg *packages.Package,
	file *ast.File,
	expr ast.Expr,
	typeParamEnv map[string]int,
) (metadata.TypeRef, error) {
	// type parameter
	if typeParamEnv != nil {
		if name := gast.GetIdentNameFromExpr(expr); name != nil {
			if idx, ok := typeParamEnv[*name]; ok {
				return &typeref.ParamTypeRef{Name: *name, Index: idx}, nil
			}
		}
	}

	resolution, ok, resolveErr := gast.ResolveNamedType(
		pkg,
		file,
		expr,
		v.context.ArbitrationProvider.Pkg().GetPackage,
	)
	if resolveErr != nil {
		return nil, resolveErr
	}
	if !ok {
		return nil, fmt.Errorf("could not resolve named type for expression (%T)", expr)
	}

	// universe -> ensure primitive/special graph presence
	if resolution.IsUniverse {
		if err := v.ensureUniverseTypeInGraph(resolution.TypeName); err != nil {
			return nil, err
		}

		key := graphs.NewUniverseSymbolKey(resolution.TypeName)
		return common.Ptr(typeref.NewNamedTypeRef(&key, nil)), nil
	}

	// declared -> must have contextual information
	if resolution.TypeSpec == nil || resolution.DeclaringPackage == nil || resolution.DeclaringAstFile == nil {
		return nil, fmt.Errorf("incomplete declaration context for type %s", resolution.TypeName)
	}

	// Handles specials like 'time.Time' or 'context.Context'
	specialKey := v.tryGetSymKeyForSpecial(
		resolution.DeclaringPackage,
		resolution.TypeSpec,
		resolution.TypeName,
	)

	if specialKey != nil {
		return common.Ptr(typeref.NewNamedTypeRef(specialKey, nil)), nil
	}

	fileVersion, err := v.context.MetadataCache.GetFileVersion(
		resolution.DeclaringAstFile,
		resolution.DeclaringPackage.Fset,
	)
	if err != nil {
		return nil, err
	}
	if fileVersion == nil {
		return nil, fmt.Errorf("missing file version for declaring file of type %s", resolution.TypeName)
	}

	key := graphs.NewSymbolKey(resolution.TypeSpec, fileVersion)
	// on-demand materialize
	if !v.context.MetadataCache.HasVisited(key) {
		if v.declVisitor == nil {
			return nil, fmt.Errorf("declaration for %s not materialized and no materializer provided", resolution.TypeName)
		}

		_, err := v.declVisitor.EnsureDeclMaterialized(
			resolution.DeclaringPackage,
			resolution.DeclaringAstFile,
			resolution.GenDecl,
			resolution.TypeSpec,
		)

		if err != nil {
			return nil, err
		}
	}

	return common.Ptr(typeref.NewNamedTypeRef(&key, nil)), nil
}

// ------------------------- type param helpers -------------------------

// buildTypeParamsFromExpr extracts top-level type-argument expressions (if this usage is a type instantiation)
// and returns their metadata via recursive VisitExpr calls.
func (v *TypeUsageVisitor) buildTypeParamsFromExpr(
	pkg *packages.Package,
	file *ast.File,
	expr ast.Expr,
	typeParamEnv map[string]int,
) ([]metadata.TypeUsageMeta, error) {

	// Walk down to find IndexExpr/IndexListExpr applied to a base identifier/selector.
	for {
		switch t := expr.(type) {
		case *ast.IndexExpr:
			return v.evalTypeParamExprs(pkg, file, []ast.Expr{t.Index}, typeParamEnv)
		case *ast.IndexListExpr:
			return v.evalTypeParamExprs(pkg, file, t.Indices, typeParamEnv)
		case *ast.StarExpr:
			expr = t.X
		case *ast.ArrayType:
			expr = t.Elt
		case *ast.ParenExpr:
			expr = t.X
		default:
			return nil, nil
		}
	}
}

func (v *TypeUsageVisitor) evalTypeParamExprs(
	pkg *packages.Package,
	file *ast.File,
	indexExprs []ast.Expr,
	typeParamEnv map[string]int,
) ([]metadata.TypeUsageMeta, error) {
	out := make([]metadata.TypeUsageMeta, 0, len(indexExprs))
	for _, ie := range indexExprs {
		tu, err := v.VisitExpr(pkg, file, ie, typeParamEnv)
		if err != nil {
			return nil, err
		}
		out = append(out, tu)
	}
	return out, nil
}

// createUsageMeta creates top level metadata for a usage site.
//
// Example: in
//
//	type SomeStruct struct { SomeField string }
//
// the usage metadata refers to 'string' — it must include the usage's package path
// (if known) and the usage node range so callers can report/anchor diagnostics.
func (v *TypeUsageVisitor) createUsageMeta(
	pkg *packages.Package,
	file *ast.File,
	name string,
	kind common.SymKind,
	expr ast.Expr,
) (*topLevelMeta, error) {
	var pkgPath string
	if pkg != nil {
		pkgPath = pkg.PkgPath
	}

	var nodeRange common.ResolvedRange
	// Resolve range using package FileSet when available. If pkg is nil, keep zero-value.
	if pkg != nil && expr != nil {
		nodeRange = common.ResolveNodeRange(pkg.Fset, expr)
	}

	fileVersion, err := v.context.MetadataCache.GetFileVersion(file, pkg.Fset)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to obtain FileVersion for type usage '%s' in file '%s' - %v",
			name,
			gast.GetAstFileNameOrFallback(file, nil),
			err,
		)
	}

	return &topLevelMeta{
		SymName:     name,
		PackagePath: pkgPath,
		NodeRange:   nodeRange,
		Annotations: nil,         // A type usage cannot have annotations - the field & declaration - yes. The usage? Nope.
		FileVersion: fileVersion, // The FileVersion is the same as the Field's - *not* the declaration's
		SymbolKind:  kind,
	}, nil
}

// ------------------------- utilities -------------------------

// chooseSymKind picks a SymKind based on the resolution (struct/interface/alias/enum/special/builtin).
func chooseSymKind(resolution gast.TypeSpecResolution, pkg *packages.Package) common.SymKind {
	if resolution.IsUniverse {
		return common.SymKindBuiltin
	}
	if resolution.TypeSpec == nil {
		return common.SymKindUnknown
	}
	switch t := resolution.TypeSpec.Type.(type) {
	case *ast.StructType:
		_ = t
		return common.SymKindStruct
	case *ast.InterfaceType:
		if resolution.IsContextOrTime() {
			return common.SymKindSpecialBuiltin
		}
		return common.SymKindInterface
	case *ast.Ident:
		if pkg != nil && gast.IsEnumLike(resolution.DeclaringPackage, resolution.TypeSpec) {
			return common.SymKindEnum
		}
		return common.SymKindAlias
	default:
		return common.SymKindUnknown
	}
}

// isNamedRef returns the NamedTypeRef when the root is a named reference.
func isNamedRef(root metadata.TypeRef) (*typeref.NamedTypeRef, bool) {
	if root == nil {
		return nil, false
	}
	n, ok := root.(*typeref.NamedTypeRef)
	return n, ok
}

// canonicalNameForUsage returns a deterministic usage-centric name for a TypeRef.
func canonicalNameForUsage(root metadata.TypeRef) string {
	if root == nil {
		return ""
	}
	return root.CanonicalString()
}

// collapseContainerTypeRef unwraps Ptr / Slice / Array TypeRefs and returns the innermost element.
// Returns the inner TypeRef and true if any container was collapsed.
func collapseContainerTypeRef(root metadata.TypeRef) (metadata.TypeRef, bool) {
	if root == nil {
		return nil, false
	}

	inner := root
	collapsed := false
	for {
		switch t := inner.(type) {
		case *typeref.PtrTypeRef:
			if t.Elem == nil {
				return inner, collapsed
			}
			inner = t.Elem
			collapsed = true
			continue
		case *typeref.SliceTypeRef:
			if t.Elem == nil {
				return inner, collapsed
			}
			inner = t.Elem
			collapsed = true
			continue
		case *typeref.ArrayTypeRef:
			if t.Elem == nil {
				return inner, collapsed
			}
			inner = t.Elem
			collapsed = true
			continue
		default:
			return inner, collapsed
		}
	}
}

// unwrapExpr removes common wrappers from an AST expr so resolution sees the inner expression.
// It strips &, * (unary/star) parentheses and array/slice types.
func unwrapExpr(expr ast.Expr) ast.Expr {
	for {
		switch e := expr.(type) {
		case *ast.UnaryExpr:
			// address (&X) and indirection (*X)
			if e.Op == token.AND || e.Op == token.MUL {
				expr = e.X
				continue
			}
			return expr
		case *ast.StarExpr:
			// pointer expr/type: *X
			expr = e.X
			continue
		case *ast.ArrayType:
			// slice/array: unwrap to element
			expr = e.Elt
			continue
		case *ast.ParenExpr:
			// (T) -> T
			expr = e.X
			continue
		default:
			return expr
		}
	}
}
