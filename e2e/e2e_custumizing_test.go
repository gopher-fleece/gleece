package e2e

import (
	"os"

	"github.com/gopher-fleece/gleece/e2e/assets"
	"github.com/gopher-fleece/gleece/e2e/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("E2E Customizing Spec", func() {
	It("Should set custom template header", func() {
		RunRouterTest(common.RouterTest{
			Name:           "Should set custom template header",
			ExpectedStatus: 200,
			ExpectedBody:   "\"works\"",
			ExpendedHeaders: map[string]string{
				"x-test-header": "test",
				"x-inject":      "true",
			},
			Path:    "/e2e/simple-get",
			Method:  "GET",
			Body:    nil,
			Query:   nil,
			Headers: nil,
		})
	})

	It("Should set custom header by template extension", func() {
		RunRouterTest(common.RouterTest{
			Name:           "Should set custom header by template extension - success case",
			ExpectedStatus: 200,
			ExpectedBody:   "\"works\"",
			ExpendedHeaders: map[string]string{
				"x-test-header":                    "test",
				"x-RouteStartRoutesExtension":      "SimpleGet",
				"x-BeforeOperationRoutesExtension": "SimpleGet",
				"x-AfterOperationRoutesExtension":  "SimpleGet",
				"x-RouteEndRoutesExtension":        "SimpleGet",
				"x-JsonResponseExtension":          "SimpleGet",
				"x-ResponseHeadersExtension":       "SimpleGet",
			},
			Path:    "/e2e/simple-get",
			Method:  "GET",
			Body:    nil,
			Query:   nil,
			Headers: nil,
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should set custom header by template extension - error case",
			ExpectedStatus:      500,
			ExpectedBodyContain: "",
			ExpendedHeaders: map[string]string{
				"x-JsonErrorResponseExtension": "DefaultError",
			},
			Path:    "/e2e/default-error",
			Method:  "GET",
			Body:    nil,
			Query:   nil,
			Headers: nil,
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should set custom header by template extension - invalid bool primitive case",
			ExpectedStatus:      422,
			ExpectedBodyContain: "",
			ExpendedHeaders: map[string]string{
				"x-ParamsValidationErrorResponseExtension": "TestPrimitiveConversions",
			},
			Path:    "/e2e/test-primitive-conversions",
			Method:  "POST",
			Query:   map[string]string{"value1": "60", "value2": "true65", "value3": "10", "value4": "3"},
			Headers: map[string]string{},
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should set custom header by template extension - invalid int primitive case",
			ExpectedStatus:      422,
			ExpectedBodyContain: "",
			ExpendedHeaders: map[string]string{
				"x-ParamsValidationErrorResponseExtension": "TestPrimitiveConversions",
			},
			Path:    "/e2e/test-primitive-conversions",
			Method:  "POST",
			Query:   map[string]string{"value1": "60fff", "value2": "true", "value3": "10", "value4": "3"},
			Headers: map[string]string{},
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should set custom header by template extension - invalid http string case",
			ExpectedStatus:      422,
			ExpectedBodyContain: "",
			ExpendedHeaders: map[string]string{
				"x-RunValidatorExtension": "GetHeaderStartWithLetter",
			},
			Path:   "/e2e/get-header-start-with-letter",
			Method: "GET",
			Body:   nil,
			Query:  nil,
			Headers: map[string]string{
				"headerparam": "1headerParam",
			},
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should set custom header by template extension - invalid http body case",
			ExpectedStatus:      422,
			ExpectedBodyContain: "",
			ExpendedHeaders: map[string]string{
				"x-JsonBodyValidationErrorResponseExtension": "PostWithAllParamsWithBodyRequiredPtr",
			},
			Path:   "/e2e/post-with-all-params-body-required-ptr",
			Method: "POST",
			Body:   assets.BodyInfo{},
			Query:  map[string]string{"queryParam": "queryParam"},
			Headers: map[string]string{
				"headerParam": "headerParam",
			},
		})
	})

	It("Should pass and use custom context from route declaration", func() {
		RunRouterTest(common.RouterTest{
			Name:           "Should pass and use custom context from route declaration",
			ExpectedStatus: 200,
			ExpectedBody:   "\"works\"",
			ExpendedHeaders: map[string]string{
				"x-level": "high",
			},
			Path:    "/e2e/template-context-1",
			Method:  "GET",
			Body:    nil,
			Query:   nil,
			Headers: nil,
		})
	})

	It("Should pass and use multiple custom context from route declaration", func() {
		RunRouterTest(common.RouterTest{
			Name:           "Should pass and use multiple custom context from route declaration",
			ExpectedStatus: 200,
			ExpectedBody:   "\"works\"",
			ExpendedHeaders: map[string]string{
				"x-level": "low",
				"x-mode":  "100",
			},
			Path:    "/e2e/template-context-2",
			Method:  "GET",
			Body:    nil,
			Query:   nil,
			Headers: nil,
		})
	})

	Context("Should set all global extensions", func() {
		ginText, _ := os.ReadFile("./gin/routes/gin.e2e.gleece.go")
		echoText, _ := os.ReadFile("./echo/routes/echo.e2e.gleece.go")
		fiberText, _ := os.ReadFile("./fiber/routes/fiber.e2e.gleece.go")
		chiText, _ := os.ReadFile("./chi/routes/chi.e2e.gleece.go")
		muxText, _ := os.ReadFile("./mux/routes/mux.e2e.gleece.go")

		It("Should have ImportsExtension ", func() {
			ginTemplateText, _ := os.ReadFile("./gin/assets/ImportsExtension.hbs")
			echoTemplateText, _ := os.ReadFile("./echo/assets/ImportsExtension.hbs")
			fiberTemplateText, _ := os.ReadFile("./fiber/assets/ImportsExtension.hbs")
			chiTemplateText, _ := os.ReadFile("./chi/assets/ImportsExtension.hbs")
			muxTemplateText, _ := os.ReadFile("./mux/assets/ImportsExtension.hbs")
			Expect(string(ginText)).To(ContainSubstring(string(ginTemplateText)))
			Expect(string(echoText)).To(ContainSubstring(string(echoTemplateText)))
			Expect(string(fiberText)).To(ContainSubstring(string(fiberTemplateText)))
			Expect(string(chiText)).To(ContainSubstring(string(chiTemplateText)))
			Expect(string(muxText)).To(ContainSubstring(string(muxTemplateText)))
		})

		It("Should have FunctionDeclarationsExtension ", func() {
			ginTemplateText, _ := os.ReadFile("./gin/assets/FunctionDeclarationsExtension.hbs")
			echoTemplateText, _ := os.ReadFile("./echo/assets/FunctionDeclarationsExtension.hbs")
			fiberTemplateText, _ := os.ReadFile("./fiber/assets/FunctionDeclarationsExtension.hbs")
			chiTemplateText, _ := os.ReadFile("./chi/assets/FunctionDeclarationsExtension.hbs")
			muxTemplateText, _ := os.ReadFile("./mux/assets/FunctionDeclarationsExtension.hbs")
			Expect(string(ginText)).To(ContainSubstring(string(ginTemplateText)))
			Expect(string(echoText)).To(ContainSubstring(string(echoTemplateText)))
			Expect(string(fiberText)).To(ContainSubstring(string(fiberTemplateText)))
			Expect(string(chiText)).To(ContainSubstring(string(chiTemplateText)))
			Expect(string(muxText)).To(ContainSubstring(string(muxTemplateText)))
		})

		It("Should have RegisterRoutesExtension ", func() {
			ginTemplateText, _ := os.ReadFile("./gin/assets/RegisterRoutesExtension.hbs")
			echoTemplateText, _ := os.ReadFile("./echo/assets/RegisterRoutesExtension.hbs")
			fiberTemplateText, _ := os.ReadFile("./fiber/assets/RegisterRoutesExtension.hbs")
			chiTemplateText, _ := os.ReadFile("./chi/assets/RegisterRoutesExtension.hbs")
			muxTemplateText, _ := os.ReadFile("./mux/assets/RegisterRoutesExtension.hbs")
			Expect(string(ginText)).To(ContainSubstring(string(ginTemplateText)))
			Expect(string(echoText)).To(ContainSubstring(string(echoTemplateText)))
			Expect(string(fiberText)).To(ContainSubstring(string(fiberTemplateText)))
			Expect(string(chiText)).To(ContainSubstring(string(chiTemplateText)))
			Expect(string(muxText)).To(ContainSubstring(string(muxTemplateText)))
		})

		It("Should have TypeDeclarationsExtension ", func() {
			ginTemplateText, _ := os.ReadFile("./gin/assets/TypeDeclarationsExtension.hbs")
			echoTemplateText, _ := os.ReadFile("./echo/assets/TypeDeclarationsExtension.hbs")
			fiberTemplateText, _ := os.ReadFile("./fiber/assets/TypeDeclarationsExtension.hbs")
			chiTemplateText, _ := os.ReadFile("./chi/assets/TypeDeclarationsExtension.hbs")
			muxTemplateText, _ := os.ReadFile("./mux/assets/TypeDeclarationsExtension.hbs")
			Expect(string(ginText)).To(ContainSubstring(string(ginTemplateText)))
			Expect(string(echoText)).To(ContainSubstring(string(echoTemplateText)))
			Expect(string(fiberText)).To(ContainSubstring(string(fiberTemplateText)))
			Expect(string(chiText)).To(ContainSubstring(string(chiTemplateText)))
			Expect(string(muxText)).To(ContainSubstring(string(muxTemplateText)))
		})
	})
})
