package typeref

import (
	"fmt"

	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
)

// MapTypeRef
type MapTypeRef struct {
	Key, Value metadata.TypeRef
}

func (m *MapTypeRef) Kind() metadata.TypeRefKind { return metadata.TypeRefKindMap }

func (m *MapTypeRef) CanonicalString() string {
	return fmt.Sprintf("map[%s]%s", m.Key.CanonicalString(), m.Value.CanonicalString())
}

func (m *MapTypeRef) SimpleTypeString() string {
	return fmt.Sprintf("map[%s]%s", m.Key.SimpleTypeString(), m.Value.SimpleTypeString())
}

func (m *MapTypeRef) CacheLookupKey(fileVersion *gast.FileVersion) (graphs.SymbolKey, error) {
	return m.ToSymKey(fileVersion)
}

func (m *MapTypeRef) ToSymKey(fileVersion *gast.FileVersion) (graphs.SymbolKey, error) {
	keyK, err := m.Key.ToSymKey(fileVersion)
	if err != nil {
		return graphs.SymbolKey{}, err
	}
	valK, err := m.Value.ToSymKey(fileVersion)
	if err != nil {
		return graphs.SymbolKey{}, err
	}
	return graphs.NewCompositeTypeKey(graphs.CompositeKindMap, fileVersion, []graphs.SymbolKey{keyK, valK}), nil
}

func (f *MapTypeRef) Flatten() []metadata.TypeRef {
	return flatten(f)
}
