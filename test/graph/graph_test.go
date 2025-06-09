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
		graph := utils.GetGraphByGleeceConfigOrFail()
		Expect(graph).ToNot(BeNil())
		dump := graph.Dump()
		Expect(dump).To(BeEquivalentTo(""))
	})
})

func TestGraphController(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Graph Controller")
}
