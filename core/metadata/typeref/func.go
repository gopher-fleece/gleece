package typeref

import (
	"fmt"
	"strings"

	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
)

// FuncTypeRef
type FuncTypeRef struct {
	Params   []metadata.TypeRef
	Results  []metadata.TypeRef
	Variadic bool
}

func (f *FuncTypeRef) Kind() metadata.TypeRefKind { return metadata.TypeRefKindFunc }

func (f *FuncTypeRef) CanonicalString() string {
	return f.stringRepresentation(true)
}

func (f *FuncTypeRef) SimpleTypeString() string {
	return f.stringRepresentation(false)
}

func (f *FuncTypeRef) stringRepresentation(canonical bool) string {
	params := make([]string, 0, len(f.Params))

	for _, p := range f.Params {
		if canonical {
			params = append(params, p.CanonicalString())
		} else {
			params = append(params, p.SimpleTypeString())
		}
	}

	rs := make([]string, 0, len(f.Results))
	for _, r := range f.Results {
		if canonical {
			rs = append(rs, r.CanonicalString())
		} else {
			rs = append(rs, r.SimpleTypeString())
		}
	}

	return fmt.Sprintf("func(%s)(%s)", strings.Join(params, ","), strings.Join(rs, ","))
}

func (f *FuncTypeRef) ToSymKey(fileVersion *gast.FileVersion) (graphs.SymbolKey, error) {
	parts := make([]graphs.SymbolKey, 0, len(f.Params)+len(f.Results))
	for _, p := range f.Params {
		k, err := p.ToSymKey(fileVersion)
		if err != nil {
			return graphs.SymbolKey{}, err
		}
		parts = append(parts, k)
	}
	for _, r := range f.Results {
		k, err := r.ToSymKey(fileVersion)
		if err != nil {
			return graphs.SymbolKey{}, err
		}
		parts = append(parts, k)
	}
	return graphs.NewCompositeTypeKey(graphs.CompositeKindFunc, fileVersion, parts), nil
}

func (f *FuncTypeRef) Flatten() []metadata.TypeRef {
	return flatten(f)
}
