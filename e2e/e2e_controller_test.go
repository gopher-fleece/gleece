package e2e

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("E2E Controller Spec", func() {

	It("Should return status code 200 for response with payload", func() {

		RunRouterTest(RouterTest{
			Name:            "Should return status code 200 for response with payload",
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

	It("Should return status code 204 for with payload get", func() {
		RunRouterTest(RouterTest{
			Name:            "Should return status code 204 for with payload get",
			ExpectedStatus:  204,
			ExpectedBody:    "",
			ExpendedHeaders: nil,
			Path:            "/e2e/simple-get-empty",
			Method:          "GET",
			Body:            nil,
			Query:           map[string]string{"queryParam": "queryParam"},
			Headers:         nil,
		})
	})

	It("Should return status code 204 for explicit set status", func() {
		RunRouterTest(RouterTest{
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
		RunRouterTest(RouterTest{
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
		RunRouterTest(RouterTest{
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
		RunRouterTest(RouterTest{
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
		RunRouterTest(RouterTest{
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
		RunRouterTest(RouterTest{
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
		RunRouterTest(RouterTest{
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
		RunRouterTest(RouterTest{
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
		RunRouterTest(RouterTest{
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
		RunRouterTest(RouterTest{
			Name:            "Should response object pointer payload",
			ExpectedStatus:  200,
			ExpectedBody:    "{\"data\":\"BodyResponse\"}",
			ExpendedHeaders: nil,
			Path:            "/e2e/simple-get-object-ptr",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
		})
	})

	It("Should response object null payload with null", func() {
		RunRouterTest(RouterTest{
			Name:            "Should response object null payload with null",
			ExpectedStatus:  200,
			ExpectedBody:    "null",
			ExpendedHeaders: nil,
			Path:            "/e2e/simple-get-object-null",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
		})
	})

	It("Should allow custom context pass", func() {
		RunRouterTest(RouterTest{
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
})
