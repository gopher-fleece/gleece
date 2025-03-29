package assets

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/gopher-fleece/runtime"
	"github.com/haimkastner/unitsnet-go/units"
	"github.com/labstack/echo/v4"
)

type StatusEnumeration string

const (
	StatusEnumerationActive   StatusEnumeration = "active"
	StatusEnumerationInactive StatusEnumeration = "inactive"
)

type NumberEnumeration int

const (
	NumberEnumerationOne NumberEnumeration = 1
	NumberEnumerationTwo NumberEnumeration = 2
)

type ObjectWithEnum struct {
	Value    string              `json:"value"`
	Values   []string            `json:"values"`
	Status   StatusEnumeration   `json:"status" validate:"required,status_enumeration_enum"`
	Statuses []StatusEnumeration `json:"statuses"`
}

// @Route(/e2e)
type E2EController struct {
	runtime.GleeceController // Embedding the GleeceController to inherit its methods
}

// @Method(GET) This text is not part of the OpenAPI spec
// @Route(/simple-get)
func (ec *E2EController) SimpleGet() (string, error) {
	ec.SetHeader("X-Test-Header", "test")
	return "works", nil
}

// @Method(GET) This text is not part of the OpenAPI spec
// @Route(/simple-get-empty-string)
func (ec *E2EController) SimpleGetEmptyString() (string, error) {
	return "", nil
}

// @Method(GET) This text is not part of the OpenAPI spec
// @Route(/simple-get-ptr-string)
func (ec *E2EController) SimpleGetPtrString() (*string, error) {
	res := "ptr"
	return &res, nil
}

// @Method(GET) This text is not part of the OpenAPI spec
// @Route(/simple-get-null-string)
func (ec *E2EController) SimpleGetNullString() (*string, error) {
	return nil, nil
}

type BodyResponse struct {
	Data string `json:"data"`
}

// @Method(GET) This text is not part of the OpenAPI spec
// @Route(/simple-get-object)
func (ec *E2EController) SimpleGetObject() (BodyResponse, error) {
	return BodyResponse{Data: "BodyResponse"}, nil
}

// @Method(GET) This text is not part of the OpenAPI spec
// @Route(/simple-get-object-ptr)
func (ec *E2EController) SimpleGetObjectPtr() (*BodyResponse, error) {
	return &BodyResponse{Data: "BodyResponse"}, nil
}

// @Method(GET) This text is not part of the OpenAPI spec
// @Route(/simple-get-object-null)
func (ec *E2EController) SimpleGetObjectNull() (*BodyResponse, error) {
	return nil, nil
}

// @Method(GET)
// @Route(/primitive-return-type)
// @Response(200)
// @ErrorResponse(500)
func (ec *E2EController) PrimitiveReturnType() (int, error) {
	return 20, nil
}

// @Method(GET)
// @Route(/primitive-array-return-type)
// @Response(200)
// @ErrorResponse(500)
func (ec *E2EController) PrimitiveArrayReturnType() ([]int, error) {
	return []int{20}, nil
}

// @Method(GET)
// @Route(/primitive-alias-return-type)
// @Response(200)
// @ErrorResponse(500)
func (ec *E2EController) PrimitiveAliasReturnType() (NumberEnumeration, error) {
	return NumberEnumerationOne, nil
}

// @Method(GET)
// @Route(/primitive-alias-array-return-type)
// @Response(200)
// @ErrorResponse(500)
func (ec *E2EController) PrimitiveAliasArrayReturnType() ([]NumberEnumeration, error) {
	return []NumberEnumeration{NumberEnumerationOne}, nil
}

// @Method(GET)
// @Route(/simple-get-empty)
// @Query(queryParam)
func (ec *E2EController) SimpleGetEmpty(queryParam string) error {
	if queryParam == "200" {
		ec.SetStatus(runtime.StatusOK)
	}
	return nil
}

