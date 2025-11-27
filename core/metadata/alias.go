package metadata

import (
	"github.com/gopher-fleece/gleece/common"
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
func (m AliasMeta) Reduce(ctx ReductionContext) (definitions.StructMetadata, error) {
	return definitions.StructMetadata{
		Name:        m.Name,
		PkgPath:     m.PkgPath,
		Description: annotations.GetDescription(m.Annotations),
		Fields: []definitions.FieldMetadata{{
			Type:        m.Type.Name,
			Description: annotations.GetDescription(m.Type.Annotations),
			Tag:         annotations.GetTag(m.Type.Annotations),
			IsEmbedded:  true,
			Deprecation: common.Ptr(GetDeprecationOpts(m.Type.Annotations)),
		}},
		Deprecation: GetDeprecationOpts(m.Annotations),
	}, nil
}
