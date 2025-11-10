package generics_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gopher-fleece/gleece/core/pipeline"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var pipe pipeline.GleecePipeline

var _ = BeforeSuite(func() {
	pipe = utils.GetPipelineOrFail()
})

var _ = Describe("Generics Controller", func() {
	Context("RecvWithPrimitiveMapInBody", func() {
		It("Bah", func() {
			err := pipe.GenerateGraph()
			Expect(err).To(BeNil())
			fmt.Println(pipe.Graph().ToDot(nil))
			interm, err := pipe.GenerateIntermediate()
			Expect(err).To(BeNil())
			Expect(interm).ToNot(BeNil())
			dump, _ := json.MarshalIndent(interm, "", "\t")
			fmt.Println(string(dump))
		})
	})
})

func TestGenericsController(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Generics Controller")
}
