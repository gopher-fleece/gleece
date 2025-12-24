package metadata_test

import (
	"github.com/gopher-fleece/gleece/v2/common"
	"github.com/gopher-fleece/gleece/v2/core/metadata"
	"github.com/gopher-fleece/gleece/v2/core/metadata/typeref"
	"github.com/gopher-fleece/gleece/v2/definitions"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - Metadata", func() {
	Describe("FuncReturnValue", func() {
		Context("Reduce", func() {
			It("Returns an error when unable to resolve the underlying type", func() {
				retVal := metadata.FuncReturnValue{
					SymNodeMeta: metadata.SymNodeMeta{
						Name: "RetVal1",
					},
					Type: metadata.TypeUsageMeta{
						SymNodeMeta: metadata.SymNodeMeta{
							Name: "string",
						},
						Root: common.Ptr(
							// Use a broken type ref to trigger an early failure
							typeref.NewNamedTypeRef(nil, nil),
						),
					},
				}

				result, err := retVal.Reduce(metadata.ReductionContext{})
				Expect(err).To(MatchError(ContainSubstring("failed to derive cache lookup symbol key for type usage")))
				Expect(result).To(Equal(definitions.FuncReturnValue{}))
			})
		})
	})
})
