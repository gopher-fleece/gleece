package graphs

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/gopher-fleece/gleece/gast"
)

const UniverseTypeSymKeyPrefix = "UniverseType:"

// SymbolKey uniquely identifies a symbol by its AST node and file version.
type SymbolKey struct {
	Name       string
	Position   token.Pos
	FileId     string
	FilePath   string
	IsUniverse bool
	IsBuiltIn  bool
}

// BaseId returns a semi-unique ID for the symbol key.
// This ID encapsulates the origin file path but not its version and is used for symbol de-duplication.
func (sk SymbolKey) BaseId() string {
	return sk.formatId(sk.FilePath)
}

// Id returns a unique ID for the symbol key.
// This ID also encapsulates the specific file version the symbol originated from.
func (sk SymbolKey) Id() string {
	return sk.formatId(sk.FileId)
}

// formatId creates a formatter ID for use in graphs
// fileIdPart is a string that represents the origin file - it's intended to either be completely unique
// or point to a specific file to allow node deduplication
func (sk SymbolKey) formatId(fileIdPart string) string {
	if sk.IsUniverse {
		return fmt.Sprintf("%s%s", UniverseTypeSymKeyPrefix, sk.Name)
	}

	if sk.Name != "" {
		return fmt.Sprintf("%s@%d@%s", sk.Name, sk.Position, fileIdPart)
	}

	return fmt.Sprintf("@%d@%s", sk.Position, fileIdPart)
}

func (sk SymbolKey) ShortLabel() string {
	file := filepath.Base(strings.Split(sk.FileId, "|")[0])
	hash := strings.Split(sk.FileId, "|")
	shortHash := ""
	if len(hash) == 3 {
		shortHash = hash[2]
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}
	}
	if sk.Name != "" {
		return fmt.Sprintf("%s@%s|%s", sk.Name, file, shortHash)
	}
	return fmt.Sprintf("%s|%s", file, shortHash)
}

func (sk SymbolKey) PrettyPrint() string {
	if sk.IsUniverse {
		// Strip prefix if present
		return strings.TrimPrefix(sk.Name, UniverseTypeSymKeyPrefix)
	}

	var sb strings.Builder

	// Name or fallback to position
	if sk.Name != "" {
		sb.WriteString(fmt.Sprintf("%s\n", sk.Name))
	} else {
		sb.WriteString(fmt.Sprintf("@%d\n", sk.Position))
	}

	// The FileId is expected to be something like "path|mod|hash"
	fVerParts := strings.Split(sk.FileId, "|")
	for _, part := range fVerParts {
		sb.WriteString(fmt.Sprintf("    • %s\n", part))
	}

	return sb.String()
}

func (sk SymbolKey) Equals(other SymbolKey) bool {
	if sk.IsUniverse {
		return other.IsUniverse && sk.Name == other.Name
	}

	return sk.Name == other.Name && sk.Position == other.Position && sk.FileId == other.FileId
}

func NewSymbolKey(node ast.Node, version *gast.FileVersion) SymbolKey {
	if node == nil || version == nil {
		return SymbolKey{}
	}

	base := version.String()

	name := ""
	pos := token.NoPos

	switch n := node.(type) {
	case *ast.FuncDecl:
		name = n.Name.Name
		pos = n.Pos()
	case *ast.TypeSpec:
		name = n.Name.Name
		pos = n.Pos()
	case *ast.ValueSpec:
		names := make([]string, len(n.Names))
		for i, id := range n.Names {
			names[i] = id.Name
		}
		name = strings.Join(names, ",")
		pos = n.Pos()
	case *ast.Field:
		if len(n.Names) > 0 {
			name = n.Names[0].Name // Problematic for multi-field declaration. Ignoring that for now.
		}
		pos = n.Pos()
	case *ast.Ident:
		name = n.Name
		pos = n.Pos()
	default:
		pos = n.Pos()
	}

	return SymbolKey{
		Name:     name,
		Position: pos,
		FileId:   base,
		FilePath: version.Path,
	}
}

func NewUniverseSymbolKey(typeName string) SymbolKey {
	return SymbolKey{
		Name:       typeName,
		IsUniverse: true,
		IsBuiltIn:  true,
	}
}

func NewNonUniverseBuiltInSymbolKey(typeName string) SymbolKey {
	return SymbolKey{
		Name:       typeName,
		IsUniverse: false,
		IsBuiltIn:  true,
	}
}
