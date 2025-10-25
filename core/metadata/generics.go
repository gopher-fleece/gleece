package metadata

import (
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
)

type TypeRefKind string

const (
	TypeRefKindNamed        TypeRefKind = "named"
	TypeRefKindParam        TypeRefKind = "param"
	TypeRefKindPtr          TypeRefKind = "ptr"
	TypeRefKindSlice        TypeRefKind = "slice"
	TypeRefKindArray        TypeRefKind = "array"
	TypeRefKindMap          TypeRefKind = "map"
	TypeRefKindFunc         TypeRefKind = "func"
	TypeRefKindInlineStruct TypeRefKind = "inline_struct"
)

// ------------------------------ TypeRef interface ----------------------------
type TypeRef interface {
	Kind() TypeRefKind
	// Deterministic structural representation used for interning/canonicalization.
	CanonicalString() string

	ToSymKey(fileVersion *gast.FileVersion) (graphs.SymbolKey, error)
	Flatten() []TypeRef
}

// TypeParamDecl is a minimal declaration-side record for a type parameter.
type TypeParamDecl struct {
	Name       string  // "T" - original name (optional, but helpful for debugging)
	Index      int     // 0-based index in the declaration; prefer populating this
	Constraint TypeRef // optional constraint (interface, union, etc.) - nil == any
}
