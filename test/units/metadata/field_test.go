package metadata_test

import (
	"github.com/gopher-fleece/gleece/core/metadata"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - Metadata - Fields", func() {
	Context("Reduce", func() {
		It("Returns an error if the encapsulated node is not an AST Field", func() {
			f := metadata.FieldMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Name: "F",
					Node: nil,
				},
			}

			_, err := f.Reduce(metadata.ReductionContext{})
			Expect(err).To(MatchError(Equal("field 'F' has a non-field node type")))
		})
	})
})
