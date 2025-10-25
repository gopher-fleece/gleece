package typeref

import (
	"fmt"
	"strings"

	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
)

// InlineStructTypeRef represents an inline anonymous struct used directly in expressions.
type InlineStructTypeRef struct {
	// Fields of the anonymous struct. They contain FieldMeta (with TypeUsage inside).
	Fields []metadata.FieldMeta

	// Representative key derived from the struct node position + file version.
	// May be zero if not set.
	RepKey graphs.SymbolKey
}

func (i *InlineStructTypeRef) Kind() metadata.TypeRefKind { return metadata.TypeRefKindInlineStruct }

func (i *InlineStructTypeRef) CanonicalString() string {
	// Build canonical from fields (short, deterministic).
	parts := make([]string, 0, len(i.Fields))
	for _, f := range i.Fields {
		// include name if present, otherwise just type
		if f.Name != "" {
			parts = append(parts, fmt.Sprintf("%s:%s", f.Name, f.Type.Root.CanonicalString()))
		} else {
			parts = append(parts, f.Type.Root.CanonicalString())
		}
	}
	base := fmt.Sprintf("inline{%s}", strings.Join(parts, ","))
	// if we have a representative key, append short key so canonical differs by location
	if !i.RepKey.Equals(graphs.SymbolKey{}) {
		return base + "|" + canonicalSymKey(i.RepKey)
	}
	return base
}

func (in *InlineStructTypeRef) ToSymKey(_ *gast.FileVersion) (graphs.SymbolKey, error) {
	if in == nil {
		return graphs.SymbolKey{}, fmt.Errorf("nil InlineStructTypeRef")
	}
	if in.RepKey.Equals(graphs.SymbolKey{}) {
		return graphs.SymbolKey{}, fmt.Errorf("inline struct missing RepKey")
	}
	return in.RepKey, nil
}

func (f *InlineStructTypeRef) Flatten() []metadata.TypeRef {
	return flatten(f)
}
