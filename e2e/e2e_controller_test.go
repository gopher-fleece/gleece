package e2e

import (
	"github.com/gopher-fleece/gleece/e2e/assets"
	"github.com/gopher-fleece/gleece/e2e/common"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("E2E Controller Spec", func() {

	It("Should return status code 200 for response with payload", func() {

		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 200 for response with payload",
			ExpectedStatus:      200,
			ExpectedBodyContain: "\"works\"",
			ExpendedHeaders:     map[string]string{"Content-Type": "application/json"},
			Path:                "/e2e/simple-get",
			Method:              "GET",
			Body:                nil,
			Query:               nil,
			Headers:             nil,
		})
	})

	It("Should return status code 204 for with payload get", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 204 for with payload get",
			ExpectedStatus:  204,
			ExpectedBody:    "",
			ExpendedHeaders: map[string]string{"Content-Type": ""},
			Path:            "/e2e/simple-get-empty",
			Method:          "GET",
			Body:            nil,
			Query:           map[string]string{"queryParam": "queryParam"},
			Headers:         nil,
		})
	})

	It("Should return status code 204 for explicit set status", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 204 for explicit set status",
			ExpectedStatus:      204,
			ExpectedBodyContain: "",
			ExpendedHeaders:     nil,
			Path:                "/e2e/get-with-all-params/pathParam",
			Method:              "GET",
			Body:                nil,
			Query:               map[string]string{"queryParam": "204"},
			Headers:             map[string]string{"headerParam": "headerParam"},
		})
	})

	It("Should return status code 200 for explicit set status", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 200 for explicit set status",
			ExpectedStatus:  200,
			ExpectedBody:    "",
			ExpendedHeaders: nil,
			Path:            "/e2e/simple-get-empty",
			Method:          "GET",
			Body:            nil,
			Query:           map[string]string{"queryParam": "200"},
			Headers:         nil,
		})
	})

	It("Should set custom header", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should set custom header",
			ExpectedStatus:  200,
			ExpectedBody:    "\"works\"",
			ExpendedHeaders: map[string]string{"X-Test-Header": "test"},
			Path:            "/e2e/simple-get",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
		})
	})

	It("Should response payload with string", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should response payload with string",
			ExpectedStatus:  200,
			ExpectedBody:    "\"works\"",
			ExpendedHeaders: nil,
			Path:            "/e2e/simple-get",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
		})
	})

	It("Should response payload with empty string", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should response payload with empty string",
			ExpectedStatus:  200,
			ExpectedBody:    "\"\"",
			ExpendedHeaders: nil,
			Path:            "/e2e/simple-get-empty-string",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
		})
	})

	It("Should response payload string pointer payload", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should response payload string pointer payload",
			ExpectedStatus:  200,
			ExpectedBody:    "\"ptr\"",
			ExpendedHeaders: nil,
			Path:            "/e2e/simple-get-ptr-string",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
		})
	})

	It("Should response payload with null string", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should response payload with null string",
			ExpectedStatus:  200,
			ExpectedBody:    "null",
			ExpendedHeaders: nil,
			Path:            "/e2e/simple-get-null-string",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
		})
	})

	It("Should response object payload", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should response object payload",
			ExpectedStatus:  200,
			ExpectedBody:    "{\"data\":\"BodyResponse\"}",
			ExpendedHeaders: nil,
			Path:            "/e2e/simple-get-object",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
		})
	})

	It("Should response object pointer payload", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should response object pointer payload",
			ExpectedStatus:  200,
			ExpectedBody:    "{\"data\":\"BodyResponse\"}",
			ExpendedHeaders: nil,
			Path:            "/e2e/simple-get-object",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
		})
	})

	It("Should response object pointer payload", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should response object pointer payload",
			ExpectedStatus:  200,
			ExpectedBody:    "{\"data\":\"BodyResponse\"}",
			ExpendedHeaders: nil,
			Path:            "/e2e/simple-get-object-ptr",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
			RunningMode:     &allRouting,
		})
	})

	It("Should response non-string primitive", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should response non-string primitive - int",
			ExpectedStatus:  200,
			ExpectedBody:    "20",
			ExpendedHeaders: nil,
			Path:            "/e2e/primitive-return-type",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
			RunningMode:     &allRouting,
		})
	})

	It("Should response slice of primitive", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should response non-string primitive - int array",
			ExpectedStatus:  200,
			ExpectedBody:    "[20]",
			ExpendedHeaders: nil,
			Path:            "/e2e/primitive-array-return-type",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
			RunningMode:     &allRouting,
		})
	})

	It("Should response non-string alias primitive", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should response non-string alias primitive - int",
			ExpectedStatus:  200,
			ExpectedBody:    "1",
			ExpendedHeaders: nil,
			Path:            "/e2e/primitive-alias-return-type",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
			RunningMode:     &allRouting,
		})
	})

	It("Should response slice of alias primitive", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should response slice of alias primitive- int array",
			ExpectedStatus:  200,
			ExpectedBody:    "[1]",
			ExpendedHeaders: nil,
			Path:            "/e2e/primitive-alias-array-return-type",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
			RunningMode:     &allRouting,
		})
	})

	It("Should response object null payload with null", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should response object null payload with null",
			ExpectedStatus:  200,
			ExpectedBody:    "null",
			ExpendedHeaders: nil,
			Path:            "/e2e/simple-get-object-null",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         map[string]string{"x-return-null": "true"}, // Bypass the error of output validation using the middleware
		})
	})

	It("Should allow custom context pass", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should allow custom context pass",
			ExpectedStatus:  204,
			ExpectedBody:    "",
			ExpendedHeaders: map[string]string{"X-Context-Pass": "true"},
			Path:            "/e2e/context-access",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
		})
	})

	It("Should handle context from auth function", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should allow context with no other input",
			ExpectedStatus:  204,
			ExpectedBody:    "",
			ExpendedHeaders: map[string]string{"x-context-auth": "123"},
			Path:            "/e2e/context-injection-empty",
			Method:          "POST",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
		})

		RunRouterTest(common.RouterTest{
			Name:            "Should allow context with other input",
			ExpectedStatus:  204,
			ExpectedBody:    "",
			ExpendedHeaders: map[string]string{"x-context-auth": "123"},
			Path:            "/e2e/context-injection",
			Method:          "POST",
			Body: assets.TheModel{
				ModelField: "model field",
				FirstLevelModel: assets.FirstLevelModel{
					FirstLevelModelField: "first level",
					SecondLevelModel: assets.SecondLevelModel{
						SecondLevelModelField: "second level",
					},
				},
				OtherModel: assets.OtherModel{
					OtherModelField: "other model",
				},
			},
			Query:   map[string]string{},
			Headers: map[string]string{},
		})
	})
})
