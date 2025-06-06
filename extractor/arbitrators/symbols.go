package arbitrators

import (
	"go/ast"
	"go/token"

	"github.com/gopher-fleece/gleece/gast"
)

type DeclInfo struct {
	FVersion  *gast.FileVersion
	Pos       token.Pos // Start position in token.FileSet
	ByteStart int       // Start offset (optional, useful for editors/tools)
	ByteEnd   int       // End offset
}

type FileAstCache struct {
	Files map[string]*ast.File
}

type SymbolLocationCache struct {
	Funcs map[*ast.FuncDecl]*DeclInfo
	Types map[*ast.TypeSpec]*DeclInfo
}

type PositionIndexCache struct {
	ByFile map[string][]SymbolSpan // file path → ordered spans
}

type SymbolSpan struct {
	Start  token.Pos
	End    token.Pos
	Symbol ast.Node
}
