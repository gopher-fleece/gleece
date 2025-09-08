package metadata_test

import (
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/graphs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - Metadata", func() {
	var _ = Describe("TypeLayer", func() {

		Context("NewPointerLayer", func() {
			It("creates a pointer layer with correct kind", func() {
				layer := metadata.NewPointerLayer()
				Expect(layer.Kind).To(Equal(metadata.TypeLayerKindPointer))
				Expect(layer.KeyType).To(BeNil())
				Expect(layer.ValueType).To(BeNil())
				Expect(layer.BaseTypeRef).To(BeNil())
			})
		})

		Context("NewArrayLayer", func() {
			It("creates an array layer with correct kind", func() {
				layer := metadata.NewArrayLayer()
				Expect(layer.Kind).To(Equal(metadata.TypeLayerKindArray))
				Expect(layer.KeyType).To(BeNil())
				Expect(layer.ValueType).To(BeNil())
				Expect(layer.BaseTypeRef).To(BeNil())
			})
		})

		Context("NewMapLayer", func() {
			It("creates a map layer with correct key and value", func() {
				key := graphs.NewUniverseSymbolKey("string")
				value := graphs.NewUniverseSymbolKey("int")
				layer := metadata.NewMapLayer(&key, &value)

				Expect(layer.Kind).To(Equal(metadata.TypeLayerKindMap))
				Expect(layer.KeyType).To(Equal(&key))
				Expect(layer.ValueType).To(Equal(&value))
				Expect(layer.BaseTypeRef).To(BeNil())
			})
		})

		Context("NewBaseLayer", func() {
			It("creates a base layer with correct base reference", func() {
				base := graphs.NewUniverseSymbolKey("MyStruct")
				layer := metadata.NewBaseLayer(&base)

				Expect(layer.Kind).To(Equal(metadata.TypeLayerKindBase))
				Expect(layer.BaseTypeRef).To(Equal(&base))
				Expect(layer.KeyType).To(BeNil())
				Expect(layer.ValueType).To(BeNil())
			})
		})
	})
})
