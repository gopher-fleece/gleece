package e2e

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("E2E Middlewares Spec", func() {

	It("Should pass succeeded middlewares", func() {
		RunRouterTest(RouterTest{
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
		RunRouterTest(RouterTest{
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
		RunRouterTest(RouterTest{
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
		RunRouterTest(RouterTest{
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
		RunRouterTest(RouterTest{
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
		RunRouterTest(RouterTest{
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
		RunRouterTest(RouterTest{
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
		RunRouterTest(RouterTest{
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
		RunRouterTest(RouterTest{
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
		RunRouterTest(RouterTest{
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
})
