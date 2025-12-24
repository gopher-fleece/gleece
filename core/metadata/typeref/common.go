package typeref

import (
	"fmt"

	"github.com/gopher-fleece/gleece/v2/core/metadata"
	"github.com/gopher-fleece/gleece/v2/graphs"
)

func flatten(root metadata.TypeRef) []metadata.TypeRef {
	switch t := root.(type) {
	case *PtrTypeRef:
		return []metadata.TypeRef{t.Elem}
	case *SliceTypeRef:
		return []metadata.TypeRef{t.Elem}
	case *ArrayTypeRef:
		return []metadata.TypeRef{t.Elem}
	case *MapTypeRef:
		return []metadata.TypeRef{t.Key, t.Value}
	case *FuncTypeRef:
		parts := make([]metadata.TypeRef, 0, len(t.Params)+len(t.Results))
		parts = append(parts, t.Params...)
		parts = append(parts, t.Results...)
		return parts
	case *NamedTypeRef:
		return t.TypeArgs
	case *InlineStructTypeRef:
		out := make([]metadata.TypeRef, 0, len(t.Fields))
		for _, f := range t.Fields {
			out = append(out, f.Type.Root)
		}
		return out
	default:
		return nil
	}
}

// ------------------------- canonicalSymKey helper ----------------------------
// Produce a stable textual identity for a graphs.SymbolKey suitable for canonical strings.
// Use FileId (if present) or FilePath to disambiguate same-name types in different files/packages.
// Builtins/universe use only the name.
func canonicalSymKey(k graphs.SymbolKey) string {
	if k.IsUniverse || k.IsBuiltIn {
		return k.Name
	}
	// Prefer FileId when available; fallback to FilePath
	if k.FileId != "" {
		return fmt.Sprintf("%s|%s", k.Name, k.FileId)
	}
	if k.FilePath != "" {
		return fmt.Sprintf("%s|%s", k.Name, k.FilePath)
	}
	// final fallback
	return k.Name
}