// @Method(GET) This text is not part of the OpenAPI spec
// @Route(/get-with-all-params/{pathParam})
// @Query(queryParam)
// @Path(pathParam)
// @Header(headerParam)
func (ec *E2EController) GetWithAllParams(queryParam string, pathParam string, headerParam string) (string, error) {
	if queryParam == "204" {
		ec.SetStatus(runtime.StatusNoContent)
	}
	return pathParam + queryParam + headerParam, nil
}

// @Method(GET) This text is not part of the OpenAPI spec
// @Route(/get-with-all-params-ptr/{pathParam})
// @Query(queryParam)
// @Path(pathParam)
// @Header(headerParam)
func (ec *E2EController) GetWithAllParamsPtr(queryParam *string, pathParam *string, headerParam *string) (string, error) {
	if queryParam == nil {
		queryParam = new(string)
	}
	if headerParam == nil {
		headerParam = new(string)
	}
	return *pathParam + *queryParam + *headerParam, nil
}

// @Method(GET) This text is not part of the OpenAPI spec
// @Route(/get-with-all-params-required-ptr/{pathParam})
// @Query(queryParam, { validate: "required" })
// @Path(pathParam, { validate: "required" })
// @Header(headerParam, { validate: "required" })
func (ec *E2EController) GetWithAllParamsRequiredPtr(queryParam *string, pathParam *string, headerParam *string) (string, error) {
	if queryParam == nil {
		queryParam = new(string)
	}
	if headerParam == nil {
		headerParam = new(string)
	}
	return *pathParam + *queryParam + *headerParam, nil
}

type BodyInfo struct {
	BodyParam string `json:"bodyParam" validate:"required"`
}

type BodyInfo2 struct {
	BodyParam int `json:"bodyParam"`
}

// @Method(POST) This text is not part of the OpenAPI spec
// @Route(/post-with-all-params-body)
// @Query(queryParam)
// @Body(theBody)
// @Header(headerParam)
func (ec *E2EController) PostWithAllParamsWithBody(queryParam string, headerParam string, theBody BodyInfo) (BodyInfo, error) {
	return BodyInfo{
		BodyParam: queryParam + headerParam + theBody.BodyParam,
	}, nil
}

// @Method(POST) This text is not part of the OpenAPI spec
// @Route(/post-with-all-params-body-ptr)
// @Query(queryParam)
// @Body(theBody)
// @Header(headerParam)
func (ec *E2EController) PostWithAllParamsWithBodyPtr(queryParam *string, headerParam *string, theBody *BodyInfo) (*BodyInfo, error) {
	if queryParam == nil {
		queryParam = new(string)
	}
	if headerParam == nil {
		headerParam = new(string)
	}
	if theBody == nil {
		theBody = new(BodyInfo)
		theBody.BodyParam = "empty"
	}
	return &BodyInfo{
		BodyParam: *queryParam + *headerParam + theBody.BodyParam,
	}, nil
}

// @Method(POST) This text is not part of the OpenAPI spec
// @Route(/post-with-all-params-body-required-ptr)
// @Body(theBody, { validate: "required" })
func (ec *E2EController) PostWithAllParamsWithBodyRequiredPtr(theBody *BodyInfo) (*BodyInfo, error) {
	return &BodyInfo{
		BodyParam: theBody.BodyParam,
	}, nil
}

// @Method(GET) This text is not part of the OpenAPI spec
// @Route(/get-header-start-with-letter)
// @Header(headerParam, { validate: "required,validate_starts_with_letter" })
func (ec *E2EController) GetHeaderStartWithLetter(headerParam string) (string, error) {
	return headerParam, nil
}

// @Route(/e2e)
// @Security(securitySchemaName, { scopes: ["class"] })
type E2EClassSecController struct {
	runtime.GleeceController // Embedding the GleeceController to inherit its methods
}

// @Method(GET)
// @Header(headerParam, { name: "x-test-scopes" })
// @Route(/with-default-class-security)
func (ec *E2EClassSecController) WithDefaultClassSecurity(headerParam string) (string, error) {
	return headerParam, nil
}

