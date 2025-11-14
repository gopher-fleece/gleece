package utils

import (
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
)

type FakeTypeRef struct {
	RefKind         metadata.TypeRefKind
	CanonicalStr    string
	SimpleStr       string
	SymKey          graphs.SymbolKey
	ToSymKeyErr     error
	FlattenResponse []metadata.TypeRef
}

func (f *FakeTypeRef) Kind() metadata.TypeRefKind { return f.RefKind }
func (f *FakeTypeRef) CanonicalString() string    { return f.CanonicalStr }
func (f *FakeTypeRef) SimpleTypeString() string   { return f.SimpleStr }
func (f *FakeTypeRef) CacheLookupKey(_ *gast.FileVersion) (graphs.SymbolKey, error) {
	return f.SymKey, nil
}
func (f *FakeTypeRef) ToSymKey(_ *gast.FileVersion) (graphs.SymbolKey, error) {
	if f.ToSymKeyErr != nil {
		return graphs.SymbolKey{}, f.ToSymKeyErr
	}
	return f.SymKey, nil
}
func (f *FakeTypeRef) Flatten() []metadata.TypeRef { return f.FlattenResponse }
