package sanity_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/v2/core/pipeline"
	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
	"github.com/gopher-fleece/gleece/v2/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var pipe pipeline.GleecePipeline

var _ = Describe("Edge-Cases Controller", func() {
	_ = BeforeEach(func() {
		pipe = utils.GetPipelineOrFail()
	})

	Context("ReturnsAny", func() {
		It("Generates correct graph", func() {
			err := pipe.GenerateGraph()
			Expect(err).To(BeNil())
		})
	})
})

func TestEdgeCasesController(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Edge-Cases Controller")
}
