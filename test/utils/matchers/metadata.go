package matchers

import (
	"github.com/gopher-fleece/gleece/core/metadata"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

type FieldDesc struct {
	Name     string
	TypeName string
}

func HaveStructFields(fields []FieldDesc) types.GomegaMatcher {
	return WithTransform(
		func(structMeta metadata.StructMeta) []FieldDesc {
			fields := []FieldDesc{}
			for _, field := range structMeta.Fields {
				fields = append(fields, FieldDesc{Name: field.Name, TypeName: field.Type.Name})
			}
			return fields
		},
		ContainElements(fields),
	)
}
