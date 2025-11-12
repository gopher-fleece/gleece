package visitors

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/packages"

	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
)

// AliasVisitor records `type Foo = Bar` relationships in the symbol graph.
type AliasVisitor struct {
	BaseVisitor
}

// NewAliasVisitor constructs an AliasVisitor.
func NewAliasVisitor(ctx *VisitContext) (*AliasVisitor, error) {
	v := &AliasVisitor{}
	if err := v.initialize(ctx); err != nil {
		return nil, err
	}
	return v, nil
}

// VisitAlias creates an alias SymbolKey for the given TypeSpec and adds an EdgeKindAlias edge
// from the alias → the aliased (target) symbol.
func (v *AliasVisitor) VisitAlias(
	pkg *packages.Package,
	file *ast.File,
	genDecl *ast.GenDecl,
	spec *ast.TypeSpec,
) (graphs.SymbolKey, error) {

	v.enterFmt("VisitAlias %s in %s", spec.Name.Name, pkg.PkgPath)
	defer v.exit()

	// strict: obtain file version for the alias declaration
	fileVersion, fvErr := v.context.MetadataCache.GetFileVersion(file, pkg.Fset)
	if fvErr != nil || fileVersion == nil {
		return graphs.SymbolKey{}, fmt.Errorf("could not obtain file version for alias %s: %w", spec.Name.Name, fvErr)
	}

	aliasKey := graphs.NewSymbolKey(spec, fileVersion)

	// Resolve the aliased type expression (the RHS of the alias)
	resolution, ok, resolveErr := gast.ResolveNamedType(
		pkg,
		file,
		spec.Type,
		v.context.ArbitrationProvider.Pkg().GetPackage,
	)
	if resolveErr != nil {
		return graphs.SymbolKey{}, resolveErr
	}
	if !ok {
		return graphs.SymbolKey{}, fmt.Errorf("could not resolve aliased type for %s", spec.Name.Name)
	}

	var targetKey graphs.SymbolKey
	// If universe/builtin, use universe key
	if resolution.IsUniverse {
		targetKey = graphs.NewUniverseSymbolKey(resolution.TypeName)
	} else {
		// Declared type must have DeclaringAstFile/Package and TypeSpec
		if resolution.DeclaringAstFile == nil || resolution.DeclaringPackage == nil || resolution.TypeSpec == nil {
			return graphs.SymbolKey{}, fmt.Errorf("incomplete resolution for aliased type of %s", spec.Name.Name)
		}
		targetFileVersion, fvErr := v.context.MetadataCache.GetFileVersion(
			resolution.DeclaringAstFile,
			resolution.DeclaringPackage.Fset,
		)
		if fvErr != nil {
			return graphs.SymbolKey{}, fvErr
		}
		if targetFileVersion == nil {
			return graphs.SymbolKey{}, fmt.Errorf("missing file version for aliased type declaring file of %s", spec.Name.Name)
		}
		targetKey = graphs.NewSymbolKey(resolution.TypeSpec, targetFileVersion)
	}

	// Add alias edge: aliasKey -> targetKey
	// NOTE: EdgeKindAlias should be the alias edge constant in your graph package.
	v.context.Graph.AddEdge(aliasKey, targetKey, symboldg.EdgeKindAlias, nil)

	return aliasKey, nil
}
