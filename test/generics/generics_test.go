package generics_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/core/pipeline"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var pipe pipeline.GleecePipeline

var _ = Describe("Generics Controller", func() {
	BeforeEach(func() {
		pipe = utils.GetPipelineOrFail()
		err := pipe.GenerateGraph()
		Expect(err).To(BeNil())
	})

	Context("RecvWithPrimitiveMapInBody", func() {
		It("Generates correct graph for primitive maps in parameter structs", func() {
			controllerNode, _ := utils.MustFindController(pipe.Graph(), "GenericsController")

			recvNode, _ := utils.MustFindControllerReceiver(
				pipe.Graph(),
				controllerNode,
				"RecvWithPrimitiveMapInBody",
			)

			paramNodes, retValNodes := utils.CollectAssertParamsAndRetVals(pipe.Graph(), recvNode)

			// Parameter tree structure checks
			Expect(paramNodes).To(HaveLen(1))

			bodyParamNode := utils.GetSingularChildNode(pipe.Graph(), paramNodes[0], symboldg.EdgeKindParam)
			bodyTypeNode := utils.GetChildTypeNode(pipe.Graph(), bodyParamNode)
			Expect(bodyTypeNode.Id.Name).To(Equal("BodyWithPrimitiveMap"))
			structMeta := utils.MustStructMeta(bodyTypeNode)
			utils.AssertFieldIsMap(structMeta, "Dict", "string", "int")

			// RetVals tree structure checks
			Expect(retValNodes).To(HaveLen(1))
			retTypeNode := utils.GetChildTypeNode(pipe.Graph(), retValNodes[0])
			Expect(retTypeNode.Id).To(Equal(graphs.NewUniverseSymbolKey("error")))
		})
	})
})

func TestGenericsController(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Generics Controller")
}
