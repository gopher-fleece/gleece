package metadata_test

import (
	"fmt"
	"go/types"

	"github.com/gopher-fleece/gleece/v2/core/metadata"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - Metadata", func() {
	var _ = Describe("EnumValueKind", func() {
		Describe("NewEnumValueKind", func() {
			DescribeTable("returns correct EnumValueKind",
				func(kind types.BasicKind, expected metadata.EnumValueKind) {
					result, err := metadata.NewEnumValueKind(kind)
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(Equal(expected))
				},
				Entry("string", types.String, metadata.EnumValueKindString),
				Entry("int", types.Int, metadata.EnumValueKindInt),
				Entry("int8", types.Int8, metadata.EnumValueKindInt8),
				Entry("int16", types.Int16, metadata.EnumValueKindInt16),
				Entry("int32", types.Int32, metadata.EnumValueKindInt32),
				Entry("int64", types.Int64, metadata.EnumValueKindInt64),
				Entry("uint", types.Uint, metadata.EnumValueKindUInt),
				Entry("uint8", types.Uint8, metadata.EnumValueKindUInt8),
				Entry("uint16", types.Uint16, metadata.EnumValueKindUInt16),
				Entry("uint32", types.Uint32, metadata.EnumValueKindUInt32),
				Entry("uint64", types.Uint64, metadata.EnumValueKindUInt64),
				Entry("float32", types.Float32, metadata.EnumValueKindFloat32),
				Entry("float64", types.Float64, metadata.EnumValueKindFloat64),
				Entry("bool", types.Bool, metadata.EnumValueKindBool),
			)

			It("returns error on unsupported kind", func() {
				// types.UnsafePointer is not supported
				unsupported := types.UnsafePointer
				_, err := metadata.NewEnumValueKind(unsupported)
				Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("unsupported basic kind: %v", unsupported))))
			})
		})
	})
})
