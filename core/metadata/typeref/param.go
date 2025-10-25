package typeref

import (
	"fmt"

	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
)

// ParamTypeRef: placeholder inside declaration bodies (e.g. "T").
// Index should be set to declaration position (0-based) whenever possible for deterministic canonicalization.
type ParamTypeRef struct {
	Name  string // original name if available ("T"); used for debugging
	Index int    // -1 if unknown; prefer setting this during parsing of declarations
}

func (p *ParamTypeRef) Kind() metadata.TypeRefKind { return metadata.TypeRefKindParam }

func (p *ParamTypeRef) CanonicalString() string {
	if p.Index >= 0 {
		return fmt.Sprintf("P#%d", p.Index)
	}
	return fmt.Sprintf("P{%s}", p.Name)
}

func (p *ParamTypeRef) ToSymKey(fileVersion *gast.FileVersion) (graphs.SymbolKey, error) {
	if fileVersion == nil {
		return graphs.SymbolKey{}, fmt.Errorf("fileVersion required for ParamTypeRef key")
	}
	return graphs.NewParamSymbolKey(fileVersion, p.Name, p.Index), nil
}

func (f *ParamTypeRef) Flatten() []metadata.TypeRef {
	return flatten(f)
}
