package graphs

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gopher-fleece/gleece/common/linq"
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

// ShortLabel returns a compact, human-friendly label used in DOT dumps.
// It delegates to small helpers for composite formatting and file info attachment.
func (sk SymbolKey) ShortLabel() string {
	fileBase, shortHash := extractFileInfo(sk)

	attach := func(label string) string {
		if fileBase == "" {
			return label
		}
		if shortHash != "" {
			return fmt.Sprintf("%s @%s|%s", label, fileBase, shortHash)
		}
		return fmt.Sprintf("%s @%s", label, fileBase)
	}

	// universe / builtin -> simple name
	if sk.IsUniverse || sk.IsBuiltIn {
		return attach(sk.Name)
	}

	// composite nodes have special pretty rendering
	if strings.HasPrefix(sk.Name, "comp:") {
		return attach(compositeLabelFromName(sk.Name))
	}

	// default: name + file info
	if sk.Name != "" {
		return attach(sk.Name)
	}

	// fallback to file info only
	if fileBase != "" {
		if shortHash != "" {
			return fmt.Sprintf("%s|%s", fileBase, shortHash)
		}
		return fileBase
	}

	return "?"
}

// ---------------------- small helpers ----------------------

// extractFileInfo extracts file basename and short hash from FileId or FilePath.
// FileId expected format: "path|mod|hash". Returns possibly-empty strings.
func extractFileInfo(sk SymbolKey) (fileBase, shortHash string) {
	if sk.FileId != "" {
		parts := strings.Split(sk.FileId, "|")
		if len(parts) >= 1 {
			fileBase = filepath.Base(parts[0])
		}
		if len(parts) >= 3 {
			shortHash = parts[2]
			if len(shortHash) > 7 {
				shortHash = shortHash[:7]
			}
		}
		return fileBase, shortHash
	}
	if sk.FilePath != "" {
		return filepath.Base(sk.FilePath), ""
	}
	return "", ""
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

// CompositeKind identifies a composite type family.
type CompositeKind string

const (
	CompositeKindPtr   CompositeKind = "ptr"
	CompositeKindSlice CompositeKind = "slice"
	CompositeKindArray CompositeKind = "array"
	CompositeKindMap   CompositeKind = "map"
	CompositeKindFunc  CompositeKind = "func"
)

// NewInstSymbolKey returns a canonical SymbolKey representing an instantiation:
//
//	inst:<base-id>[<arg-id>,...]
//
// The returned key uses the base's FileId so instantiated keys for the same base
// are dedupable and stable across usage sites.
func NewInstSymbolKey(base SymbolKey, argKeys []SymbolKey) SymbolKey {
	argIds := make([]string, 0, len(argKeys))
	for i := range argKeys {
		argIds[i] = argKeys[i].Id()
	}
	// Build canonical name using base.Id() and arg ids.
	name := "inst:" + base.Id() + "[" + strings.Join(argIds, ",") + "]"

	return SymbolKey{
		Name:     name,
		Position: token.NoPos,
		FileId:   base.FileId, // scope instantiated type to base's declaring file id
		FilePath: base.FilePath,
	}
}

// NewCompositeTypeKey returns a canonical SymbolKey for composites like ptr/slice/map/func.
// The canonical Name embeds operand Ids; FileId is derived from fileVersion (if provided).
func NewCompositeTypeKey(kind CompositeKind, fileVersion *gast.FileVersion, operands []SymbolKey) SymbolKey {
	operandIds := linq.Map(operands, func(op SymbolKey) string {
		return op.Id()
	})

	name := "comp:" + string(kind) + "[" + strings.Join(operandIds, ",") + "]"
	var fileId string
	var filePath string
	if fileVersion != nil {
		fileId = fileVersion.String()
		filePath = fileVersion.Path
	}
	return SymbolKey{
		Name:     name,
		Position: token.NoPos,
		FileId:   fileId,
		FilePath: filePath,
	}
}

// NewParamSymbolKey returns a stable key for a type parameter occurrence scoped to the given fileVersion.
func NewParamSymbolKey(fileVersion *gast.FileVersion, paramName string, index int) SymbolKey {
	var fileId string
	var filePath string
	if fileVersion != nil {
		fileId = fileVersion.String()
		filePath = fileVersion.Path
	}
	name := "typeparam:" + paramName + "#" + strconv.Itoa(index)
	return SymbolKey{
		Name:     name,
		Position: token.NoPos,
		FileId:   fileId,
		FilePath: filePath,
	}
}

// compositeLabelFromName pretty-prints composite symbol names produced by your canonicalizer.
// Expected input examples:
//
//	"comp:slice[SimpleStruct@...]" -> "[]SimpleStruct"
//	"comp:map[Key@... , Val@...]" -> "map[Key]Val"
func compositeLabelFromName(comp string) string {
	inner := strings.TrimPrefix(comp, "comp:")
	kind, args := parseCompositeName(inner)

	trim := func(s string) string {
		// keep the leftmost token up to '@' if present
		if at := strings.Index(s, "@"); at >= 0 {
			return s[:at]
		}
		return s
	}

	switch kind {
	case "slice":
		return "[]" + trim(args)
	case "array":
		// args might encode length or inner type; show compactly
		return "[" + trim(args) + "]"
	case "ptr":
		return "*" + trim(args)
	case "map":
		// args expected "Key,Val" (or "Key@... , Val@...") -> compact "map[Key]Val"
		parts := splitArgs(args, 2)
		if len(parts) == 2 {
			return fmt.Sprintf("map[%s]%s", trim(parts[0]), trim(parts[1]))
		}
		return "map[" + trim(args) + "]"
	case "func":
		// show compact func signature (not fully pretty for complex cases)
		parts := splitArgs(args, 2)
		if len(parts) == 2 {
			return fmt.Sprintf("func(%s)(%s)", trim(parts[0]), trim(parts[1]))
		}
		return "func(" + trim(args) + ")"
	default:
		// unknown composite kind: return short inner
		return trim(args)
	}
}

// parseCompositeName splits "kind[args]" into kind and args.
func parseCompositeName(name string) (kind string, args string) {
	br := strings.Index(name, "[")
	if br < 0 {
		return name, ""
	}
	kind = name[:br]
	args = name[br+1:]
	args = strings.ReplaceAll(args, "UniverseType:", "")
	args = strings.TrimSuffix(args, "]")
	return kind, args
}

// splitArgs splits a comma-separated argument list but only up to n parts.
// It is permissive about spaces.
func splitArgs(s string, n int) []string {
	if s == "" {
		return nil
	}
	parts := strings.SplitN(s, ",", n)
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
