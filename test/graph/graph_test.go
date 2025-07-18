package graph_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Graph Controller", func() {
	It("Creates a valid Symbol Graph", func() {
		pipe := utils.GetPipelineOrFail()
		pipe.Run()
		//Expect(graph).ToNot(BeNil())
		//dotFormat := graph.ToDot()
		//utils.WriteFileByRelativePathOrFail("./dot.txt", []byte(dotFormat))
	})
})

func TestGraphController(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Graph Controller")
}
