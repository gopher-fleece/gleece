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

type TypeRef interface {
	Kind() TypeRefKind
	// Deterministic structural representation used for interning/canonicalization.
	CanonicalString() string

	// SimpleTypeString returns a simplified representation of the reference.
	// An example could be as follows:
	// A MapTypeRef has a string key and an imported value module.SomeStruct.
	// Simple string will return a 'schema-compatible' type string map[string]SomeStruct
	//
	// This can be used to feed the spec generator as it does not require any package/language level information
	SimpleTypeString() string

	// A key to be used for metadata cache lookups.
	// Usually the same as the SimpleTypeString
	CacheLookupKey(fileVersion *gast.FileVersion) (graphs.SymbolKey, error)

	ToSymKey(fileVersion *gast.FileVersion) (graphs.SymbolKey, error)
	Flatten() []TypeRef
}

// TypeParamDecl is a minimal declaration-side record for a type parameter.
type TypeParamDecl struct {
	Name       string  // "T" - original name (optional, but helpful for debugging)
	Index      int     // 0-based index in the declaration; prefer populating this
	Constraint TypeRef // optional constraint (interface, union, etc.) - nil == any
}
