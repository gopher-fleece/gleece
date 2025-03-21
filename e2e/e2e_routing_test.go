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
	km := units.LengthKilometer
	m := units.LengthMeter
	dtoInCm := length.ToDto(&cm)
	dtoInBroken := length.ToDto(&cm)
	dtoInBroken.Unit = "KiloBoom" // Set invalid value
	dtoInKMJson, _ := length.ToDtoJSON(&km)
	dtoInMJson, _ := length.ToDtoJSON(&m)

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

	It("Should return status code 200 for external packages models", func() {

		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 200 for external packages models - with all options",
			ExpectedStatus:  200,
			ExpectedBody:    string(dtoInKMJson),
			ExpendedHeaders: nil,
			Path:            "/e2e/external-packages",
			Method:          "POST",
			Body:            dtoInCm,
			Query:           map[string]string{"unit": "Kilometer"},
			Headers:         map[string]string{},
			RunningMode:     &allRouting,
		})

		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 200 for external packages models - with default options",
			ExpectedStatus:  200,
			ExpectedBody:    string(dtoInMJson),
			ExpendedHeaders: nil,
			Path:            "/e2e/external-packages",
			Method:          "POST",
			Body:            dtoInCm,
			Query:           map[string]string{},
			Headers:         map[string]string{},
			RunningMode:     &allRouting,
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

	It("Should return status code 200 for arrays inside body and response", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 200 for arrays in body and response - in body",
			ExpectedStatus:  200,
			ExpectedBody:    "[{\"listOfLength\":[{\"value\":100000,\"unit\":\"Centimeter\"}]}]",
			ExpendedHeaders: nil,
			Path:            "/e2e/arrays-inside-body-and-res",
			Method:          "POST",
			Body: []assets.BlaBla{{
				ListOfLength: []units.LengthDto{dtoInCm},
			}},
			Query:       map[string]string{},
			Headers:     map[string]string{},
			RunningMode: &allRouting,
		})
	})

	It("Should return status code 200 for deep arrays", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 200 for arrays in body and response - in body",
			ExpectedStatus:  200,
			ExpectedBody:    "[[[{\"value\":10}]]]",
			ExpendedHeaders: nil,
			Path:            "/e2e/deep-arrays-with-validation",
			Method:          "POST",
			Body: [][]assets.BlaBla2{{{
				Value: 10,
			}}},
			Query:       map[string]string{},
			Headers:     map[string]string{},
			RunningMode: &allRouting,
		})
	})

	It("Should return status code 422 for deep arrays", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 200 for arrays in body and response - empty body",
			ExpectedStatus:  422,
			ExpectedBody:    "",
			ExpendedHeaders: nil,
			Path:            "/e2e/deep-arrays-with-validation",
			Method:          "POST",
			Query:           map[string]string{},
			Headers:         map[string]string{},
			RunningMode:     &allRouting,
		})

		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 200 for arrays in body and response - deeper from needed",
			ExpectedStatus:  422,
			ExpectedBody:    "",
			ExpendedHeaders: nil,
			Path:            "/e2e/deep-arrays-with-validation",
			Method:          "POST",
			Body: [][][]assets.BlaBla2{{{{
				Value: 10,
			}}}},
			Query:       map[string]string{},
			Headers:     map[string]string{},
			RunningMode: &allRouting,
		})

		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 200 for arrays in body and response - not deeper as needed",
			ExpectedStatus:  422,
			ExpectedBody:    "",
			ExpendedHeaders: nil,
			Path:            "/e2e/deep-arrays-with-validation",
			Method:          "POST",
			Body: []assets.BlaBla2{{
				Value: 10,
			}},
			Query:       map[string]string{},
			Headers:     map[string]string{},
			RunningMode: &allRouting,
		})

		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 200 for arrays in body and response - internal validation not passed",
			ExpectedStatus:  422,
			ExpectedBody:    "",
			ExpendedHeaders: nil,
			Path:            "/e2e/deep-arrays-with-validation",
			Method:          "POST",
			Body: [][]assets.BlaBla2{{{
				Value: -1,
			}}},
			Query:       map[string]string{},
			Headers:     map[string]string{},
			RunningMode: &allRouting,
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

	It("Should return status code 200 for enums", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 200 for enum - queries & body",
			ExpectedStatus:  200,
			ExpectedBody:    "{\"value\":\"active 2\",\"values\":[\"active\",\"2\"],\"status\":\"active\",\"statuses\":[\"inactive\"]}",
			ExpendedHeaders: nil,
			Path:            "/e2e/test-enums",
			Method:          "POST",
			Query:           map[string]string{"value1": "active", "value2": "2"},
			Body: assets.ObjectWithEnum{
				Value:    "some value",
				Values:   []string{"some", "values"},
				Status:   assets.StatusEnumerationActive,
				Statuses: []assets.StatusEnumeration{assets.StatusEnumerationInactive},
			},
			Headers:     map[string]string{},
			RunningMode: &fullyFeaturedRouting,
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 200 for enum - header, path & forms",
			ExpectedStatus:      200,
			ExpectedBodyContain: "active 1 inactive",
			ExpendedHeaders:     nil,
			Path:                "/e2e/test-enums-in-all/active",
			Method:              "POST",
			Form:                map[string]string{"value3": "inactive"},
			Headers:             map[string]string{"value2": "1"},
			RunningMode:         &fullyFeaturedRouting,
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 200 for enum - optional param sent",
			ExpectedStatus:      200,
			ExpectedBodyContain: "active",
			ExpendedHeaders:     nil,
			Path:                "/e2e/test-enums-optional",
			Method:              "POST",
			Form:                map[string]string{},
			Headers:             map[string]string{"value1": "active"},
			RunningMode:         &fullyFeaturedRouting,
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 200 for enum - optional param not sent",
			ExpectedStatus:      200,
			ExpectedBodyContain: "nil",
			ExpendedHeaders:     nil,
			Path:                "/e2e/test-enums-optional",
			Method:              "POST",
			Form:                map[string]string{},
			Headers:             map[string]string{},
			RunningMode:         &fullyFeaturedRouting,
		})
	})

	It("Should NOT return status code 200 for enums", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should NOT return status code 200 for enum - header, path & forms",
			ExpectedStatus:      200,
			ExpectedBodyContain: "active 1 inactive",
			ExpendedHeaders:     nil,
			Path:                "/e2e/test-enums-in-all/active",
			Method:              "POST",
			Form:                map[string]string{"value3": "inactive"},
			Headers:             map[string]string{"value2": "1"},
			RunningMode:         &exExtraRouting,
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should NOT return status code 200 for enum - optional param sent",
			ExpectedStatus:      200,
			ExpectedBodyContain: "active",
			ExpendedHeaders:     nil,
			Path:                "/e2e/test-enums-optional",
			Method:              "POST",
			Form:                map[string]string{},
			Headers:             map[string]string{"value1": "active"},
			RunningMode:         &exExtraRouting,
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should NOT return status code 200 for enum - optional param not sent",
			ExpectedStatus:      200,
			ExpectedBodyContain: "nil",
			ExpendedHeaders:     nil,
			Path:                "/e2e/test-enums-optional",
			Method:              "POST",
			Form:                map[string]string{},
			Headers:             map[string]string{},
			RunningMode:         &exExtraRouting,
		})
	})

	It("Should return status code 422 for wrong enum prams and body", func() {
		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 422 for enum prams and body - wrong direct string enum",
			ExpectedStatus:  422,
			ExpectedBody:    "{\"type\":\"Unprocessable Entity\",\"title\":\"\",\"detail\":\"A request was made to operation 'TestEnums' but parameter 'value1' was not properly sent - Expected StatusEnumeration but got string\",\"status\":422,\"instance\":\"/gleece/validation/error/TestEnums\",\"extensions\":{\"error\":\"value1 must be one of \\\"active, inactive\\\" options only but got blabla\"}}",
			ExpendedHeaders: nil,
			Path:            "/e2e/test-enums",
			Method:          "POST",
			Query:           map[string]string{"value1": "blabla", "value2": "2"},
			Body: assets.ObjectWithEnum{
				Value:    "some value",
				Values:   []string{"some", "values"},
				Status:   assets.StatusEnumerationActive,
				Statuses: []assets.StatusEnumeration{assets.StatusEnumerationInactive},
			},
			Headers:     map[string]string{},
			RunningMode: &fullyFeaturedRouting,
		})

		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 422 for enum prams and body - wrong direct number enum",
			ExpectedStatus:  422,
			ExpectedBody:    "{\"type\":\"Unprocessable Entity\",\"title\":\"\",\"detail\":\"A request was made to operation 'TestEnums' but parameter 'value2' was not properly sent - Expected NumberEnumeration but got string\",\"status\":422,\"instance\":\"/gleece/validation/error/TestEnums\",\"extensions\":{\"error\":\"value2 must be one of \\\"1, 2\\\" options only but got 4\"}}",
			ExpendedHeaders: nil,
			Path:            "/e2e/test-enums",
			Method:          "POST",
			Query:           map[string]string{"value1": "active", "value2": "4"},
			Body: assets.ObjectWithEnum{
				Value:    "some value",
				Values:   []string{"some", "values"},
				Status:   assets.StatusEnumerationActive,
				Statuses: []assets.StatusEnumeration{assets.StatusEnumerationInactive},
			},
			Headers:     map[string]string{},
			RunningMode: &fullyFeaturedRouting,
		})

		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 422 for enum prams and body - wrong in struct enum validated by generated validator",
			ExpectedStatus:  422,
			ExpectedBody:    "{\"type\":\"Unprocessable Entity\",\"title\":\"\",\"detail\":\"A request was made to operation 'TestEnums' but body parameter 'value3' did not pass validation of 'ObjectWithEnum' - Field 'Status' failed validation with tag 'status_enumeration_enum'. \",\"status\":422,\"instance\":\"/gleece/validation/error/TestEnums\",\"extensions\":null}",
			ExpendedHeaders: nil,
			Path:            "/e2e/test-enums",
			Method:          "POST",
			Query:           map[string]string{"value1": "active", "value2": "2"},
			Body: assets.ObjectWithEnum{
				Value:    "some value",
				Values:   []string{"some", "values"},
				Status:   "blabla",
				Statuses: []assets.StatusEnumeration{assets.StatusEnumerationInactive},
			},
			Headers:     map[string]string{},
			RunningMode: &fullyFeaturedRouting,
		})

		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 422 for enum prams and body - wrong enum in optional param",
			ExpectedStatus:  422,
			ExpectedBody:    "{\"type\":\"Unprocessable Entity\",\"title\":\"\",\"detail\":\"A request was made to operation 'TestEnumsOptional' but parameter 'value1' was not properly sent - Expected StatusEnumeration but got string\",\"status\":422,\"instance\":\"/gleece/validation/error/TestEnumsOptional\",\"extensions\":{\"error\":\"value1 must be one of \\\"active, inactive\\\" options only but got active222\"}}",
			ExpendedHeaders: nil,
			Path:            "/e2e/test-enums-optional",
			Method:          "POST",
			Form:            map[string]string{},
			Headers:         map[string]string{"value1": "active222"},
			RunningMode:     &fullyFeaturedRouting,
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 422 for enum prams and body - wrong enum param external package",
			ExpectedStatus:      422,
			ExpectedBodyContain: "options only but got KiloBoom",
			ExpendedHeaders:     nil,
			Path:                "/e2e/external-packages",
			Method:              "POST",
			Body:                dtoInCm,
			Query:               map[string]string{"unit": "KiloBoom"},
			RunningMode:         &fullyFeaturedRouting,
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 422 for enum prams and body - wrong enum param external package",
			ExpectedStatus:      422,
			ExpectedBodyContain: "Field 'Unit' failed validation with tag 'length_units_enum'",
			ExpendedHeaders:     nil,
			Path:                "/e2e/external-packages-validation",
			Method:              "POST",
			Body:                dtoInBroken,
			Query:               map[string]string{"unit": "Kilometer"},
			RunningMode:         &fullyFeaturedRouting,
		})
	})

	It("Should NOT return status code 422 for wrong enum prams and body", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should NOT return status code 422 for enum prams and body - wrong direct number enum",
			ExpectedStatus:      200,
			ExpectedBodyContain: "active 10 inactive",
			ExpendedHeaders:     nil,
			Path:                "/e2e/test-enums-in-all/active",
			Method:              "POST",
			Form:                map[string]string{"value3": "inactive"},
			Headers:             map[string]string{"value2": "10"},
			RunningMode:         &exExtraRouting,
		})

		RunRouterTest(common.RouterTest{
			Name:            "Should NOT return status code 422 for enum prams and body - wrong enum in optional param",
			ExpectedStatus:  200,
			ExpectedBody:    "",
			ExpendedHeaders: nil,
			Path:            "/e2e/test-enums-optional",
			Method:          "POST",
			Form:            map[string]string{},
			Headers:         map[string]string{"value1": "active222"},
			RunningMode:     &exExtraRouting,
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should NOT return status code 422 for enum prams and body - wrong enum param external package",
			ExpectedStatus:      200,
			ExpectedBodyContain: "{\"value\":9991,\"unit\":\"KiloBoom\"}",
			ExpendedHeaders:     nil,
			Path:                "/e2e/external-packages",
			Method:              "POST",
			Body:                dtoInCm,
			Query:               map[string]string{"unit": "KiloBoom"},
			RunningMode:         &exExtraRouting,
		})

		RunRouterTest(common.RouterTest{
			Name:                "Should NOT return status code 422 for enum prams and body - wrong enum param external package",
			ExpectedStatus:      200,
			ExpectedBodyContain: "{\"value\":9.992,\"unit\":\"Kilometer\"}",
			ExpendedHeaders:     nil,
			Path:                "/e2e/external-packages",
			Method:              "POST",
			Body:                dtoInBroken,
			Query:               map[string]string{"unit": "Kilometer"},
			RunningMode:         &exExtraRouting,
		})
	})
})
