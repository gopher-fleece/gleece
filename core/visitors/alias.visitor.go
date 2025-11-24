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

// AliasVisitor records `type Foo = Bar` relationships in the symbol graph.
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
func (v *AliasVisitor) VisitAlias(
	pkg *packages.Package,
	file *ast.File,
	genDecl *ast.GenDecl,
	spec *ast.TypeSpec,
) (graphs.SymbolKey, error) {
	v.enterFmt("VisitAlias %s in %s", spec.Name.Name, pkg.PkgPath)
	defer v.exit()

	// 1) strict: obtain alias file version
	fileVersion, fvErr := v.context.MetadataCache.GetFileVersion(file, pkg.Fset)
	if fvErr != nil || fileVersion == nil {
		return graphs.SymbolKey{}, fmt.Errorf("could not obtain file version for alias %s: %w", spec.Name.Name, fvErr)
	}

	// 2) Resolve RHS: If it's a plain ident/selector try the light-weight resolver;
	//    otherwise parse the RHS fully using the typeUsageVisitor.
	var typeUsage metadata.TypeUsageMeta
	var err error
	switch spec.Type.(type) {
	case *ast.Ident, *ast.SelectorExpr:
		typeUsage, err = v.resolveAliasRhsIdentOrSelector(pkg, file, spec.Type)
	default:
		// For composite RHS (map[], []T, func(...), inline struct...), use the TypeUsageVisitor so
		// we get a complete TypeUsageMeta describing the RHS (and materialize any declared nodes).
		if v.typeUsageVisitor == nil {
			return graphs.SymbolKey{}, fmt.Errorf("alias RHS is composite and no typeUsageVisitor provided")
		}
		typeUsage, err = v.typeUsageVisitor.VisitExpr(pkg, file, spec.Type, nil)
	}
	if err != nil {
		return graphs.SymbolKey{}, err
	}

	holder, err := v.getAnnotations(spec.Doc, genDecl)
	if err != nil {
		return graphs.SymbolKey{}, fmt.Errorf("failed to obtain annotations for alias '%s' - %v", spec.Name.Name, err)
	}

	// The alias's kind.
	// Currently, we consider both
	//
	// type A string
	// and
	// type A = string
	//
	// to be functionally equivalent. We do want to keep track of this for any future needs though.
	//
	aliasKind := common.Ternary(spec.Assign != 0, metadata.AliasKindAssigned, metadata.AliasKindTypedef)

	// 3) Build final AliasMeta (fully populated)
	aliasMeta := metadata.AliasMeta{
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
	}

	// 4) AddAlias in fully-formed state (idempotent, clean)
	created, err := v.context.Graph.AddAlias(symboldg.CreateAliasNode{
		Data:        aliasMeta,
		Annotations: aliasMeta.Annotations,
	})
	if err != nil {
		return graphs.SymbolKey{}, err
	}

	// 5) If possible, and if the typeUsage yields a stable symbol key for the aliased target,
	//    add an explicit alias->target edge for easy traversal (optional but usually desirable).
	if targetKey, kerr := aliasMeta.Type.Root.ToSymKey(aliasMeta.Type.FVersion); kerr == nil {
		v.context.Graph.AddEdge(created.Id, targetKey, symboldg.EdgeKindType, nil)
	}

	return created.Id, nil
}

// resolveAliasRhsIdentOrSelector resolves a RHS that is an Ident or Selector.
// Returns a fully-populated TypeUsageMeta representing the aliased type (and materializes declared targets).
func (v *AliasVisitor) resolveAliasRhsIdentOrSelector(
	pkg *packages.Package,
	file *ast.File,
	expr ast.Expr,
) (metadata.TypeUsageMeta, error) {

	// Use the existing resolver to get the named-type resolution
	resolution, ok, resolveErr := gast.ResolveNamedType(
		pkg,
		file,
		expr,
		v.context.ArbitrationProvider.Pkg().GetPackage,
	)
	if resolveErr != nil {
		return metadata.TypeUsageMeta{}, resolveErr
	}
	if !ok {
		return metadata.TypeUsageMeta{}, fmt.Errorf("could not resolve aliased type expression")
	}

	// Built-ins have an easier materialization flow.
	if resolution.IsBuiltin() {
		return v.processUniverseTypeResolution(resolution)
	}

	// Declared target: ensure target is materialized (requires declVisitor)
	if resolution.DeclaringAstFile == nil || resolution.DeclaringPackage == nil || resolution.TypeSpec == nil {
		return metadata.TypeUsageMeta{}, fmt.Errorf("incomplete declaration context for aliased type %s", resolution.TypeName)
	}

	targetFileVersion, fvErr := v.context.MetadataCache.GetFileVersion(
		resolution.DeclaringAstFile,
		resolution.DeclaringPackage.Fset,
	)
	if fvErr != nil {
		return metadata.TypeUsageMeta{}, fvErr
	}
	if targetFileVersion == nil {
		return metadata.TypeUsageMeta{}, fmt.Errorf("missing file version for declaring file of aliased type %s", resolution.TypeName)
	}

	targetKey := graphs.NewSymbolKey(resolution.TypeSpec, targetFileVersion)

	// Ask declVisitor to materialize if needed
	if !v.context.MetadataCache.HasVisited(targetKey) {
		if v.declVisitor == nil {
			return metadata.TypeUsageMeta{}, fmt.Errorf("declaration for %s not materialized and no materializer provided", resolution.TypeName)
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

func (v *AliasVisitor) processUniverseTypeResolution(resolution gast.TypeSpecResolution) (metadata.TypeUsageMeta, error) {
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
