package e2e

import (
	"github.com/gopher-fleece/gleece/e2e/assets"
	"github.com/gopher-fleece/gleece/e2e/common"
	"github.com/haimkastner/unitsnet-go/units"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("E2E Routing Spec", func() {
	lf := units.LengthFactory{}
	length, _ := lf.FromMeters(1000)
	cm := units.LengthCentimeter
	dtoInCm := length.ToDto(&cm)

	It("Should return status code 200 for simple get", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 200 for simple get",
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

	It("Should use custom validator", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should use custom validator - valid header",
			ExpectedStatus:  200,
			ExpectedBody:    "\"headerParam\"",
			ExpendedHeaders: nil,
			Path:            "/e2e/get-header-start-with-letter",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers: map[string]string{
				"headerparam": "headerParam",
			},
			RunningMode: &fullyFeaturedRouting,
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should use custom validator - invalid header",
			ExpectedStatus:      422,
			ExpectedBodyContain: "Field 'headerParam' failed validation with tag 'validate_starts_with_letter'",
			ExpendedHeaders:     nil,
			Path:                "/e2e/get-header-start-with-letter",
			Method:              "GET",
			Body:                nil,
			Query:               nil,
			Headers: map[string]string{
				"headerparam": "1headerParam",
			},
			RunningMode: &fullyFeaturedRouting,
		})
	})

	It("Should return status code 200 for arrays in body and response", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 200 for arrays in body and response - in body",
			ExpectedStatus:  200,
			ExpectedBody:    "[{\"value\":100000,\"unit\":\"Centimeter\"}]",
			ExpendedHeaders: nil,
			Path:            "/e2e/arrays-in-body-and-res",
			Method:          "POST",
			Body:            []units.LengthDto{dtoInCm},
			Query:           map[string]string{},
			Headers:         map[string]string{},
			RunningMode:     &allRouting,
		})
	})

	It("Should return status code 200 for get with all params in use", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 200 for get with all params in use",
			ExpectedStatus:  200,
			ExpectedBody:    "\"pathParamqueryParamheaderParam\"",
			ExpendedHeaders: nil,
			Path:            "/e2e/get-with-all-params/pathParam",
			Method:          "GET",
			Body:            nil,
			Query:           map[string]string{"queryParam": "queryParam"},
			Headers: map[string]string{
				"headerParam": "headerParam",
			},
		})
	})

	It("Should return status code 200 for get with all params ptr", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 200 for get with all params ptr",
			ExpectedStatus:  200,
			ExpectedBody:    "\"pathParamqueryParamheaderParam\"",
			ExpendedHeaders: nil,
			Path:            "/e2e/get-with-all-params-ptr/pathParam",
			Method:          "GET",
			Body:            nil,
			Query:           map[string]string{"queryParam": "queryParam"},
			Headers: map[string]string{
				"headerParam": "headerParam",
			},
		})
	})

	It("Should return status code 200 for get with all params empty ptr", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 200 for get with all params empty ptr",
			ExpectedStatus:  200,
			ExpectedBody:    "\"pathParam\"",
			ExpendedHeaders: nil,
			Path:            "/e2e/get-with-all-params-ptr/pathParam",
			Method:          "GET",
			Body:            nil,
			Query:           nil,
			Headers:         nil,
		})
	})

	It("Should return status code 422 for get with all params empty ptr", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 422 for get with all params empty ptr - missing header",
			ExpectedStatus:      422,
			ExpectedBodyContain: "A request was made to operation 'GetWithAllParamsRequiredPtr' but parameter 'headerParam' did not pass validation - Field 'headerParam' failed validation with tag 'required'",
			ExpendedHeaders:     nil,
			Path:                "/e2e/get-with-all-params-required-ptr/pathParam",
			Method:              "GET",
			Body:                nil,
			Query:               map[string]string{"queryParam": "queryParam"},
			Headers:             nil,
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 422 for get with all params empty ptr - missing query",
			ExpectedStatus:      422,
			ExpectedBodyContain: "A request was made to operation 'GetWithAllParamsRequiredPtr' but parameter 'queryParam' did not pass validation - Field 'queryParam' failed validation with tag 'required'",
			ExpendedHeaders:     nil,
			Path:                "/e2e/get-with-all-params-required-ptr/pathParam",
			Method:              "GET",
			Body:                nil,
			Query:               nil,
			Headers: map[string]string{
				"headerParam": "headerParam",
			},
		})
	})

	It("Should return status code 200 for get with all params with body", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 200 for get with all params with body",
			ExpectedStatus:  200,
			ExpectedBody:    "{\"bodyParam\":\"queryParamheaderParamthebody\"}",
			ExpendedHeaders: nil,
			Path:            "/e2e/post-with-all-params-body",
			Method:          "POST",
			Body:            assets.BodyInfo{BodyParam: "thebody"},
			Query:           map[string]string{"queryParam": "queryParam"},
			Headers: map[string]string{
				"headerParam": "headerParam",
			},
		})
	})

	It("Should return status code 422 for missing body", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 422 for missing body",
			ExpectedStatus:      422,
			ExpectedBodyContain: "A request was made to operation 'PostWithAllParamsWithBody' but body parameter 'theBody' did not pass validation of 'BodyInfo' - body is required but was not provided",
			ExpendedHeaders:     nil,
			Path:                "/e2e/post-with-all-params-body",
			Method:              "POST",
			Body:                nil,
			Query:               map[string]string{"queryParam": "queryParam"},
			Headers: map[string]string{
				"headerParam": "headerParam",
			},
		})
	})

	It("Should return status code 200 for get with all params with body ptr", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 200 for get with all params with body ptr",
			ExpectedStatus:  200,
			ExpectedBody:    "{\"bodyParam\":\"queryParamheaderParamthebody\"}",
			ExpendedHeaders: nil,
			Path:            "/e2e/post-with-all-params-body-ptr",
			Method:          "POST",
			Body:            assets.BodyInfo{BodyParam: "thebody"},
			Query:           map[string]string{"queryParam": "queryParam"},
			Headers: map[string]string{
				"headerParam": "headerParam",
			},
		})

		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 200 for get with all params with body ptr - empty body",
			ExpectedStatus:  200,
			ExpectedBody:    "{\"bodyParam\":\"queryParamheaderParamempty\"}",
			ExpendedHeaders: nil,
			Path:            "/e2e/post-with-all-params-body-ptr",
			Method:          "POST",
			Body:            nil,
			Query:           map[string]string{"queryParam": "queryParam"},
			Headers: map[string]string{
				"headerParam": "headerParam",
			},
		})
	})

	It("Should return status code 200 for get with all params with body required ptr", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 200 for get with all params with body required ptr - missing body",
			ExpectedStatus:      422,
			ExpectedBodyContain: "A request was made to operation 'PostWithAllParamsWithBodyRequiredPtr' but body parameter 'theBody' did not pass validation of 'BodyInfo' - body is required but was not provided",
			ExpendedHeaders:     nil,
			Path:                "/e2e/post-with-all-params-body-required-ptr",
			Method:              "POST",
			Body:                nil,
			Query:               map[string]string{"queryParam": "queryParam"},
			Headers: map[string]string{
				"headerParam": "headerParam",
			},
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 200 for get with all params with body required ptr - empty body",
			ExpectedStatus:      422,
			ExpectedBodyContain: "A request was made to operation 'PostWithAllParamsWithBodyRequiredPtr' but body parameter 'theBody' did not pass validation of 'BodyInfo' - Field 'BodyParam' failed validation with tag 'required'.",
			ExpendedHeaders:     nil,
			Path:                "/e2e/post-with-all-params-body-required-ptr",
			Method:              "POST",
			Body:                assets.BodyInfo{},
			Query:               map[string]string{"queryParam": "queryParam"},
			Headers: map[string]string{
				"headerParam": "headerParam",
			},
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 200 for get with all params with body required ptr - invalid body",
			ExpectedStatus:      422,
			ExpectedBodyContain: "cannot unmarshal number into Go struct field BodyInfo.bodyParam of type string",
			ExpendedHeaders:     nil,
			Path:                "/e2e/post-with-all-params-body-required-ptr",
			Method:              "POST",
			Body:                assets.BodyInfo2{BodyParam: 1},
			Query:               map[string]string{"queryParam": "queryParam"},
			Headers: map[string]string{
				"headerParam": "headerParam",
			},
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 200 for get with all params with body required ptr - empty body with validation",
			ExpectedStatus:      422,
			ExpectedBodyContain: "Field 'BodyParam' failed validation with tag 'required'",
			ExpendedHeaders:     nil,
			Path:                "/e2e/post-with-all-params-body-required-ptr",
			Method:              "POST",
			Body:                assets.BodyInfo{},
			Query:               map[string]string{"queryParam": "queryParam"},
			Headers: map[string]string{
				"headerParam": "headerParam",
			},
		})
	})

	It("Should handle GET request", func() {
		RunRouterTest(common.RouterTest{
			Name:           "Should handle GET request",
			ExpectedStatus: 204,
			ExpectedBody:   "",
			Headers:        nil,
			Path:           "/e2e/http-method",
			Method:         "GET",
			Body:           nil,
			Query:          nil,
			ExpendedHeaders: map[string]string{
				"x-method": "get",
			},
		})
	})

	It("Should handle POST request", func() {
		RunRouterTest(common.RouterTest{
			Name:           "Should handle POST request",
			ExpectedStatus: 204,
			ExpectedBody:   "",
			Headers:        nil,
			Path:           "/e2e/http-method",
			Method:         "POST",
			Body:           nil,
			Query:          nil,
			ExpendedHeaders: map[string]string{
				"x-method": "post",
			},
		})
	})

	It("Should handle PUT request", func() {
		RunRouterTest(common.RouterTest{
			Name:           "Should handle PUT request",
			ExpectedStatus: 204,
			ExpectedBody:   "",
			Headers:        nil,
			Path:           "/e2e/http-method",
			Method:         "PUT",
			Body:           nil,
			Query:          nil,
			ExpendedHeaders: map[string]string{
				"x-method": "put",
			},
		})
	})

	It("Should handle DELETE request", func() {
		RunRouterTest(common.RouterTest{
			Name:           "Should handle DELETE request",
			ExpectedStatus: 204,
			ExpectedBody:   "",
			Headers:        nil,
			Path:           "/e2e/http-method",
			Method:         "DELETE",
			Body:           nil,
			Query:          nil,
			ExpendedHeaders: map[string]string{
				"x-method": "delete",
			},
		})
	})

	It("Should handle PATCH request", func() {
		RunRouterTest(common.RouterTest{
			Name:           "Should handle PATCH request",
			ExpectedStatus: 204,
			ExpectedBody:   "",
			Headers:        nil,
			Path:           "/e2e/http-method",
			Method:         "PATCH",
			Body:           nil,
			Query:          nil,
			ExpendedHeaders: map[string]string{
				"x-method": "patch",
			},
		})
	})

	It("Should return status code 200 for form parameters", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 200 for form parameters",
			ExpectedStatus:      200,
			ExpectedBodyContain: "form1form2",
			ExpendedHeaders:     nil,
			Path:                "/e2e/form",
			Method:              "POST",
			Form:                map[string]string{"item1": "form1", "item2": "form2"},
			Headers:             map[string]string{},
		})
	})

	It("Should return status code 422 for missing form parameters", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 200 for form parameters",
			ExpectedStatus:      422,
			ExpectedBodyContain: "",
			ExpendedHeaders:     nil,
			Path:                "/e2e/form",
			Method:              "POST",
			Form:                map[string]string{"item1": "form1"},
			Headers:             map[string]string{},
		})
	})

	It("Should return status code 200 for empty form parameters", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 200 for empty form parameters",
			ExpectedStatus:      200,
			ExpectedBodyContain: "form1",
			ExpendedHeaders:     nil,
			Path:                "/e2e/form",
			Method:              "POST",
			Form:                map[string]string{"item1": "form1", "item2": ""},
			Headers:             map[string]string{},
		})
	})

	It("Should return status code 200 for valid form extra", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 200 for valid form extra",
			ExpectedStatus:      200,
			ExpectedBodyContain: "99|second|100",
			ExpendedHeaders:     nil,
			Path:                "/e2e/form-extra",
			Method:              "POST",
			Form:                map[string]string{"item1": "99", "item2": "second"},
			Headers:             map[string]string{},
			Query:               map[string]string{"item3": "100"},
		})
	})

	It("Should return status code 422 for not matching form extra validation tag", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 422 for not matching form extra validation tag",
			ExpectedStatus:      422,
			ExpectedBodyContain: "Field 'item1' failed validation with tag 'gte'",
			ExpendedHeaders:     nil,
			Path:                "/e2e/form-extra",
			Method:              "POST",
			Form:                map[string]string{"item1": "70", "item2": "second"},
			Headers:             map[string]string{},
			Query:               map[string]string{"item3": "100"},
		})
	})

	It("Should return status code 422 for form required tag", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 422 for form required tag",
			ExpectedStatus:      422,
			ExpectedBodyContain: "Field 'item2' failed validation with tag 'required'",
			ExpendedHeaders:     nil,
			Path:                "/e2e/form-extra",
			Method:              "POST",
			Form:                map[string]string{"item1": "99"},
			Headers:             map[string]string{},
			Query:               map[string]string{"item3": "100"},
		})
	})

	It("Should return status code 200 for primitive parameters", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 200 for primitive parameters",
			ExpectedStatus:      200,
			ExpectedBodyContain: "60 true 10 3.14",
			ExpendedHeaders:     nil,
			Path:                "/e2e/test-primitive-conversions",
			Method:              "POST",
			Query:               map[string]string{"value1": "60", "value2": "true", "value3": "10", "value4": "3.14"},
			Headers:             map[string]string{},
		})
	})

	It("Should return status code 422 for primitive wrong parameters", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 200 for primitive parameters",
			ExpectedStatus:      422,
			ExpectedBodyContain: "A request was made to operation 'TestPrimitiveConversions' but parameter 'value1' was not properly sent - Expected int64 but got string",
			ExpendedHeaders:     nil,
			Path:                "/e2e/test-primitive-conversions",
			Method:              "POST",
			Query:               map[string]string{"value1": "sixty", "value2": "true", "value3": "10", "value4": "3.14"},
			Headers:             map[string]string{},
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 200 for primitive parameters",
			ExpectedStatus:      422,
			ExpectedBodyContain: "A request was made to operation 'TestPrimitiveConversions' but parameter 'value2' was not properly sent - Expected bool but got string",
			ExpendedHeaders:     nil,
			Path:                "/e2e/test-primitive-conversions",
			Method:              "POST",
			Query:               map[string]string{"value1": "60", "value2": "p1", "value3": "10", "value4": "3.14"},
			Headers:             map[string]string{},
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 200 for primitive parameters",
			ExpectedStatus:      422,
			ExpectedBodyContain: "A request was made to operation 'TestPrimitiveConversions' but parameter 'value3' was not properly sent - Expected int but got string",
			ExpendedHeaders:     nil,
			Path:                "/e2e/test-primitive-conversions",
			Method:              "POST",
			Query:               map[string]string{"value1": "60", "value2": "true", "value3": "10.6", "value4": "3"},
			Headers:             map[string]string{},
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 200 for primitive parameters",
			ExpectedStatus:      422,
			ExpectedBodyContain: "A request was made to operation 'TestPrimitiveConversions' but parameter 'value4' was not properly sent - Expected float64 but got string",
			ExpendedHeaders:     nil,
			Path:                "/e2e/test-primitive-conversions",
			Method:              "POST",
			Query:               map[string]string{"value1": "60", "value2": "true", "value3": "10", "value4": "3true"},
			Headers:             map[string]string{},
		})
	})

})
