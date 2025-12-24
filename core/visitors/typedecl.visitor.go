package visitors

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/packages"

	"github.com/gopher-fleece/gleece/v2/gast"
	"github.com/gopher-fleece/gleece/v2/graphs"
)

// TypeDeclVisitor dispatches type declarations to specialized visitors for processing
type TypeDeclVisitor struct {
	BaseVisitor

	structVisitor *StructVisitor
	enumVisitor   *EnumVisitor
	aliasVisitor  *AliasVisitor
}

// NewTypeDeclVisitor constructs a TypeDeclVisitor and internal sub-visitors.
func NewTypeDeclVisitor(ctx *VisitContext) (*TypeDeclVisitor, error) {
	v := &TypeDeclVisitor{}
	if err := v.initialize(ctx); err != nil {
		return nil, err
	}

	return v, nil
}

func (v *TypeDeclVisitor) setStructVisitor(visitor *StructVisitor) {
	v.structVisitor = visitor
}

func (v *TypeDeclVisitor) setEnumVisitor(visitor *EnumVisitor) {
	v.enumVisitor = visitor
}

func (v *TypeDeclVisitor) setAliasVisitor(visitor *AliasVisitor) {
	v.aliasVisitor = visitor
}

// VisitTypeDecl inspects the TypeSpec and routes to the appropriate decl visitor.
// Returns the resulting symbol key for the type (graph node id) or an error.
func (v *TypeDeclVisitor) VisitTypeDecl(
	pkg *packages.Package,
	file *ast.File,
	genDecl *ast.GenDecl,
	typeSpec *ast.TypeSpec,
) (graphs.SymbolKey, error) {

	v.enterFmt("TypeDeclVisitor.VisitTypeDecl %s in %s", typeSpec.Name.Name, pkg.PkgPath)
	defer v.exit()

	fileVersion, fvErr := v.context.MetadataCache.GetFileVersion(file, pkg.Fset)
	if fvErr != nil || fileVersion == nil {
		return graphs.SymbolKey{}, fmt.Errorf("could not obtain file version for type %s: %w", typeSpec.Name.Name, fvErr)
	}

	// An assigned type, e.g. `type Foo = Bar`
	if typeSpec.Assign != 0 {
		return v.visitAssignedType(pkg, file, fileVersion, genDecl, typeSpec)
	}

	// Dispatch the node to the relevant visitor
	switch typeSpec.Type.(type) {
	case *ast.StructType:
		// StructVisitor is responsible for materializing the struct + fields in graph
		_, symKey, err := v.structVisitor.VisitStructType(pkg, file, genDecl, typeSpec)
		return symKey, err

	case *ast.Ident:
		// Idents here may be enums or assigned aliases.
		// Since Go's AST does not actually have a concept of enum or even an alias, we have to
		// do some heuristics here.
		//
		// We basically try extracting an enum for the declaration; If it works, we're looking at an enum.
		// If it doesn't, we're looking at an alias (or a logic error)
		if gast.IsEnumLike(pkg, typeSpec) {
			_, enumSymKey, err := v.enumVisitor.VisitEnumType(pkg, file, fileVersion, genDecl, typeSpec)
			return enumSymKey, err
		}
		return v.visitAssignedType(pkg, file, fileVersion, genDecl, typeSpec)

	case *ast.SelectorExpr:
		// An alias for a type import like
		//
		// type A = time.Time
		return v.aliasVisitor.VisitAlias(pkg, file, genDecl, typeSpec)

	case *ast.InterfaceType:
		// Currently, the only interface we care about or support is context.Context
		specName := gast.GetIdentNameOrFallback(typeSpec.Name, "")
		if pkg.PkgPath == "context" && specName == "Context" {
			return graphs.NewNonUniverseBuiltInSymbolKey("context.Context"), nil
		}
	}

	return graphs.SymbolKey{}, fmt.Errorf(
		"type-spec '%s' has an unexpected underlying type expression '%v'",
		typeSpec.Name.Name,
		typeSpec.Type,
	)
}

// visitAssignedType attempts to parse an assigned type declaration like
//
//	type Something = string
//
// into either an enum or an alias.
//
// This process is somewhat heuristic due to Go's AST not including a true 'enum' or 'alias' type nodes.
func (v *TypeDeclVisitor) visitAssignedType(
	pkg *packages.Package,
	file *ast.File,
	fileVersion *gast.FileVersion,
	genDecl *ast.GenDecl,
	typeSpec *ast.TypeSpec,
) (graphs.SymbolKey, error) {
	// Aliases are logically a subset of enums. As such, we have to first see if a declaration is an enum before deciding its an alias.
	//
	// As enum parsing is generally pretty expensive we actually try the full enum extraction here instead of having a dedicated check (which wouldn't be much cheaper)
	// If it works, great, we're done. If it doesn't we try parsing as an alias.
	_, enumKey, err := v.enumVisitor.VisitEnumType(pkg, file, fileVersion, genDecl, typeSpec)
	if err == nil {
		return enumKey, nil
	}

	if _, isUnexpectedEntity := err.(UnexpectedEntityError); isUnexpectedEntity {
		return v.aliasVisitor.VisitAlias(pkg, file, genDecl, typeSpec)
	}

	return graphs.SymbolKey{}, fmt.Errorf(
		"enum visitor for type %s yielded an error - %v",
		gast.GetIdentNameOrFallback(typeSpec.Name, "Unknown"),
		err,
	)
}

// EnsureDeclMaterialized guarantees the given TypeSpec is materialized (cached + inserted in graph).
// It is idempotent: repeated calls return quickly. It uses the cache's StartMaterializing/FinishMaterializing
// to avoid recursion with recursive/mutually-recursive types.
func (v *TypeDeclVisitor) EnsureDeclMaterialized(
	pkg *packages.Package,
	file *ast.File,
	genDecl *ast.GenDecl,
	typeSpec *ast.TypeSpec,
) (graphs.SymbolKey, error) {
	v.enterFmt("EnsureDeclMaterialized %s in %s", typeSpec.Name.Name, pkg.PkgPath)
	defer v.exit()

	// Strictly obtain file version for this declaring file.
	fileVersion, fvErr := v.context.MetadataCache.GetFileVersion(file, pkg.Fset)
	if fvErr != nil || fileVersion == nil {
		return graphs.SymbolKey{}, fmt.Errorf("could not obtain file version for type %s: %w", typeSpec.Name.Name, fvErr)
	}

	declKey := graphs.NewSymbolKey(typeSpec, fileVersion)

	// Fast path - already materialized/visited
	if v.context.MetadataCache.HasVisited(declKey) {
		return declKey, nil
	}

	// Try to claim materialization responsibility.
	// If StartMaterializing returns false, the declaration is either already visited
	// or another in-flight materializer is processing it — return the declKey to break recursion.
	if !v.context.MetadataCache.StartMaterializing(declKey) {
		return declKey, nil
	}

	// We are now the materializer for this declaration.
	symKey, err := v.VisitTypeDecl(pkg, file, genDecl, typeSpec)
	if err != nil {
		// Mark the materialization as 'failed'
		v.context.MetadataCache.FinishMaterializing(declKey, false)
		return declKey, err
	}

	// Mark the materialization as 'successful'
	v.context.MetadataCache.FinishMaterializing(declKey, true)
	return symKey, nil
}
