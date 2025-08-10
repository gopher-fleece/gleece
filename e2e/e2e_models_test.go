package e2e

import (
	"github.com/gopher-fleece/gleece/e2e/assets"
	"github.com/gopher-fleece/gleece/e2e/common"
	"github.com/haimkastner/unitsnet-go/units"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("E2E Models Spec", func() {
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

	It("Should return status code 200 for external packages referenced in local struct", func() {
		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 200 for external packages referenced in local struct",
			ExpectedStatus:      200,
			ExpectedBodyContain: "FootPerHour:KilometerPerHour",
			ExpendedHeaders:     nil,
			Path:                "/e2e/external-packages-unique-in-struct",
			Method:              "POST",
			Body: assets.UniqueExternalUsage{
				Enum: units.SpeedFootPerHour,
				Struct: units.SpeedDto{
					Value: 1000,
					Unit:  units.SpeedKilometerPerHour,
				},
			},
			Query:       map[string]string{},
			Headers:     map[string]string{},
			RunningMode: &allRouting,
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

	It("Should return status code 200 for embedded models fields", func() {

		body := assets.TheModel{
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
		}

		RunRouterTest(common.RouterTest{
			Name:            "Should return status code 200 for embedded models fields",
			ExpectedStatus:  200,
			ExpectedBody:    "{\"secondLevelModelField\":\"second level\",\"firstLevelModelField\":\"first level\",\"otherModelField\":\"other model\",\"modelField\":\"model field\"}",
			ExpendedHeaders: nil,
			Path:            "/e2e/embedded-structs",
			Method:          "POST",
			Body:            body,
			Query:           map[string]string{},
			Headers:         map[string]string{},
			RunningMode:     &allRouting,
		})
	})

	It("Should test struct with pointer fields", func() {
		data := "data"
		dataPtr := &data
		model := assets.TheModel{
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
		}

		body := assets.TheModelWithInnerPointer{
			Field1: &data,
			Field2: &dataPtr,
			Model:  &model,
			RecursiveModelWithPointer: &assets.RecursiveModelWithPointer{
				Prop1: data,
				Prop2: dataPtr,
				Prop3: &dataPtr,
				Prop4: &model,
			},
		}

		RunRouterTest(common.RouterTest{
			Name:                "Should return status code",
			ExpectedStatus:      200,
			ExpectedBodyContain: "data",
			ExpendedHeaders:     nil,
			Path:                "/e2e/structs-with-inner-pointer",
			Method:              "POST",
			Body:                body,
			Query:               map[string]string{},
			Headers:             map[string]string{},
			RunningMode:         &allRouting,
		})

		body2 := assets.TheModelWithInnerPointer{
			Field1: &data,
			Field2: &dataPtr,
			Model:  &model,
			RecursiveModelWithPointer: &assets.RecursiveModelWithPointer{
				Prop1: data,
				Prop3: &dataPtr,
				Prop4: &model,
			},
		}

		RunRouterTest(common.RouterTest{
			Name:                "Should return status code",
			ExpectedStatus:      422,
			ExpectedBodyContain: "Prop2",
			ExpendedHeaders:     nil,
			Path:                "/e2e/structs-with-inner-pointer",
			Method:              "POST",
			Body:                body2,
			Query:               map[string]string{},
			Headers:             map[string]string{},
			RunningMode:         &allRouting,
		})

	})

	It("Should return status code 422 for missing embedded models fields", func() {

		body := assets.TheModel{
			ModelField: "model field",
			FirstLevelModel: assets.FirstLevelModel{
				FirstLevelModelField: "first level",
				SecondLevelModel:     assets.SecondLevelModel{},
			},
			OtherModel: assets.OtherModel{
				OtherModelField: "other model",
			},
		}

		RunRouterTest(common.RouterTest{
			Name:                "Should return status code 422 for missing embedded models fields",
			ExpectedStatus:      422,
			ExpectedBodyContain: "A request was made to operation 'EmbeddedStructs' but body parameter 'data' did not pass validation of 'TheModel' - Field 'SecondLevelModelField' failed validation with tag 'required'.",
			ExpendedHeaders:     nil,
			Path:                "/e2e/embedded-structs",
			Method:              "POST",
			Body:                body,
			Query:               map[string]string{},
			Headers:             map[string]string{},
			RunningMode:         &allRouting,
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
			ExpectedBody:    "{\"type\":\"Unprocessable Entity\",\"title\":\"\",\"detail\":\"A request was made to operation 'TestEnums' but parameter 'value1' was not properly sent - Expected StatusEnumeration but got string\",\"status\":422,\"instance\":\"/validation/error/TestEnums\",\"extensions\":{\"error\":\"value1 must be one of \\\"active, inactive\\\" options only but got \\\"blabla\\\"\"}}",
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
			ExpectedBody:    "{\"type\":\"Unprocessable Entity\",\"title\":\"\",\"detail\":\"A request was made to operation 'TestEnums' but parameter 'value2' was not properly sent - Expected NumberEnumeration but got string\",\"status\":422,\"instance\":\"/validation/error/TestEnums\",\"extensions\":{\"error\":\"value2 must be one of \\\"1, 2\\\" options only but got \\\"4\\\"\"}}",
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
			ExpectedBody:    "{\"type\":\"Unprocessable Entity\",\"title\":\"\",\"detail\":\"A request was made to operation 'TestEnums' but body parameter 'value3' did not pass validation of 'ObjectWithEnum' - Field 'Status' failed validation with tag 'status_enumeration_enum'. \",\"status\":422,\"instance\":\"/validation/error/TestEnums\",\"extensions\":null}",
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
			ExpectedBody:    "{\"type\":\"Unprocessable Entity\",\"title\":\"\",\"detail\":\"A request was made to operation 'TestEnumsOptional' but parameter 'value1' was not properly sent - Expected StatusEnumeration but got string\",\"status\":422,\"instance\":\"/validation/error/TestEnumsOptional\",\"extensions\":{\"error\":\"value1 must be one of \\\"active, inactive\\\" options only but got \\\"active222\\\"\"}}",
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
			ExpectedBodyContain: "options only but got \\\"KiloBoom\\\"",
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
