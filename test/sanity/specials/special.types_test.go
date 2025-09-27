package specials_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/pipeline"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	"github.com/gopher-fleece/gleece/test/utils/matchers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var pipe pipeline.GleecePipeline
var _ = BeforeSuite(func() {
	pipe = utils.GetPipelineOrFail()
})

var _ = Describe("Special Type Handling", func() {
	Context("Maps", func() {
		It("Correctly handles both primitive and non-primitive maps in request bodies", func() {
			err := pipe.GenerateGraph()
			Expect(err).To(BeNil(), "Could not generate graph during Special Type Handling tests")
			specials := pipe.Graph().FindByKind(common.SymKindSpecialBuiltin)
			Expect(specials).To(HaveLen(3))
			Expect(specials).To(ContainElements(
				matchers.BeAnEquivalentSymbolTo(matchers.ComparableNode{
					Name:       "map:UniverseType:string:UniverseType:int",
					IsUniverse: true,
					IsBuiltIn:  true,
					Kind:       common.SymKindSpecialBuiltin,
				}),

				matchers.BeAnEquivalentSymbolTo(matchers.ComparableNode{
					Name:       "map:UniverseType:string:SomeNestedStruct@",
					IsUniverse: true,
					IsBuiltIn:  true,
					Kind:       common.SymKindSpecialBuiltin,
				}),

				matchers.BeAnEquivalentSymbolTo(matchers.ComparableNode{
					Name:       "error",
					IsUniverse: true,
					IsBuiltIn:  true,
					Kind:       common.SymKindSpecialBuiltin,
				}),
			))
		})

		It("Correctly reduces metadata containing maps", func() {
			err := pipe.GenerateGraph()
			Expect(err).To(BeNil(), "Could not generate graph during Special Type Handling tests")

			ir, err := pipe.GenerateIntermediate()
			Expect(err).To(BeNil())

			Expect(ir.Models.Structs).To(HaveLen(2))
			Expect(ir.Models.Structs[0].Name).To(Equal("ObjectWithMapField"))
			Expect(ir.Models.Structs[0].PkgPath).To(Equal("github.com/gopher-fleece/gleece/test/sanity/specials"))
			Expect(ir.Models.Structs[0].Fields).To(HaveLen(1))
			Expect(ir.Models.Structs[0].Fields[0].Name).To(Equal("MapField"))
			Expect(ir.Models.Structs[0].Fields[0].Name).To(Equal("map[string]int"))
		})
	})
})

func TestSpecialTypeHandling(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Special Type Handling")
}
