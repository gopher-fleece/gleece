package typeref

import (
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
)

// PtrTypeRef
type PtrTypeRef struct{ Elem metadata.TypeRef }

func (p *PtrTypeRef) Kind() metadata.TypeRefKind { return metadata.TypeRefKindPtr }

func (p *PtrTypeRef) CanonicalString() string {
	return "*" + p.Elem.CanonicalString()
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
