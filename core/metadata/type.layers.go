package metadata

import (
	"github.com/gopher-fleece/gleece/graphs"
)

// TypeLayerKind represents the kind of syntactic type layer.
type TypeLayerKind string

const (
	TypeLayerKindPointer TypeLayerKind = "pointer"
	TypeLayerKindArray   TypeLayerKind = "array"
	TypeLayerKindMap     TypeLayerKind = "map"
	TypeLayerKindBase    TypeLayerKind = "base"
)

// TypeLayer represents one syntactic layer of a type.
// Types are modeled from outermost to innermost.
type TypeLayer struct {
	Kind TypeLayerKind

	// For map types:
	KeyType   *graphs.SymbolKey
	ValueType *graphs.SymbolKey

	// For base types:
	BaseTypeRef *graphs.SymbolKey // Yo, what if this is a map?
}

// NewPointerLayer creates a pointer type layer (*T).
func NewPointerLayer() TypeLayer {
	return TypeLayer{Kind: TypeLayerKindPointer}
}

// NewArrayLayer creates an array/slice type layer ([]T).
func NewArrayLayer() TypeLayer {
	return TypeLayer{Kind: TypeLayerKindArray}
}

// NewMapLayer creates a map type layer (map[K]V).
func NewMapLayer(key, value *graphs.SymbolKey) TypeLayer {
	return TypeLayer{
		Kind:      TypeLayerKindMap,
		KeyType:   key,
		ValueType: value,
	}
}

// NewBaseLayer creates the base type layer (e.g., a struct, interface, primitive, etc).
func NewBaseLayer(base *graphs.SymbolKey) TypeLayer {
	return TypeLayer{
		Kind:        TypeLayerKindBase,
		BaseTypeRef: base,
	}
}
