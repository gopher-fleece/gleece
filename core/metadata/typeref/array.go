package typeref

import (
	"fmt"

	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
)

// ArrayTypeRef (Len == nil treated like a slice in some contexts; we preserve Len for fixed-arrays)
type ArrayTypeRef struct {
	Len  *int
	Elem metadata.TypeRef
}

func (a *ArrayTypeRef) Kind() metadata.TypeRefKind { return metadata.TypeRefKindArray }

func (a *ArrayTypeRef) CanonicalString() string {
	if a.Len == nil {
		return "[]" + a.Elem.CanonicalString()
	}
	return fmt.Sprintf("[%d]%s", *a.Len, a.Elem.CanonicalString())
}

func (a *ArrayTypeRef) SimpleTypeString() string {
	return "[]" + a.Elem.SimpleTypeString()
}

func (a *ArrayTypeRef) CacheLookupKey(fileVersion *gast.FileVersion) (graphs.SymbolKey, error) {
	return a.Elem.CacheLookupKey(fileVersion)
}

func (a *ArrayTypeRef) ToSymKey(fileVersion *gast.FileVersion) (graphs.SymbolKey, error) {
	elemKey, err := a.Elem.ToSymKey(fileVersion)
	if err != nil {
		return graphs.SymbolKey{}, err
	}
	return graphs.NewCompositeTypeKey(graphs.CompositeKindArray, fileVersion, []graphs.SymbolKey{elemKey}), nil
}

func (f *ArrayTypeRef) Flatten() []metadata.TypeRef {
	return flatten(f)
}
