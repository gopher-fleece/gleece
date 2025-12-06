package visitors

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/packages"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/metadata/typeref"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
)

// AliasVisitor specializes in 'Alias'-type AST structures.
//
// It parses declarations like
//
//	type Foo = Bar
//
// and stores the HIR in the graph for further use.
type AliasVisitor struct {
	BaseVisitor

	declVisitor      *TypeDeclVisitor
	typeUsageVisitor *TypeUsageVisitor
}

// NewAliasVisitor constructs an AliasVisitor.
func NewAliasVisitor(ctx *VisitContext) (*AliasVisitor, error) {
	v := &AliasVisitor{}
	if err := v.initialize(ctx); err != nil {
		return nil, err
	}
	return v, nil
}

func (v *AliasVisitor) setTypeDeclVisitor(visitor *TypeDeclVisitor) {
	v.declVisitor = visitor
}

func (v *AliasVisitor) setTypeUsageVisitor(visitor *TypeUsageVisitor) {
	v.typeUsageVisitor = visitor
}

// VisitAlias attempts to parse the given GenDecl & TypeSpec as a type alias
//
// For our purposes, both forms of aliasing are valid and equivalent:
//
//	type A = string
//	type A string
//
// both result in the same HIR branch structure and are eventually reduced to the same output.
func (v *AliasVisitor) VisitAlias(
	pkg *packages.Package,
	file *ast.File,
	genDecl *ast.GenDecl,
	spec *ast.TypeSpec,
) (graphs.SymbolKey, error) {
	specName := gast.GetIdentNameOrFallback(spec.Name, "N/A")
	v.enterFmt("VisitAlias %s in %s", specName, pkg.PkgPath)
	defer v.exit()

	// Parse the spec into metadata structures
	aliasMeta, err := v.parseAliasSpec(pkg, file, genDecl, spec)
	if err != nil {
		return graphs.SymbolKey{}, v.getFrozenError("failed to visit alias '%s' - %w", specName, err)
	}

	// Add the alias to the graph
	created, err := v.context.Graph.AddAlias(symboldg.CreateAliasNode{
		Data:        aliasMeta,
		Annotations: aliasMeta.Annotations,
	})

	if err != nil {
		return graphs.SymbolKey{}, v.getFrozenError(
			"failed to add alias '%s' to the symbol graph - %w",
			aliasMeta.Name,
			err,
		)
	}

	if v.context.MetadataCache.AddAlias(&aliasMeta) != nil {
		return graphs.SymbolKey{}, v.getFrozenError(
			"failed to add alias '%s' to the metadata cache - %w",
			aliasMeta.Name,
			err,
		)
	}

	// Derive a canonical symbol key for the alias's underlying type
	targetKey, err := aliasMeta.Type.Root.ToSymKey(aliasMeta.Type.FVersion)
	if err != nil {
		return graphs.SymbolKey{}, v.getFrozenError(
			"failed to obtain SymKey for alias metadata '%s' - %v",
			aliasMeta.Name,
			err,
		)
	}

	// Add an edge from the newly added Alias node to the underlying type
	v.context.Graph.AddEdge(created.Id, targetKey, symboldg.EdgeKindType, nil)

	// Return the new alias node's ID so the caller can add their own edges to it
	return created.Id, nil
}

func (v *AliasVisitor) parseAliasSpec(
	pkg *packages.Package,
	file *ast.File,
	genDecl *ast.GenDecl,
	spec *ast.TypeSpec,
) (metadata.AliasMeta, error) {
	specName := gast.GetIdentNameOrFallback(spec.Name, "N/A")

	fileVersion, fvErr := v.context.MetadataCache.GetFileVersion(file, pkg.Fset)
	if fvErr != nil || fileVersion == nil {
		return metadata.AliasMeta{}, v.getFrozenError(
			"could not obtain file version for alias '%s' - %w",
			specName,
			fvErr,
		)
	}

	holder, err := v.getAnnotations(spec.Doc, genDecl)
	if err != nil {
		return metadata.AliasMeta{}, fmt.Errorf(
			"failed to obtain annotations for alias '%s' - %v",
			specName,
			err,
		)
	}

	// Resolve the right-hand-side (RHS) of the expression
	typeUsage, err := v.resolveTypeDeclRhs(pkg, file, spec)
	if err != nil {
		return metadata.AliasMeta{}, v.getFrozenError(
			"failed to resolve RHS expression for TypeSpec '%s' - %w",
			specName,
			err,
		)
	}

	aliasKind := common.Ternary(spec.Assign != 0, metadata.AliasKindAssigned, metadata.AliasKindTypedef)

	return metadata.AliasMeta{
		SymNodeMeta: metadata.SymNodeMeta{
			Name:        spec.Name.Name,
			Node:        spec,
			SymbolKind:  common.SymKindAlias,
			PkgPath:     pkg.PkgPath,
			FVersion:    fileVersion,
			Annotations: holder,
			Range:       common.ResolveNodeRange(pkg.Fset, spec),
		},
		AliasType: aliasKind,
		Type:      typeUsage,
	}, nil
}

