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

	TypeParams []*graphs.SymbolKey

	// For base types:
	BaseTypeRef *graphs.SymbolKey
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
		Kind:       TypeLayerKindMap,
		TypeParams: []*graphs.SymbolKey{key, value},
	}
}

// NewBaseLayer creates the base type layer (e.g., a struct, interface, primitive, etc).
func NewBaseLayer(base *graphs.SymbolKey) TypeLayer {
	return TypeLayer{
		Kind:        TypeLayerKindBase,
		BaseTypeRef: base,
	}
}

func (l TypeLayer) formatGenericLayerParam(syncedProvider IdProvider, paramIndex int) string {
	if l.TypeParams[paramIndex].IsUniverse {
		return l.TypeParams[paramIndex].Name
	}
	return fmt.Sprint(syncedProvider.GetIdForKey(*l.TypeParams[paramIndex]))
}

func (l TypeLayer) FormatToIr(syncedProvider IdProvider) string {
	if l.Kind == TypeLayerKindMap {
		keySpecifier := l.formatGenericLayerParam(syncedProvider, 0)
		valueSpecifier := l.formatGenericLayerParam(syncedProvider, 1)
		return fmt.Sprintf("map[%s]%s", keySpecifier, valueSpecifier)
	}

	formattedTypeParams := []string{}
	for idx := range l.TypeParams {
		formattedTypeParams = append(formattedTypeParams, l.formatGenericLayerParam(syncedProvider, idx))
	}

	typeSuffix := ""
	if len(formattedTypeParams) > 0 {
		typeSuffix = fmt.Sprintf("[%s]", strings.Join(formattedTypeParams, ", "))
	}

	prefix := ""
	switch l.Kind {
	case TypeLayerKindPointer:
		prefix = "*"
	case TypeLayerKindArray:
		prefix = "[]"
	}

	return fmt.Sprintf("%s%s%s", prefix, l.BaseTypeRef.Name, typeSuffix)
}
