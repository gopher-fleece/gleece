package graphs

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/gopher-fleece/gleece/gast"
)

const UniverseTypeSymKeyPrefix = "UniverseType:"

// SymbolKey uniquely identifies a symbol by its AST node and file version.
type SymbolKey string

func SymbolKeyForUniverseType(name string) SymbolKey {
	return SymbolKey(fmt.Sprintf("%s%s", UniverseTypeSymKeyPrefix, name))
}

// SymbolKeyFor builds a stable key for any AST node of interest
// using its kind, position, and the file version.
func SymbolKeyFor(node ast.Node, version *gast.FileVersion) SymbolKey {
	if node == nil || version == nil {
		return SymbolKey("nil")
	}
	base := version.String() // your "path|mod|hash" string

	switch n := node.(type) {
	case *ast.FuncDecl:
		return SymbolKey(fmt.Sprintf("Func:%s@%s", n.Name.Name, base))

	case *ast.TypeSpec:
		return SymbolKey(fmt.Sprintf("Type:%s@%s", n.Name.Name, base))

	case *ast.ValueSpec:
		// const/var declarations
		// join all names if multiple, e.g. "X,Y"
		names := make([]string, len(n.Names))
		for i, id := range n.Names {
			names[i] = id.Name
		}
		return SymbolKey(fmt.Sprintf("Value:%s@%s", strings.Join(names, ","), base))

	case *ast.Field:
		// parameters or struct fields
		// if the field has a name, use it; otherwise fallback to position
		if len(n.Names) > 0 {
			return SymbolKey(fmt.Sprintf("Field:%s@%s", n.Names[0].Name, base))
		}
		return SymbolKey(fmt.Sprintf("Field@%d@%s", n.Pos(), base))

	case *ast.Ident:
		// bare identifiers
		return SymbolKey(fmt.Sprintf("Ident:%s@%s", n.Name, base))

	case ast.Expr:
		// composite expressions: selector, pointer, array, etc.
		return SymbolKey(fmt.Sprintf("Expr:%T@%d@%s", n, n.Pos(), base))

	default:
		// any other AST node: use its type and position
		return SymbolKey(fmt.Sprintf("%T@%d@%s", n, n.Pos(), base))
	}
}
