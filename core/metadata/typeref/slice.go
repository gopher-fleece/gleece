package typeref

import (
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
)

// SliceTypeRef
type SliceTypeRef struct{ Elem metadata.TypeRef }

func (s *SliceTypeRef) Kind() metadata.TypeRefKind { return metadata.TypeRefKindSlice }

func (s *SliceTypeRef) CanonicalString() string {
	return "[]" + s.Elem.CanonicalString()
}

func (s *SliceTypeRef) ToSymKey(fileVersion *gast.FileVersion) (graphs.SymbolKey, error) {
	elemKey, err := s.Elem.ToSymKey(fileVersion)
	if err != nil {
		return graphs.SymbolKey{}, err
	}
	return graphs.NewCompositeTypeKey(graphs.CompositeKindSlice, fileVersion, []graphs.SymbolKey{elemKey}), nil
}

func (f *SliceTypeRef) Flatten() []metadata.TypeRef {
	return flatten(f)
}
