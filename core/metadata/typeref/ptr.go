package typeref

import (
	"github.com/gopher-fleece/gleece/v2/core/metadata"
	"github.com/gopher-fleece/gleece/v2/gast"
	"github.com/gopher-fleece/gleece/v2/graphs"
)

// PtrTypeRef
type PtrTypeRef struct{ Elem metadata.TypeRef }

func (p *PtrTypeRef) Kind() metadata.TypeRefKind { return metadata.TypeRefKindPtr }

func (p *PtrTypeRef) CanonicalString() string {
	return "*" + p.Elem.CanonicalString()
}

func (p *PtrTypeRef) SimpleTypeString() string {
	// Currently, this is a bit of a mess. The models list expects arrays to appear with '[]'
	// but pointers are omitted, i.e., the lack of '*' in the output here is intentional.
	return p.Elem.SimpleTypeString()
}

func (p *PtrTypeRef) CacheLookupKey(fileVersion *gast.FileVersion) (graphs.SymbolKey, error) {
	return p.Elem.CacheLookupKey(fileVersion)
}

func (p *PtrTypeRef) ToSymKey(fileVersion *gast.FileVersion) (graphs.SymbolKey, error) {
	elemKey, err := p.Elem.ToSymKey(fileVersion)
	if err != nil {
		return graphs.SymbolKey{}, err
	}
	return graphs.NewCompositeTypeKey(graphs.CompositeKindPtr, fileVersion, []graphs.SymbolKey{elemKey}), nil
}

func (f *PtrTypeRef) Flatten() []metadata.TypeRef {
	return flatten(f)
}
