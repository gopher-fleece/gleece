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
})
