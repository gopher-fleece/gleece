package arbitrators

import (
	"go/ast"
	"go/token"

	"github.com/gopher-fleece/gleece/definitions"
)

type FileAstCache struct {
	Files map[string]*ast.File
}

type SymbolLocationCache struct {
	Funcs map[*ast.FuncDecl]*definitions.DeclInfo
	Types map[*ast.TypeSpec]*definitions.DeclInfo
}

type PositionIndexCache struct {
	ByFile map[string][]SymbolSpan // file path → ordered spans
}

type SymbolSpan struct {
	Start  token.Pos
	End    token.Pos
	Symbol ast.Node
}