// @Method(GET)
// @Route(/with-default-override-class-security)
// @Header(headerParam, { name: "x-test-scopes" })
// @Security(securitySchemaName, { scopes: ["method"] })
func (ec *E2EClassSecController) WithOverrideClassSecurity(headerParam string) (string, error) {
	return headerParam, nil
}

// @Method(GET)
// @Header(headerParam, { name: "x-test-scopes" })
// @Route(/with-default-config-security)
func (ec *E2EController) WithDefaultConfigSecurity(headerParam string) (string, error) {
	return headerParam, nil
}

// @Method(GET)
// @Route(/with-one-security)
// @Header(headerParam, { name: "x-test-scopes" })
// @Security(securitySchemaName, { scopes: ["other"] })
func (ec *E2EController) WithOneSecurity(headerParam string) (string, error) {
	return headerParam, nil
}

// @Method(GET)
// @Route(/with-two-security)
// @Header(headerParam, { name: "x-test-scopes" })
// @Security(securitySchemaName, { scopes: ["other"] })
// @Security(securitySchemaName2, { scopes: ["write", "read"] })
func (ec *E2EController) WithTwoSecurity(headerParam string) (string, error) {
	return headerParam, nil
}

// @Method(GET)
// @Route(/with-two-security-same-method)
// @Header(headerParam, { name: "x-test-scopes" })
// @Security(securitySchemaName, { scopes: ["other"] })
// @Security(securitySchemaName, { scopes: ["write", "read"] })
func (ec *E2EController) WithTwoSecuritySameMethod(headerParam string) (string, error) {
	context := ec.GetContext()
	ginContext, isGin := context.(*gin.Context)
	if isGin {

		ginContext.GetHeader("x-some-header")
	}

	// Now echo...
	echoContext, isEcho := context.(echo.Context)
	if isEcho {
		echoContext.Request().Header.Get("x-some-header")
	}
	return headerParam, nil
}

// @Method(GET)
// @Route(/default-error)
func (ec *E2EController) DefaultError() error {
	return fmt.Errorf("default error")
}

// @Method(GET)
// @Route(/default-error-with-payload)
func (ec *E2EController) DefaultErrorWithPayload() (*string, error) {
	return nil, fmt.Errorf("default error")
}

// TODO: is pointer error not officially supported?
// // @Method(GET)
// // @Route(/default-error-ptr)
// func (ec *E2EController) DefaultErrorPtr() *error {
// 	err := fmt.Errorf("default error")
// 	return &err
// }

type CustomError struct {
	error
	Message string `json:"message"`
}

// @Method(GET)
// @Route(/custom-error)
func (ec *E2EController) CustomError() CustomError {
	return CustomError{
		Message: "custom error",
	}
}

// @Method(GET)
// @Route(/custom-error-ptr)
func (ec *E2EController) CustomPtrError() *CustomError {
	return &CustomError{
		Message: "custom error",
	}
}

// @Method(GET)
// @Route(/503-error-code)
func (ec *E2EController) Error503() error {
	ec.SetStatus(runtime.StatusServiceUnavailable)
	return fmt.Errorf("default error")
}

// @Method(GET)
// @Route(/custom-error-503)
func (ec *E2EController) CustomError503() CustomError {
	ec.SetStatus(runtime.StatusServiceUnavailable)
	return CustomError{
		Message: "custom error",
	}
}

// @Method(GET)
// @Route(/context-access)
func (ec *E2EController) ContextAccess() error {
	context := ec.GetContext()
	switch context.(type) {
	case *gin.Context:
		ginContext := context.(*gin.Context)
		ginContext.Header("x-context-pass", "true")
	case echo.Context:
		echoContext := context.(echo.Context)
		echoContext.Response().Header().Set("x-context-pass", "true")
	case *fiber.Ctx:
		fiberContext := context.(*fiber.Ctx)
		fiberContext.Set("x-context-pass", "true")
	case *http.Request:
		httpRequest := context.(*http.Request)
		ec.SetHeader("x-context-host", httpRequest.Host)
		ec.SetHeader("x-context-pass", "true")
	}
	return nil
}

