package metadata

import (
	"fmt"

	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/definitions"
)

type StructMeta struct {
	SymNodeMeta
	Fields []FieldMeta
}

func (s StructMeta) Reduce(ctx ReductionContext) (definitions.StructMetadata, error) {
	reducedFields := make([]definitions.FieldMetadata, len(s.Fields))
	for idx, field := range s.Fields {
		reduced, err := field.Reduce(ctx)
		if err != nil {
			return definitions.StructMetadata{}, fmt.Errorf("failed to reduce field '%s' - %v", field.Name, err)
		}
		reducedFields[idx] = reduced
	}

	return definitions.StructMetadata{
		Name:        s.Name,
		PkgPath:     s.PkgPath,
		Description: annotations.GetDescription(s.Annotations),
		Fields:      reducedFields,
		Deprecation: GetDeprecationOpts(s.Annotations),
	}, nil
}
