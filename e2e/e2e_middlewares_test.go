package e2e

import (
	"github.com/gopher-fleece/gleece/e2e/assets"
	"github.com/gopher-fleece/gleece/e2e/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("E2E Middlewares Spec", func() {

	It("Should pass succeeded middlewares", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should pass succeeded middlewares",
			ExpectedStatus:      200,
			ExpectedBodyContain: "works",
			Path:                "/e2e/simple-get",
			Method:              "GET",
			Body:                nil,
			Query:               nil,
			Headers:             nil,
			ExpendedHeaders: map[string]string{
				"X-pass-before-operation":        "true",
				"X-pass-on-error":                "",
				"X-pass-after-succeed-operation": "true",
			},
		})
	})

	It("Should pass failed middlewares for default error", func() {
		RunRouterTest(common.RouterTest{
			Name:           "Should pass failed middlewares for default error",
			ExpectedStatus: 500,
			ExpectedBody:   "",
			Path:           "/e2e/default-error",
			Method:         "GET",
			Body:           nil,
			Query:          nil,
			Headers:        nil,
			ExpendedHeaders: map[string]string{
				"X-pass-before-operation":        "true",
				"X-pass-on-error":                "true",
				"X-pass-after-succeed-operation": "",
			},
		})
	})

	It("Should pass failed middlewares for default error with payload", func() {
		RunRouterTest(common.RouterTest{
			Name:           "Should pass failed middlewares for default error with payload",
			ExpectedStatus: 500,
			ExpectedBody:   "",
			Path:           "/e2e/default-error-with-payload",
			Method:         "GET",
			Body:           nil,
			Query:          nil,
			Headers:        nil,
			ExpendedHeaders: map[string]string{
				"X-pass-before-operation":        "true",
				"X-pass-on-error":                "true",
				"X-pass-after-succeed-operation": "",
			},
		})
	})

	It("Should pass failed middlewares for custom error", func() {
		RunRouterTest(common.RouterTest{
			Name:           "Should pass failed middlewares for custom error",
			ExpectedStatus: 500,
			ExpectedBody:   "",
			Path:           "/e2e/custom-error",
			Method:         "GET",
			Body:           nil,
			Query:          nil,
			Headers:        nil,
			ExpendedHeaders: map[string]string{
				"X-pass-before-operation":        "true",
				"X-pass-on-error":                "true",
				"X-pass-after-succeed-operation": "",
			},
		})
	})

	It("Should abort on before operation middlewares", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should abort on before operation middlewares",
			ExpectedStatus:      400,
			ExpectedBodyContain: "abort-before-operation header is set to true",
			Path:                "/e2e/simple-get",
			Method:              "GET",
			Body:                nil,
			Query:               nil,
			Headers:             map[string]string{"abort-before-operation": "true"},
			ExpendedHeaders: map[string]string{
				"X-pass-before-operation":        "true",
				"X-pass-on-error":                "",
				"X-pass-after-succeed-operation": "",
			},
		})
	})

	It("Should abort on after operation middlewares", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should abort on after operation middlewares",
			ExpectedStatus:      400,
			ExpectedBodyContain: "abort-after-operation header is set to true",
			Path:                "/e2e/simple-get",
			Method:              "GET",
			Body:                nil,
			Query:               nil,
			Headers:             map[string]string{"abort-after-operation": "true"},
			ExpendedHeaders: map[string]string{
				"X-pass-before-operation":        "true",
				"X-pass-on-error":                "",
				"X-pass-after-succeed-operation": "true",
			},
		})
	})

	It("Should abort on error operation middlewares", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should abort on error operation middlewares",
			ExpectedStatus:      400,
			ExpectedBodyContain: "abort-on-error header is set to true default error",
			Path:                "/e2e/default-error",
			Method:              "GET",
			Body:                nil,
			Query:               nil,
			Headers:             map[string]string{"abort-on-error": "true"},
			ExpendedHeaders: map[string]string{
				"X-pass-before-operation":        "true",
				"X-pass-on-error":                "true",
				"X-pass-after-succeed-operation": "",
			},
		})
	})

	It("Should abort on custom error operation middlewares", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should abort on custom error operation middlewares",
			ExpectedStatus:      400,
			ExpectedBodyContain: "abort-on-error header is set to true custom error",
			Path:                "/e2e/custom-error",
			Method:              "GET",
			Body:                nil,
			Query:               nil,
			Headers:             map[string]string{"abort-on-error": "true"},
			ExpendedHeaders: map[string]string{
				"X-pass-before-operation":        "true",
				"X-pass-on-error":                "true",
				"X-pass-after-succeed-operation": "",
			},
		})
	})

	It("Should run all on error operation middlewares", func() {
		RunRouterTest(common.RouterTest{
			Name:           "Should run all on error operation middlewares",
			ExpectedStatus: 500,
			ExpectedBody:   "",
			Path:           "/e2e/default-error",
			Method:         "GET",
			Body:           nil,
			Query:          nil,
			Headers:        nil,
			ExpendedHeaders: map[string]string{
				"X-pass-before-operation":        "true",
				"X-pass-on-error":                "true",
				"X-pass-on-error-2":              "true",
				"X-pass-after-succeed-operation": "",
			},
		})
	})

	It("Should abort on after operation middlewares and skip next middleware", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should abort on after operation middlewares and skip next middleware",
			ExpectedStatus:      400,
			ExpectedBodyContain: "abort-after-operation header is set to true",
			Path:                "/e2e/simple-get",
			Method:              "GET",
			Body:                nil,
			Query:               nil,
			Headers:             map[string]string{"abort-after-operation": "true"},
			ExpendedHeaders: map[string]string{
				"X-pass-before-operation":        "true",
				"X-pass-on-error":                "",
				"X-pass-after-succeed-operation": "true",
				"X-pass-on-error-2":              "",
			},
		})
	})

	It("Should pass thro validation error middleware for string param", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should pass thro validation error middleware for non-sent string param",
			ExpectedStatus:      422,
			ExpectedBodyContain: "A request was made to operation 'GetWithAllParamsRequiredPtr' but parameter 'headerParam' did not pass validation - Field 'headerParam' failed validation with tag 'required'",
			ExpendedHeaders: map[string]string{
				"X-pass-error-validation":        "true",
				"X-pass-after-succeed-operation": "",
			},
			Path:    "/e2e/get-with-all-params-required-ptr/pathParam",
			Method:  "GET",
			Body:    nil,
			Query:   map[string]string{"queryParam": "queryParam"},
			Headers: nil,
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should pass thro validation error middleware for non-sent string param - missing query",
			ExpectedStatus:      422,
			ExpectedBodyContain: "A request was made to operation 'GetWithAllParamsRequiredPtr' but parameter 'queryParam' did not pass validation - Field 'queryParam' failed validation with tag 'required'",
			ExpendedHeaders: map[string]string{
				"X-pass-error-validation":        "true",
				"X-pass-after-succeed-operation": "",
			},
			Path:   "/e2e/get-with-all-params-required-ptr/pathParam",
			Method: "GET",
			Body:   nil,
			Query:  nil,
			Headers: map[string]string{
				"headerParam": "headerParam",
			},
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should pass thro validation error middleware for non-sent string param  - invalid value",
			ExpectedStatus:      422,
			ExpectedBodyContain: "Field 'headerParam' failed validation with tag 'validate_starts_with_letter'",
			ExpendedHeaders: map[string]string{
				"X-pass-error-validation":        "true",
				"X-pass-after-succeed-operation": "",
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
			Name:                "Should pass thro validation error middleware for string param - primitive number conversion",
			ExpectedStatus:      422,
			ExpectedBodyContain: "A request was made to operation 'TestPrimitiveConversions' but parameter 'value3' was not properly sent - Expected int but got string",
			ExpendedHeaders: map[string]string{
				"X-pass-error-validation":        "true",
				"X-pass-after-succeed-operation": "",
			},
			Path:    "/e2e/test-primitive-conversions",
			Method:  "POST",
			Query:   map[string]string{"value1": "60", "value2": "true", "value3": "10.6", "value4": "3"},
			Headers: map[string]string{},
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should pass thro validation error middleware for string param - primitive bool conversion",
			ExpectedStatus:      422,
			ExpectedBodyContain: "A request was made to operation 'TestPrimitiveConversions' but parameter 'value2' was not properly sent - Expected bool but got string",
			ExpendedHeaders: map[string]string{
				"X-pass-error-validation":        "true",
				"X-pass-after-succeed-operation": "",
			},
			Path:    "/e2e/test-primitive-conversions",
			Method:  "POST",
			Query:   map[string]string{"value1": "60", "value2": "true65", "value3": "10", "value4": "3"},
			Headers: map[string]string{},
		})
	})

	It("Should pass thro validation error middleware for struct param", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should pass thro validation error middleware for struct param - missing body",
			ExpectedStatus:      422,
			ExpectedBodyContain: "A request was made to operation 'PostWithAllParamsWithBody' but body parameter 'theBody' did not pass validation of 'BodyInfo' - body is required but was not provided",
			ExpendedHeaders: map[string]string{
				"X-pass-error-validation":        "true",
				"X-pass-after-succeed-operation": "",
			},
			Path:   "/e2e/post-with-all-params-body",
			Method: "POST",
			Body:   nil,
			Query:  map[string]string{"queryParam": "queryParam"},
			Headers: map[string]string{
				"headerParam": "headerParam",
			},
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should pass thro validation error middleware for struct param - invalid body",
			ExpectedStatus:      422,
			ExpectedBodyContain: "cannot unmarshal number into Go struct field BodyInfo.bodyParam of type string",
			ExpendedHeaders: map[string]string{
				"X-pass-error-validation":        "true",
				"X-pass-after-succeed-operation": "",
			},
			Path:   "/e2e/post-with-all-params-body-required-ptr",
			Method: "POST",
			Body:   assets.BodyInfo2{BodyParam: 1},
			Query:  map[string]string{"queryParam": "queryParam"},
			Headers: map[string]string{
				"headerParam": "headerParam",
			},
		})
	})

	It("Should abort by validation error middleware for string param", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should pass thro validation error middleware for non-sent string param",
			ExpectedStatus:      400,
			ExpectedBodyContain: "abort-on-error header is set to true ",
			ExpendedHeaders: map[string]string{
				"X-pass-error-validation":        "true",
				"X-pass-after-succeed-operation": "",
			},
			Path:    "/e2e/get-with-all-params-required-ptr/pathParam",
			Method:  "GET",
			Body:    nil,
			Query:   map[string]string{"queryParam": "queryParam"},
			Headers: map[string]string{"abort-on-error": "true"},
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should pass thro validation error middleware for non-sent string param - missing query",
			ExpectedStatus:      400,
			ExpectedBodyContain: "abort-on-error header is set to true ",
			ExpendedHeaders: map[string]string{
				"X-pass-error-validation":        "true",
				"X-pass-after-succeed-operation": "",
			},
			Path:   "/e2e/get-with-all-params-required-ptr/pathParam",
			Method: "GET",
			Body:   nil,
			Query:  nil,
			Headers: map[string]string{
				"headerParam":    "headerParam",
				"abort-on-error": "true",
			},
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should pass thro validation error middleware for non-sent string param  - invalid value",
			ExpectedStatus:      400,
			ExpectedBodyContain: "failed on the 'validate_starts_with_letter' tag",
			ExpendedHeaders: map[string]string{
				"X-pass-error-validation":        "true",
				"X-pass-after-succeed-operation": "",
			},
			Path:   "/e2e/get-header-start-with-letter",
			Method: "GET",
			Body:   nil,
			Query:  nil,
			Headers: map[string]string{
				"headerparam":    "1headerParam",
				"abort-on-error": "true",
			},
		})
	})

	It("Should abort by validation error middleware for struct param", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should pass thro validation error middleware for struct param - missing body",
			ExpectedStatus:      400,
			ExpectedBodyContain: "body is required but was not provided",
			ExpendedHeaders: map[string]string{
				"X-pass-error-validation":        "true",
				"X-pass-after-succeed-operation": "",
			},
			Path:   "/e2e/post-with-all-params-body",
			Method: "POST",
			Body:   nil,
			Query:  map[string]string{"queryParam": "queryParam"},
			Headers: map[string]string{
				"headerParam":    "headerParam",
				"abort-on-error": "true",
			},
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should pass thro validation error middleware for struct param - invalid body",
			ExpectedStatus:      400,
			ExpectedBodyContain: "cannot unmarshal number into Go struct field BodyInfo.bodyParam of type string",
			ExpendedHeaders: map[string]string{
				"X-pass-error-validation":        "true",
				"X-pass-after-succeed-operation": "",
			},
			Path:   "/e2e/post-with-all-params-body-required-ptr",
			Method: "POST",
			Body:   assets.BodyInfo2{BodyParam: 1},
			Query:  map[string]string{"queryParam": "queryParam"},
			Headers: map[string]string{
				"headerParam":    "headerParam",
				"abort-on-error": "true",
			},
		})
	})

	It("Should pass thro output validation error middleware", func() {
		RunRouterTest(common.RouterTest{
			Name:           "Should set validation error for invalid response payload - direct",
			ExpectedStatus: 500,
			ExpectedBody:   "{\"type\":\"Internal Server Error\",\"title\":\"\",\"detail\":\"Encountered an error during operation 'TestResponseValidation'\",\"status\":500,\"instance\":\"/gleece/controller/error/TestResponseValidation\",\"extensions\":{}}",
			ExpendedHeaders: map[string]string{
				"X-pass-output-validation":       "true",
				"X-pass-after-succeed-operation": "",
			},
			Path:    "/e2e/test-response-validation",
			Method:  "POST",
			Body:    nil,
			Query:   nil,
			Headers: nil,
		})

		RunRouterTest(common.RouterTest{
			Name:           "Should set validation error for invalid response payload -ptr",
			ExpectedStatus: 500,
			ExpectedBody:   "{\"type\":\"Internal Server Error\",\"title\":\"\",\"detail\":\"Encountered an error during operation 'TestResponseValidationPtr'\",\"status\":500,\"instance\":\"/gleece/controller/error/TestResponseValidationPtr\",\"extensions\":{}}",
			ExpendedHeaders: map[string]string{
				"X-pass-output-validation":       "true",
				"X-pass-after-succeed-operation": "",
			},
			Path:    "/e2e/test-response-validation-ptr",
			Method:  "POST",
			Body:    nil,
			Query:   nil,
			Headers: nil,
		})
	})

	It("Should abort by output validation error middleware", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should set validation error for invalid response payload - direct",
			ExpectedStatus:      400,
			ExpectedBodyContain: "Key: 'ResponseTest.Index' Error:Field validation for 'Index' failed on the 'gte' tag",
			ExpendedHeaders: map[string]string{
				"X-pass-output-validation":       "true",
				"X-pass-after-succeed-operation": "",
			},
			Path:   "/e2e/test-response-validation",
			Method: "POST",
			Body:   nil,
			Query:  nil,
			Headers: map[string]string{
				"abort-on-error":                 "true",
				"X-pass-after-succeed-operation": "true",
			},
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should set validation error for invalid response payload -ptr",
			ExpectedStatus:      400,
			ExpectedBodyContain: "Key: 'ResponseTest.Index' Error:Field validation for 'Index' failed on the 'gte' tag",
			ExpendedHeaders: map[string]string{
				"X-pass-output-validation":       "true",
				"X-pass-after-succeed-operation": "",
			},
			Path:   "/e2e/test-response-validation-ptr",
			Method: "POST",
			Body:   nil,
			Query:  nil,
			Headers: map[string]string{
				"abort-on-error": "true",
			},
		})
	})
})
