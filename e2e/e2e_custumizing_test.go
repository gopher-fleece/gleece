package e2e

import (
	"github.com/gopher-fleece/gleece/e2e/common"
	. "github.com/onsi/ginkgo/v2"
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
			Name:           "Should set custom header by template extension",
			ExpectedStatus: 200,
			ExpectedBody:   "\"works\"",
			ExpendedHeaders: map[string]string{
				"x-test-header": "test",
				"x-extended":    "SimpleGet",
			},
			Path:    "/e2e/simple-get",
			Method:  "GET",
			Body:    nil,
			Query:   nil,
			Headers: nil,
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
})
