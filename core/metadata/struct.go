package metadata

import (
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/definitions"
)

type StructMeta struct {
	SymNodeMeta
	Fields []FieldMeta
}

func (s StructMeta) Reduce() definitions.StructMetadata {
	reducedFields := make([]definitions.FieldMetadata, len(s.Fields))
	for idx, field := range s.Fields {
		reducedFields[idx] = field.Reduce()
	}

	return definitions.StructMetadata{
		Name:        s.Name,
		PkgPath:     s.PkgPath,
		Description: annotations.GetDescription(s.Annotations),
		Fields:      reducedFields,
		Deprecation: GetDeprecationOpts(s.Annotations),
	}
}
