package e2e

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/gopher-fleece/gleece/generator/swagen/swagtool"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:embed assets/openapi3.0.0.json
var expectedOpenapi300 string

//go:embed assets/openapi3.1.0.json
var expectedOpenapi310 string

func GetEngineSpecification(engine string, path string) string {
	// Read the file as a string
	content, _ := os.ReadFile(fmt.Sprintf("./%s/%s", engine, path))
	return string(content)
}

var _ = Describe("E2E Specification", func() {

	engines := []string{
		"gin",
		"echo",
		"fiber",
		"chi",
		"mux",
	}

	It("Should generate a valid 3.0.0 specification", func() {
		for _, engine := range engines {
			spec := GetEngineSpecification(engine, "openapi/openapi3.0.0.json")
			areEqual, _ := swagtool.AreJSONsIdentical([]byte(spec), []byte(expectedOpenapi300))
			Expect(areEqual).To(BeTrue())
		}
	})

	It("Should generate a valid 3.1.0 specification", func() {
		for _, engine := range engines {
			spec := GetEngineSpecification(engine, "openapi/openapi3.1.0.json")
			areEqual, _ := swagtool.AreJSONsIdentical([]byte(spec), []byte(expectedOpenapi310))
			Expect(areEqual).To(BeTrue())
		}
	})
})