// @Method(GET)
// @Route(/http-method)
func (ec *E2EController) Get() error {
	ec.SetHeader("x-method", "get")
	return nil
}

// @Method(POST)
// @Route(/http-method)
func (ec *E2EController) Post() error {
	ec.SetHeader("x-method", "post")
	return nil
}

// @Method(PUT)
// @Route(/http-method)
func (ec *E2EController) Put() error {
	ec.SetHeader("x-method", "put")
	return nil
}

// @Method(DELETE)
// @Route(/http-method)
func (ec *E2EController) Delete() error {
	ec.SetHeader("x-method", "delete")
	return nil
}

// @Method(PATCH)
// @Route(/http-method)
func (ec *E2EController) Patch() error {
	ec.SetHeader("x-method", "patch")
	return nil
}

// @Method(GET)
// @Route(/template-context-1)
// @TemplateContext(LEVEL, {value: "high"})
func (ec *E2EController) TemplateContext1() (string, error) {
	return "works", nil
}

// @Method(GET)
// @Route(/template-context-2)
// @TemplateContext(LEVEL, {value: "low"})
// @TemplateContext(MODE, {mode: "100"})
func (ec *E2EController) TemplateContext2() (string, error) {
	return "works", nil
}

// @Description Create a new user
// @Method(POST) This text is not part of the OpenAPI spec
// @Route(/form) Same here
// @FormField(item1) The user's ID
// @FormField(item2) The user's ID
// @Response(200) The ID of the newly created user
// @ErrorResponse(500) The error when process failed
func (ec *E2EController) TestForm(item1 string, item2 string) (string, error) {
	return item1 + item2, nil
}

type ResponseTest struct {
	Success string `json:"success"`
	Index   int    `json:"index" validate:"required,gte=0"`
}

// @Method(POST)
// @Route(/test-response-validation)
// @Response(200)
// @ErrorResponse(500)
func (ec *E2EController) TestResponseValidation() (ResponseTest, error) {
	return ResponseTest{
		Success: "success",
		Index:   -1,
	}, nil
}

// @Method(POST)
// @Route(/test-response-validation-ptr)
// @Response(200) The ID of the newly created user
// @ErrorResponse(500) The error when process failed
func (ec *E2EController) TestResponseValidationPtr() (*ResponseTest, error) {
	return &ResponseTest{
		Success: "success",
		Index:   -1,
	}, nil
}

// @Method(POST)
// @Route(/test-response-validation-null)
// @Response(200) The ID of the newly created user
// @ErrorResponse(500) The error when process failed
func (ec *E2EController) TestResponseValidationNull() (*ResponseTest, error) {
	return nil, nil
}

// @Method(POST)
// @Route(/test-primitive-conversions)
// @Query(value1)
// @Query(value2)
// @Query(value3)
// @Query(value4)
// @Response(200) The ID of the newly created user
// @ErrorResponse(500) The error when process failed
func (ec *E2EController) TestPrimitiveConversions(value1 int64, value2 bool, value3 int, value4 float64) (string, error) {
	return fmt.Sprintf("%d %t %d %f", value1, value2, value3, value4), nil
}

// @Method(POST)
// @Route(/test-enums)
// @Query(value1)
// @Query(value2)
// @Body(value3)
// @Response(200) The ID of the newly created user
// @ErrorResponse(500) The error when process failed
func (ec *E2EController) TestEnums(value1 StatusEnumeration, value2 NumberEnumeration, value3 ObjectWithEnum) (ObjectWithEnum, error) {
	return ObjectWithEnum{
		Value: string(fmt.Sprintf("%s %d", value1, value2)),
		Values: []string{
			string(value1),
			string(fmt.Sprintf("%d", value2)),
		},
		Status:   value3.Status,
		Statuses: value3.Statuses,
	}, nil
}

