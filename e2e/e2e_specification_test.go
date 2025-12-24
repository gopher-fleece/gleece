package e2e

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/gopher-fleece/gleece/v2/definitions"
	"github.com/gopher-fleece/gleece/v2/generator/swagen/swagtool"
	"github.com/nsf/jsondiff"
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

	It("Should generate a valid 3.0.0 specification", func() {
		for _, engine := range definitions.SupportedRoutingEngineStrings {
			spec := GetEngineSpecification(engine, "openapi/openapi3.0.0.json")
			actual := []byte(spec)
			expected := []byte(expectedOpenapi300)

			areEqual, _ := swagtool.AreJSONsIdentical(actual, expected)

			if !areEqual {
				opts := jsondiff.DefaultConsoleOptions()
				diffType, diffStr := jsondiff.Compare(expected, actual, &opts)

				fmt.Printf("\n❌ Diff for engine '%s' (type: %s):\n%s\n", engine, diffType, diffStr)
			}
			Expect(areEqual).To(BeTrueBecause(
				"Test for engine '%s' in version 3.0.0 yielded a difference between expected and generated spec",
				engine,
			))
		}
	})

	It("Should generate a valid 3.1.0 specification", func() {
		for _, engine := range definitions.SupportedRoutingEngineStrings {
			spec := GetEngineSpecification(engine, "openapi/openapi3.1.0.json")
			actual := []byte(spec)
			expected := []byte(expectedOpenapi310)

			areEqual, _ := swagtool.AreJSONsIdentical(actual, expected)

			if !areEqual {
				opts := jsondiff.DefaultConsoleOptions()
				diffType, diffStr := jsondiff.Compare(expected, actual, &opts)

				fmt.Printf("\n❌ Diff for engine '%s' (type: %s):\n%s\n", engine, diffType, diffStr)
			}
			Expect(areEqual).To(BeTrueBecause(
				"Test for engine '%s' in version 3.1.0 yielded a difference between expected and generated spec",
				engine,
			))
		}
	})
})
