package e2e

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("E2E Authorization Spec", func() {

	It("Should as default use class' security", func() {
		RunRouterTest(RouterTest{
			Name:                "Should as default use class' security",
			ExpectedStatus:      200,
			ExpectedBodyContain: "securitySchemaNameclass",
			ExpendedHeaders:     nil,
			Path:                "/e2e/with-default-class-security",
			Method:              "GET",
			Body:                nil,
			Query:               nil,
			Headers:             nil,
		})
	})

	It("Should override default class' security", func() {
		RunRouterTest(RouterTest{
			Name:                "Should override default class' security",
			ExpectedStatus:      200,
			ExpectedBodyContain: "securitySchemaNamemethod",
			ExpendedHeaders:     nil,
			Path:                "/e2e/with-default-override-class-security",
			Method:              "GET",
			Body:                nil,
			Query:               nil,
			Headers:             nil,
		})
	})

	It("Should as default use config security", func() {
		RunRouterTest(RouterTest{
			Name:                "Should as default use config security",
			ExpectedStatus:      200,
			ExpectedBodyContain: "securitySchemaName2config",
			ExpendedHeaders:     nil,
			Path:                "/e2e/with-default-config-security",
			Method:              "GET",
			Body:                nil,
			Query:               nil,
			Headers:             nil,
		})
	})

	It("Should override default config security", func() {
		RunRouterTest(RouterTest{
			Name:                "Should override default config security",
			ExpectedStatus:      200,
			ExpectedBodyContain: "securitySchemaNameother",
			ExpendedHeaders:     nil,
			Path:                "/e2e/with-one-security",
			Method:              "GET",
			Body:                nil,
			Query:               nil,
			Headers:             nil,
		})
	})

	It("Should fail one security check", func() {
		RunRouterTest(RouterTest{
			Name:            "Should fail one security check",
			ExpectedStatus:  401,
			ExpectedBody:    "{\"type\":\"Unauthorized\",\"title\":\"\",\"detail\":\"Failed to authorize\",\"status\":401,\"instance\":\"/gleece/authorization/error/WithOneSecurity\",\"extensions\":null}",
			ExpendedHeaders: nil,
			Path:            "/e2e/with-one-security",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers: map[string]string{
				"fail-auth": "securitySchemaName",
			},
		})
	})

	It("Should not fail on other security check", func() {
		RunRouterTest(RouterTest{
			Name:                "Should not fail on other security check",
			ExpectedStatus:      200,
			ExpectedBodyContain: "securitySchemaNameother",
			ExpendedHeaders:     nil,
			Path:                "/e2e/with-one-security",
			Method:              "GET",
			Body:                nil,
			Query:               nil,
			Headers: map[string]string{
				"fail-auth": "securitySchemaName2",
			},
		})
	})

	It("Should not fail by only first security check", func() {
		RunRouterTest(RouterTest{
			Name:                "Should not fail by only first security check",
			ExpectedStatus:      200,
			ExpectedBodyContain: "securitySchemaName2write",
			ExpendedHeaders:     nil,
			Path:                "/e2e/with-two-security",
			Method:              "GET",
			Body:                nil,
			Query:               nil,
			Headers: map[string]string{
				"fail-auth": "securitySchemaName",
			},
		})
	})

	It("Should not fail by only second security check", func() {
		RunRouterTest(RouterTest{
			Name:                "Should not fail by only second security check",
			ExpectedStatus:      200,
			ExpectedBodyContain: "securitySchemaNameother",
			ExpendedHeaders:     nil,
			Path:                "/e2e/with-two-security",
			Method:              "GET",
			Body:                nil,
			Query:               nil,
			Headers: map[string]string{
				"fail-auth": "securitySchemaName2",
			},
		})
	})

	It("Should fail by both security check", func() {
		RunRouterTest(RouterTest{
			Name:            "Should fail by both security check",
			ExpectedStatus:  401,
			ExpectedBody:    "{\"type\":\"Unauthorized\",\"title\":\"\",\"detail\":\"Failed to authorize\",\"status\":401,\"instance\":\"/gleece/authorization/error/WithTwoSecuritySameMethod\",\"extensions\":null}",
			ExpendedHeaders: nil,
			Path:            "/e2e/with-two-security-same-method",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers: map[string]string{
				"fail-auth": "securitySchemaName",
			},
		})
	})

	It("Should pass if both security passes check", func() {
		RunRouterTest(RouterTest{
			Name:                "Should pass if both security passes check",
			ExpectedStatus:      200,
			ExpectedBodyContain: "securitySchemaNameother",
			ExpendedHeaders:     nil,
			Path:                "/e2e/with-two-security-same-method",
			Method:              "GET",
			Body:                nil,
			Query:               nil,
			Headers:             nil,
		})
	})

	It("Should allow set custom auth code", func() {
		RunRouterTest(RouterTest{
			Name:                "Should allow set custom auth code",
			ExpectedStatus:      403,
			ExpectedBodyContain: "Failed to authorize\",\"status\":403",
			ExpendedHeaders:     nil,
			Path:                "/e2e/with-one-security",
			Method:              "GET",
			Body:                nil,
			Query:               nil,
			Headers: map[string]string{
				"fail-auth": "securitySchemaName",
				"fail-code": "403",
			},
		})
	})

	It("Should allow set custom auth code and error", func() {
		RunRouterTest(RouterTest{
			Name:            "Should allow set custom auth code and error",
			ExpectedStatus:  403,
			ExpectedBody:    "{\"message\":\"Custom error message\",\"description\":\"Custom error description\"}",
			ExpendedHeaders: nil,
			Path:            "/e2e/with-one-security",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers: map[string]string{
				"fail-auth-custom": "securitySchemaName",
				"fail-code":        "403",
			},
		})
	})
})