// resolveTypeDeclRhs attempts to resolve the right-hand-side of an alias type declaration,
//
// i.e., 'string' in
//
//	type A string
func (v *AliasVisitor) resolveTypeDeclRhs(
	pkg *packages.Package,
	file *ast.File,
	spec *ast.TypeSpec,
) (metadata.TypeUsageMeta, error) {
	switch spec.Type.(type) {
	case *ast.Ident, *ast.SelectorExpr:
		// An Ident or Selector imply a a primitive or an imported type
		return v.resolveAliasRhsIdentOrSelector(pkg, file, spec.Type)
	default:
		// For composite RHS (map[], []T, func(...), inline struct...), use the generic TypeUsageVisitor
		// to get a complete TypeUsageMeta and materialize any pre-requisite declared nodes
		return v.typeUsageVisitor.VisitExpr(pkg, file, spec.Type, nil)
	}
}

// resolveAliasRhsIdentOrSelector resolves a RHS that is an Ident or Selector.
// Returns a fully-populated TypeUsageMeta representing the aliased type (and materializes declared targets).
func (v *AliasVisitor) resolveAliasRhsIdentOrSelector(
	pkg *packages.Package,
	file *ast.File,
	expr ast.Expr,
) (metadata.TypeUsageMeta, error) {
	resolution, ok, resolveErr := gast.ResolveNamedType(
		pkg,
		file,
		expr,
		v.context.ArbitrationProvider.Pkg().GetPackage,
	)
	if resolveErr != nil {
		return metadata.TypeUsageMeta{}, v.getFrozenError(
			"failed to resolve named type for expression '%v' - %w",
			expr,
			resolveErr,
		)
	}
	if !ok {
		return metadata.TypeUsageMeta{}, fmt.Errorf("could not resolve aliased type expression")
	}

	// Built-ins have an easier materialization flow.
	if resolution.IsBuiltin() {
		return v.processBuiltinTypeResolution(resolution)
	}

	// If the resolved type isn't a built-in and doesn't have an associated file/package/spec,
	// something has gone wrong during resolution and we must halt.
	if resolution.DeclaringAstFile == nil || resolution.DeclaringPackage == nil || resolution.TypeSpec == nil {
		return metadata.TypeUsageMeta{}, v.getFrozenError(
			"incomplete declaration context for aliased type %s",
			resolution.TypeName,
		)
	}

	targetFileVersion, fvErr := v.context.MetadataCache.GetFileVersion(
		resolution.DeclaringAstFile,
		resolution.DeclaringPackage.Fset,
	)
	if fvErr != nil {
		return metadata.TypeUsageMeta{}, v.getFrozenError(
			"could not obtain file version for '%s' - %w",
			gast.GetAstFileName(resolution.DeclaringPackage.Fset, resolution.DeclaringAstFile),
			fvErr,
		)
	}
	if targetFileVersion == nil {
		return metadata.TypeUsageMeta{}, v.getFrozenError(
			"missing file version for declaring file of aliased type %s",
			resolution.TypeName,
		)
	}

	targetKey := graphs.NewSymbolKey(resolution.TypeSpec, targetFileVersion)

	// Ask declVisitor to materialize if needed
	if !v.context.MetadataCache.HasVisited(targetKey) {
		if v.declVisitor == nil {
			return metadata.TypeUsageMeta{}, v.getFrozenError(
				"declaration for %s not materialized and no materializer provided",
				resolution.TypeName,
			)
		}
		if _, err := v.declVisitor.EnsureDeclMaterialized(
			resolution.DeclaringPackage,
			resolution.DeclaringAstFile,
			resolution.GenDecl,
			resolution.TypeSpec,
		); err != nil {
			return metadata.TypeUsageMeta{}, err
		}
	}

	// sanity: target node must now exist
	if !v.context.Graph.Exists(targetKey) {
		return metadata.TypeUsageMeta{}, fmt.Errorf("aliased declarative target %s materialized but not present in graph", targetKey.Id())
	}

	// Build TypeUsageMeta whose Root is a NamedTypeRef pointing to the materialized targetKey.
	named := typeref.NewNamedTypeRef(&targetKey, nil)
	tu := metadata.TypeUsageMeta{
		SymNodeMeta: metadata.SymNodeMeta{
			Name:     resolution.TypeName,
			Node:     nil,
			FVersion: targetFileVersion,
		},
		Root: &named,
	}
	return tu, nil
}

// processBuiltinTypeResolution inserts the given built-in type's resolution and inserts it into
// the symbol graph (as either a 'primitive' or a 'special') and returns a new TypeUsageMeta
// fitting the resolution
func (v *AliasVisitor) processBuiltinTypeResolution(resolution gast.TypeSpecResolution) (metadata.TypeUsageMeta, error) {
	qualifiedTypeName := resolution.GetQualifiedName()

	var symKind common.SymKind
	var symKey graphs.SymbolKey

	if prim, isPrim := common.ToPrimitiveType(qualifiedTypeName); isPrim {
		node := v.context.Graph.AddPrimitive(prim)
		symKind = common.SymKindBuiltin
		symKey = node.Id
	} else if sp, isSp := common.ToSpecialType(qualifiedTypeName); isSp {
		node := v.context.Graph.AddSpecial(sp)

		symKind = common.SymKindSpecialBuiltin
		symKey = node.Id
	} else {
		return metadata.TypeUsageMeta{}, fmt.Errorf(
			"type resolution named '%s' is a universe type but is neither a primitive nor a special",
			qualifiedTypeName,
		)
	}

	named := typeref.NewNamedTypeRef(&symKey, nil)
	usage := metadata.TypeUsageMeta{
		SymNodeMeta: metadata.SymNodeMeta{
			Name:       resolution.TypeName,
			SymbolKind: symKind,
		},
		Import: common.ImportTypeNone,
		Root:   &named,
	}

	return usage, nil
}
