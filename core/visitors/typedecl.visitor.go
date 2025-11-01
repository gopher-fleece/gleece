package visitors

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/packages"

	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
)

// TypeDeclVisitor dispatches type declarations to specialized decl visitors.
type TypeDeclVisitor struct {
	BaseVisitor

	structVisitor *StructVisitor
	enumVisitor   *EnumVisitor
	aliasVisitor  *AliasVisitor // optional; may be nil
}

// NewTypeDeclVisitor constructs a TypeDeclVisitor and internal sub-visitors.
func NewTypeDeclVisitor(ctx *VisitContext) (*TypeDeclVisitor, error) {
	v := &TypeDeclVisitor{}
	if err := v.initialize(ctx); err != nil {
		return nil, err
	}

	// alias visitor is optional: construct only if you want alias behavior
	// av, _ := NewAliasVisitor(ctx) // uncomment if/when you implement AliasVisitor
	// v.aliasVisitor = av

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

	// Must obtain file version for this file (strict).
	fileVersion, fvErr := v.context.MetadataCache.GetFileVersion(file, pkg.Fset)
	if fvErr != nil || fileVersion == nil {
		return graphs.SymbolKey{}, fmt.Errorf("could not obtain file version for type %s: %w", typeSpec.Name.Name, fvErr)
	}

	// Alias: `type Foo = Bar`
	if typeSpec.Assign != 0 {
		return v.visitAssignedType(pkg, file, fileVersion, genDecl, typeSpec)
	}

	// Dispatch by syntactic type
	switch typeSpec.Type.(type) {
	case *ast.StructType:
		// StructVisitor is responsible for materializing the struct + fields in graph
		_, symKey, err := v.structVisitor.VisitStructType(pkg, file, genDecl, typeSpec)
		return symKey, err

	case *ast.Ident:
		// Could be an enum-like alias to a basic type (e.g., type X string)
		if gast.IsEnumLike(pkg, typeSpec) && v.enumVisitor != nil {
			_, enumSymKey, err := v.enumVisitor.VisitEnumType(pkg, file, fileVersion, genDecl, typeSpec)
			return enumSymKey, err
		}
		// Not enum-like (or no enumVisitor): return stable key but don't materialize
		return graphs.NewSymbolKey(typeSpec, fileVersion), nil

	case *ast.SelectorExpr:
		// selector (possibly referencing external type) — not a declaration we can expand here.
		return graphs.NewSymbolKey(typeSpec, fileVersion), nil

	default:
		// Other types (interface, func types, maps, etc.) — don't attempt to materialize here.
		return graphs.NewSymbolKey(typeSpec, fileVersion), nil
	}
}

func (v *TypeDeclVisitor) visitAssignedType(
	pkg *packages.Package,
	file *ast.File,
	fileVersion *gast.FileVersion,
	genDecl *ast.GenDecl,
	typeSpec *ast.TypeSpec,
) (graphs.SymbolKey, error) {
	_, enumKey, err := v.enumVisitor.VisitEnumType(pkg, file, fileVersion, genDecl, typeSpec)
	if err == nil {
		return enumKey, nil
	}

	if _, isUnexpectedEntity := err.(UnexpectedEntityError); isUnexpectedEntity {
		return v.aliasVisitor.VisitAlias(pkg, file, genDecl, typeSpec)
	}

	var tsName string
	if typeSpec.Name != nil {
		tsName = typeSpec.Name.Name
	} else {
		tsName = "Unknown"
	}

	return graphs.SymbolKey{}, fmt.Errorf(
		"enum visitor for type %s yielded an error - %v",
		tsName,
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

	// Fast path: already materialized/visited
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
		// mark finish with failure so retries are possible later
		v.context.MetadataCache.FinishMaterializing(declKey, false)
		return declKey, err
	}

	// Success: finish and ensure visited is set.
	v.context.MetadataCache.FinishMaterializing(declKey, true)
	return symKey, nil
}
