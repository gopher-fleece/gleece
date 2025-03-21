package e2e

import (
	"github.com/gopher-fleece/gleece/e2e/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("E2E Errors Spec", func() {

	It("Should return default rfc7807 error", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should return default rfc7807 error",
			ExpectedStatus:  500,
			ExpectedBody:    "{\"type\":\"Internal Server Error\",\"title\":\"\",\"detail\":\"Encountered an error during operation 'DefaultError'\",\"status\":500,\"instance\":\"/gleece/controller/error/DefaultError\",\"extensions\":{\"error\":\"default error\"}}",
			ExpendedHeaders: nil,
			Path:            "/e2e/default-error",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
		})
	})

	It("Should return default error rfc7807 with payload response", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should return default error rfc7807 with payload response",
			ExpectedStatus:  500,
			ExpectedBody:    "{\"type\":\"Internal Server Error\",\"title\":\"\",\"detail\":\"Encountered an error during operation 'DefaultErrorWithPayload'\",\"status\":500,\"instance\":\"/gleece/controller/error/DefaultErrorWithPayload\",\"extensions\":{\"error\":\"default error\"}}",
			ExpendedHeaders: nil,
			Path:            "/e2e/default-error-with-payload",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
		})
	})

	It("Should return custom error", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should return custom error",
			ExpectedStatus:  500,
			ExpectedBody:    "{\"message\":\"custom error\"}",
			ExpendedHeaders: nil,
			Path:            "/e2e/custom-error",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
		})
	})

	It("Should set custom error code", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should set custom error code",
			ExpectedStatus:  503,
			ExpectedBody:    "{\"type\":\"Service Unavailable\",\"title\":\"\",\"detail\":\"Encountered an error during operation 'Error503'\",\"status\":503,\"instance\":\"/gleece/controller/error/Error503\",\"extensions\":{\"error\":\"default error\"}}",
			ExpendedHeaders: nil,
			Path:            "/e2e/503-error-code",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
		})
	})

	It("Should set custom error code with custom error", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should set custom error code with custom error",
			ExpectedStatus:  503,
			ExpectedBody:    "{\"message\":\"custom error\"}",
			ExpendedHeaders: nil,
			Path:            "/e2e/custom-error-503",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
		})
	})

	It("Should set validation error for invalid response payload", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should set validation error for invalid response payload",
			ExpectedStatus:  500,
			ExpectedBody:    "{\"type\":\"Internal Server Error\",\"title\":\"\",\"detail\":\"Encountered an error during operation 'TestResponseValidation'\",\"status\":500,\"instance\":\"/gleece/controller/error/TestResponseValidation\",\"extensions\":{}}",
			ExpendedHeaders: nil,
			Path:            "/e2e/test-response-validation",
			Method:          "POST",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
			RunningMode:     &fullyFeaturedRouting,
		})
	})

	It("Should NOT set validation error for invalid response payload", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should NOT set validation error for invalid response payload",
			ExpectedStatus:  200,
			ExpectedBody:    "{\"success\":\"success\",\"index\":-1}",
			ExpendedHeaders: nil,
			Path:            "/e2e/test-response-validation",
			Method:          "POST",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
			RunningMode:     &exExtraRouting,
		})
	})

	It("Should set validation error for invalid response payload pointer", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should set validation error for invalid response payload pointer",
			ExpectedStatus:  500,
			ExpectedBody:    "{\"type\":\"Internal Server Error\",\"title\":\"\",\"detail\":\"Encountered an error during operation 'TestResponseValidationPtr'\",\"status\":500,\"instance\":\"/gleece/controller/error/TestResponseValidationPtr\",\"extensions\":{}}",
			ExpendedHeaders: nil,
			Path:            "/e2e/test-response-validation-ptr",
			Method:          "POST",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
			RunningMode:     &fullyFeaturedRouting,
		})
	})

	It("Should NOT set validation error for invalid response payload pointer", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should NOT set validation error for invalid response payload pointer",
			ExpectedStatus:  200,
			ExpectedBody:    "",
			ExpendedHeaders: nil,
			Path:            "/e2e/test-response-validation-ptr",
			Method:          "POST",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
			RunningMode:     &exExtraRouting,
		})
	})

	It("Should set validation error for empty response payload pointer", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should set validation error for empty response payload pointer",
			ExpectedStatus:  500,
			ExpectedBody:    "{\"type\":\"Internal Server Error\",\"title\":\"\",\"detail\":\"Encountered an error during operation 'TestResponseValidationNull'\",\"status\":500,\"instance\":\"/gleece/controller/error/TestResponseValidationNull\",\"extensions\":{}}",
			ExpendedHeaders: nil,
			Path:            "/e2e/test-response-validation-null",
			Method:          "POST",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
			RunningMode:     &fullyFeaturedRouting,
		})
	})

	It("Should NOT set validation error for empty response payload pointer", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should NOT set validation error for empty response payload pointer",
			ExpectedStatus:  200,
			ExpectedBody:    "",
			ExpendedHeaders: nil,
			Path:            "/e2e/test-response-validation-null",
			Method:          "POST",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
			RunningMode:     &exExtraRouting,
		})
	})
})
