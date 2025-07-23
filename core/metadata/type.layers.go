package metadata

import (
	"fmt"
	"strings"

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

func LayersToTypeString(layers []TypeLayer) string {
	var prefixBuilder strings.Builder
	var suffixBuilder strings.Builder

	for _, layer := range layers {
		switch layer.Kind {
		case TypeLayerKindPointer:
			prefixBuilder.WriteString("*")
		case TypeLayerKindArray:
			prefixBuilder.WriteString("[]")
		case TypeLayerKindMap:
			keyStr := formatSymKey(layer.KeyType)
			valStr := formatSymKey(layer.ValueType)
			prefixBuilder.WriteString(fmt.Sprintf("map[%s]%s", keyStr, valStr))
			// map is terminal
			return prefixBuilder.String()
		case TypeLayerKindBase:
			suffixBuilder.WriteString(formatSymKey(layer.BaseTypeRef))
		default:
			suffixBuilder.WriteString("<?>")
		}
	}

	return prefixBuilder.String() + suffixBuilder.String()
}

func formatSymKey(key *graphs.SymbolKey) string {
	if key == nil {
		return "<?>"
	}
	return key.Name
}
