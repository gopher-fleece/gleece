package metadata

import (
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/definitions"
)

type AliasKind int

const (
	AliasKindTypedef AliasKind = iota
	AliasKindAssigned
)

type AliasMeta struct {
	SymNodeMeta
	AliasType AliasKind
	Type      TypeUsageMeta
}

// Reduce converts the HIR representation of an Alias (type A = string or type A string)
// into a StructMetadata that can be used by the spec emitters to create OAS alias models
func (m AliasMeta) Reduce(_ ReductionContext) (definitions.NakedAliasMetadata, error) {
	return definitions.NakedAliasMetadata{
		Name:        m.Name,
		PkgPath:     m.PkgPath,
		Description: annotations.GetDescription(m.Annotations),
		Type:        m.Type.Name,
		Deprecation: GetDeprecationOpts(m.Annotations),
	}, nil
}
