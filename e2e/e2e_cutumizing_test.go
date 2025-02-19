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
				"x-inject":      "true",
				"x-extended":    "SimpleGet",
			},
			Path:    "/e2e/simple-get",
			Method:  "GET",
			Body:    nil,
			Query:   nil,
			Headers: nil,
		})
	})
})