// @Method(POST)
// @Route(/test-enums-in-all/{value1})
// @Path(value1)
// @Header(value2)
// @FormField(value3)
// @Response(200) The ID of the newly created user
// @ErrorResponse(500) The error when process failed
func (ec *E2EController) TestEnumsInAll(value1 StatusEnumeration, value2 NumberEnumeration, value3 StatusEnumeration) (string, error) {
	return fmt.Sprintf("%s %d %s", value1, value2, value3), nil
}

// @Method(POST)
// @Route(/test-enums-optional)
// @Header(value1)
// @Response(200) The ID of the newly created user
// @ErrorResponse(500) The error when process failed
func (ec *E2EController) TestEnumsOptional(value1 *StatusEnumeration) (string, error) {
	if value1 == nil {
		return "nil", nil
	}
	return string(*value1), nil
}

// @Method(POST)
// @Route(/external-packages)
// @Query(unit)
// @Body(data)
// @Response(200)
// @ErrorResponse(500)
func (ec *E2EController) ExternalPackages(unit *units.LengthUnits, data units.LengthDto) (units.LengthDto, error) {
	lf := units.LengthFactory{}
	unitData, _ := lf.FromDto(data)

	// Broken unit body
	if math.IsNaN(unitData.BaseValue()) {
		unitData, _ = lf.FromMeters(9992)
	}

	returnDto := unitData.ToDto(unit)

	// Broken unit param
	if math.IsNaN(returnDto.Value) {
		returnDto.Value = 9991
	}

	return returnDto, nil
}

type LengthDtoWithValidation struct {
	// units.LengthDto
	Value float64           `json:"value"`
	Unit  units.LengthUnits `json:"unit" validate:"required,length_units_enum"` // Extend it, to add the validation for the enum
}

// @Method(POST)
// @Route(/external-packages-validation)
// @Query(unit)
// @Body(data)
// @Response(200)
// @ErrorResponse(500)
func (ec *E2EController) ExternalPackagesValidation(unit *units.LengthUnits, data LengthDtoWithValidation) (units.LengthDto, error) {
	lf := units.LengthFactory{}
	// Dump to json
	dataJson, _ := json.Marshal(data)
	unitData, _ := lf.FromDtoJSON(dataJson)
	return unitData.ToDto(unit), nil
}

type BlaBla struct {
	ListOfLength []units.LengthDto `json:"listOfLength"`
}

// @Method(POST)
// @Route(/arrays-in-body-and-res)
// @Body(data)
// @Response(200)
// @ErrorResponse(500)
func (ec *E2EController) ArraysInBodyAndRes(data []units.LengthDto) ([]units.LengthDto, error) {
	return data, nil
}

// @Method(POST)
// @Route(/arrays-inside-body-and-res)
// @Body(data)
// @Response(200)
// @ErrorResponse(500)
func (ec *E2EController) ArraysInsideBodyAndRes(data *[]BlaBla) (*[]BlaBla, error) {
	return data, nil
}

type BlaBla2 struct {
	Value int `json:"value" validate:"required,gte=0"`
}

// @Method(POST)
// @Route(/deep-arrays-with-validation)
// @Body(data)
// @Response(200)
// @ErrorResponse(500)
func (ec *E2EController) DeepArraysWithValidation(data [][]BlaBla2) ([][][]BlaBla2, error) {
	return [][][]BlaBla2{data}, nil
}

type FirstLevelModel struct {
	SecondLevelModel
	FirstLevelModelField string `json:"firstLevelModelField" validate:"required"`
}

type SecondLevelModel struct {
	SecondLevelModelField string `json:"secondLevelModelField" validate:"required"`
}

type OtherModel struct {
	OtherModelField string `json:"otherModelField" validate:"required"`
}

type TheModel struct {
	FirstLevelModel
	OtherModel
	ModelField string `json:"modelField" validate:"required"`
}

// @Method(POST)
// @Route(/embedded-structs)
// @Body(data)
// @Response(200)
// @ErrorResponse(500)
func (ec *E2EController) EmbeddedStructs(data TheModel) (TheModel, error) {
	return data, nil
}
