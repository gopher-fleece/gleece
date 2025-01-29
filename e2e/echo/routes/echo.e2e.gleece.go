/*
--
This file is automatically generated. Any manual changes to this file may be overwritten.
It includes routes and handlers by the Gleece API Routes Generator.
--
Authors: Haim Kastner & Yuval Pomerchik
Generated by: Gleece Routes Generator
Generated Date: 29 Jan 25 23:07 IST
--
Usage:
Refer to the Gleece documentation for details on how to use the generated routes and handlers.
--
Repository: https://github.com/gopher-fleece/gleece
--
*/
package routes
import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"github.com/go-playground/validator/v10"
	E2EControllerImport "github.com/gopher-fleece/gleece/e2e/assets"
	RequestAuth "github.com/gopher-fleece/gleece/e2e/echo/auth"
	"github.com/gopher-fleece/gleece/external"
	"github.com/gopher-fleece/gleece/infrastructure/validation"
	"github.com/labstack/echo/v4"
	Param19theBody "github.com/gopher-fleece/gleece/e2e/assets"
	Param24theBody "github.com/gopher-fleece/gleece/e2e/assets"
	Param27theBody "github.com/gopher-fleece/gleece/e2e/assets"
)
var validatorInstance *validator.Validate
var urlParamRegex *regexp.Regexp
type SecurityListRelation string
const (
	SecurityListRelationAnd SecurityListRelation = "AND"
)
type SecurityCheckList struct {
	Checks   []external.SecurityCheck
	Relation SecurityListRelation
}
func getStatusCode(controller external.Controller, hasReturnValue bool, err error) int {
	if controller.GetStatus() != nil {
		return int(*controller.GetStatus())
	}
	if err != nil {
		return http.StatusInternalServerError
	}
	if hasReturnValue {
		return http.StatusOK
	}
	return http.StatusNoContent
}
func bindAndValidateBody[TOutput any](ctx echo.Context, contentType string, validation string, output **TOutput) error {
	var err error
	bodyBytes, err := io.ReadAll(ctx.Request().Body)
	if err != nil || len(bodyBytes) == 0 {
		if strings.Contains(validation, "required") {
			return fmt.Errorf("body is required but was not provided")
		}
		return nil
	}
	var deserializedOutput TOutput
	switch contentType {
	case "application/json":
		err = json.Unmarshal(bodyBytes, &deserializedOutput)
	default:
		return fmt.Errorf("content-type %s is not currently supported by the validation subsystem", contentType)
	}
	if err != nil {
		return err
	}
	if err = validatorInstance.Struct(&deserializedOutput); err != nil {
		return err
	}
	*output = &deserializedOutput
	return nil
}
func toEchoUrl(url string) string {
	processedUrl := urlParamRegex.ReplaceAllString(url, ":$1")
	processedUrl = strings.ReplaceAll(processedUrl, "//", "/")
	if processedUrl == "" {
		return "/"
	}
	if !strings.HasPrefix(processedUrl, "/") {
		processedUrl = "/" + processedUrl
	}
	return processedUrl
}
func authorize(ctx echo.Context, checksLists []SecurityCheckList) *external.SecurityError {
	var lastError *external.SecurityError
	for _, list := range checksLists {
		if list.Relation != SecurityListRelationAnd {
			panic(
				"Encountered a security list relation of type '%s' - this is unexpected and indicates a bug in Gleece itself." +
					"Please open an issue at https://github.com/gopher-fleece/gleece/issues",
			)
		}
		// Iterate over each security list
		encounteredErrorInList := false
		for _, check := range list.Checks {
			secErr := RequestAuth.GleeceRequestAuthorization(ctx, check)
			if secErr != nil {
				lastError = secErr
				encounteredErrorInList = true
				break
			}
		}
		// If no error was encountered, validation is considered successful
		// otherwise, we continue over to the next iteration whilst keeping track of the last error
		if !encounteredErrorInList {
			return nil
		}
	}
	// If we got here it means authentication has failed
	return lastError
}
func handleAuthorizationError(ctx echo.Context, authErr *external.SecurityError, operationId string) error {
	statusCode := int(authErr.StatusCode)
	if authErr.CustomError != nil {
		// For now, we support JSON only
		return ctx.JSON(statusCode, authErr.CustomError.Payload)
	}
	stdError := external.Rfc7807Error{
		Type:     http.StatusText(statusCode),
		Detail:   authErr.Message,
		Status:   statusCode,
		Instance: "/gleece/authorization/error/" + operationId,
	}
	return ctx.JSON(statusCode, stdError)
}
func wrapValidatorError(validatorErr error, operationId string, fieldName string) external.Rfc7807Error {
	return external.Rfc7807Error{
		Type: http.StatusText(http.StatusUnprocessableEntity),
		Detail: fmt.Sprintf(
			"A request was made to operation '%s' but parameter '%s' did not pass validation - %s",
			operationId,
			fieldName,
			validation.ExtractValidationErrorMessage(validatorErr, &fieldName),
		),
		Status:   http.StatusUnprocessableEntity,
		Instance: fmt.Sprintf("/gleece/validation/error/%s", operationId),
	}
}
func RegisterRoutes(engine *echo.Echo) {
	validatorInstance = validator.New()
	urlParamRegex = regexp.MustCompile(`\{([\w\d-_]+)\}`)
	// E2EController
	engine.GET(toEchoUrl("/e2e/simple-get"), func(ctx echo.Context) error {
		authErr := authorize(
			ctx,
			[]SecurityCheckList{
				{
					Relation: SecurityListRelationAnd,
					Checks: []external.SecurityCheck{
						{
							SchemaName: "securitySchemaName",
							Scopes: []string{
								"read",
								"write",
							},
						},
					},
				},
			},
		)
		if authErr != nil {
			return handleAuthorizationError(ctx, authErr, "SimpleGet")
		}
		controller := E2EControllerImport.E2EController{}
		controller.InitController(ctx)
		value, opError := controller.SimpleGet()
		for key, value := range controller.GetHeaders() {
			ctx.Response().Header().Set(key, value)
		}
		statusCode := getStatusCode(&controller, true, opError)
		if opError != nil {
			stdError := external.Rfc7807Error{
				Type:       http.StatusText(statusCode),
				Detail:     "Encountered an error during operation 'SimpleGet'",
				Status:     statusCode,
				Instance:   "/gleece/controller/error/SimpleGet",
				Extensions: map[string]string{"error": opError.Error()},
			}
			return ctx.JSON(statusCode, stdError)
		}
		return ctx.JSON(statusCode, value)
	})
	engine.GET(toEchoUrl("/e2e/get-with-all-params/{pathParam}"), func(ctx echo.Context) error {
		authErr := authorize(
			ctx,
			[]SecurityCheckList{
				{
					Relation: SecurityListRelationAnd,
					Checks: []external.SecurityCheck{
						{
							SchemaName: "securitySchemaName",
							Scopes: []string{
								"read",
								"write",
							},
						},
					},
				},
			},
		)
		if authErr != nil {
			return handleAuthorizationError(ctx, authErr, "GetWithAllParams")
		}
		controller := E2EControllerImport.E2EController{}
		controller.InitController(ctx)
		var queryParamRawPtr *string = nil
		queryParamRaw := ctx.QueryParam("queryParam")
		isqueryParamExists := ctx.Request().URL.Query().Has("queryParam")
		if isqueryParamExists {
			queryParam := queryParamRaw
			queryParamRawPtr = &queryParam
		}
		if validatorErr := validatorInstance.Var(queryParamRawPtr, "required"); validatorErr != nil {
			fieldName := "queryParam"
			validationError := wrapValidatorError(validatorErr, "GetWithAllParams", fieldName)
			return ctx.JSON(http.StatusUnprocessableEntity, validationError)
		}
		var pathParamRawPtr *string = nil
		pathParamRaw := ctx.Param("pathParam")
		ispathParamExists := true // if parameter is in route but not provided, it won't reach this handler
		if ispathParamExists {
			pathParam := pathParamRaw
			pathParamRawPtr = &pathParam
		}
		if validatorErr := validatorInstance.Var(pathParamRawPtr, "required"); validatorErr != nil {
			fieldName := "pathParam"
			validationError := wrapValidatorError(validatorErr, "GetWithAllParams", fieldName)
			return ctx.JSON(http.StatusUnprocessableEntity, validationError)
		}
		var headerParamRawPtr *string = nil
		headerParamRaw := ctx.Request().Header.Get("headerParam")
		_, isheaderParamExists := ctx.Request().Header["headerParam"]
		if !isheaderParamExists {
			// In echo, the ctx.Request().Header["key"] is not 100% reliable, so we need other check, but only if is was not found in the first method
			headerValues := ctx.Request().Header.Values("headerParam")
			isheaderParamExists = len(headerValues) > 0
		}
		if isheaderParamExists {
			headerParam := headerParamRaw
			headerParamRawPtr = &headerParam
		}
		if validatorErr := validatorInstance.Var(headerParamRawPtr, "required"); validatorErr != nil {
			fieldName := "headerParam"
			validationError := wrapValidatorError(validatorErr, "GetWithAllParams", fieldName)
			return ctx.JSON(http.StatusUnprocessableEntity, validationError)
		}
		value, opError := controller.GetWithAllParams(*queryParamRawPtr, *pathParamRawPtr, *headerParamRawPtr)
		for key, value := range controller.GetHeaders() {
			ctx.Response().Header().Set(key, value)
		}
		statusCode := getStatusCode(&controller, true, opError)
		if opError != nil {
			stdError := external.Rfc7807Error{
				Type:       http.StatusText(statusCode),
				Detail:     "Encountered an error during operation 'GetWithAllParams'",
				Status:     statusCode,
				Instance:   "/gleece/controller/error/GetWithAllParams",
				Extensions: map[string]string{"error": opError.Error()},
			}
			return ctx.JSON(statusCode, stdError)
		}
		return ctx.JSON(statusCode, value)
	})
	engine.GET(toEchoUrl("/e2e/get-with-all-params-ptr/{pathParam}"), func(ctx echo.Context) error {
		authErr := authorize(
			ctx,
			[]SecurityCheckList{
				{
					Relation: SecurityListRelationAnd,
					Checks: []external.SecurityCheck{
						{
							SchemaName: "securitySchemaName",
							Scopes: []string{
								"read",
								"write",
							},
						},
					},
				},
			},
		)
		if authErr != nil {
			return handleAuthorizationError(ctx, authErr, "GetWithAllParamsPtr")
		}
		controller := E2EControllerImport.E2EController{}
		controller.InitController(ctx)
		var queryParamRawPtr *string = nil
		queryParamRaw := ctx.QueryParam("queryParam")
		isqueryParamExists := ctx.Request().URL.Query().Has("queryParam")
		if isqueryParamExists {
			queryParam := queryParamRaw
			queryParamRawPtr = &queryParam
		}
		var pathParamRawPtr *string = nil
		pathParamRaw := ctx.Param("pathParam")
		ispathParamExists := true // if parameter is in route but not provided, it won't reach this handler
		if ispathParamExists {
			pathParam := pathParamRaw
			pathParamRawPtr = &pathParam
		}
		if validatorErr := validatorInstance.Var(pathParamRawPtr, "required"); validatorErr != nil {
			fieldName := "pathParam"
			validationError := wrapValidatorError(validatorErr, "GetWithAllParamsPtr", fieldName)
			return ctx.JSON(http.StatusUnprocessableEntity, validationError)
		}
		var headerParamRawPtr *string = nil
		headerParamRaw := ctx.Request().Header.Get("headerParam")
		_, isheaderParamExists := ctx.Request().Header["headerParam"]
		if !isheaderParamExists {
			// In echo, the ctx.Request().Header["key"] is not 100% reliable, so we need other check, but only if is was not found in the first method
			headerValues := ctx.Request().Header.Values("headerParam")
			isheaderParamExists = len(headerValues) > 0
		}
		if isheaderParamExists {
			headerParam := headerParamRaw
			headerParamRawPtr = &headerParam
		}
		value, opError := controller.GetWithAllParamsPtr(queryParamRawPtr, pathParamRawPtr, headerParamRawPtr)
		for key, value := range controller.GetHeaders() {
			ctx.Response().Header().Set(key, value)
		}
		statusCode := getStatusCode(&controller, true, opError)
		if opError != nil {
			stdError := external.Rfc7807Error{
				Type:       http.StatusText(statusCode),
				Detail:     "Encountered an error during operation 'GetWithAllParamsPtr'",
				Status:     statusCode,
				Instance:   "/gleece/controller/error/GetWithAllParamsPtr",
				Extensions: map[string]string{"error": opError.Error()},
			}
			return ctx.JSON(statusCode, stdError)
		}
		return ctx.JSON(statusCode, value)
	})
	engine.GET(toEchoUrl("/e2e/get-with-all-params-required-ptr/{pathParam}"), func(ctx echo.Context) error {
		authErr := authorize(
			ctx,
			[]SecurityCheckList{
				{
					Relation: SecurityListRelationAnd,
					Checks: []external.SecurityCheck{
						{
							SchemaName: "securitySchemaName",
							Scopes: []string{
								"read",
								"write",
							},
						},
					},
				},
			},
		)
		if authErr != nil {
			return handleAuthorizationError(ctx, authErr, "GetWithAllParamsRequiredPtr")
		}
		controller := E2EControllerImport.E2EController{}
		controller.InitController(ctx)
		var queryParamRawPtr *string = nil
		queryParamRaw := ctx.QueryParam("queryParam")
		isqueryParamExists := ctx.Request().URL.Query().Has("queryParam")
		if isqueryParamExists {
			queryParam := queryParamRaw
			queryParamRawPtr = &queryParam
		}
		if validatorErr := validatorInstance.Var(queryParamRawPtr, "required"); validatorErr != nil {
			fieldName := "queryParam"
			validationError := wrapValidatorError(validatorErr, "GetWithAllParamsRequiredPtr", fieldName)
			return ctx.JSON(http.StatusUnprocessableEntity, validationError)
		}
		var pathParamRawPtr *string = nil
		pathParamRaw := ctx.Param("pathParam")
		ispathParamExists := true // if parameter is in route but not provided, it won't reach this handler
		if ispathParamExists {
			pathParam := pathParamRaw
			pathParamRawPtr = &pathParam
		}
		if validatorErr := validatorInstance.Var(pathParamRawPtr, "required"); validatorErr != nil {
			fieldName := "pathParam"
			validationError := wrapValidatorError(validatorErr, "GetWithAllParamsRequiredPtr", fieldName)
			return ctx.JSON(http.StatusUnprocessableEntity, validationError)
		}
		var headerParamRawPtr *string = nil
		headerParamRaw := ctx.Request().Header.Get("headerParam")
		_, isheaderParamExists := ctx.Request().Header["headerParam"]
		if !isheaderParamExists {
			// In echo, the ctx.Request().Header["key"] is not 100% reliable, so we need other check, but only if is was not found in the first method
			headerValues := ctx.Request().Header.Values("headerParam")
			isheaderParamExists = len(headerValues) > 0
		}
		if isheaderParamExists {
			headerParam := headerParamRaw
			headerParamRawPtr = &headerParam
		}
		if validatorErr := validatorInstance.Var(headerParamRawPtr, "required"); validatorErr != nil {
			fieldName := "headerParam"
			validationError := wrapValidatorError(validatorErr, "GetWithAllParamsRequiredPtr", fieldName)
			return ctx.JSON(http.StatusUnprocessableEntity, validationError)
		}
		value, opError := controller.GetWithAllParamsRequiredPtr(queryParamRawPtr, pathParamRawPtr, headerParamRawPtr)
		for key, value := range controller.GetHeaders() {
			ctx.Response().Header().Set(key, value)
		}
		statusCode := getStatusCode(&controller, true, opError)
		if opError != nil {
			stdError := external.Rfc7807Error{
				Type:       http.StatusText(statusCode),
				Detail:     "Encountered an error during operation 'GetWithAllParamsRequiredPtr'",
				Status:     statusCode,
				Instance:   "/gleece/controller/error/GetWithAllParamsRequiredPtr",
				Extensions: map[string]string{"error": opError.Error()},
			}
			return ctx.JSON(statusCode, stdError)
		}
		return ctx.JSON(statusCode, value)
	})
	engine.POST(toEchoUrl("/e2e/post-with-all-params-body"), func(ctx echo.Context) error {
		authErr := authorize(
			ctx,
			[]SecurityCheckList{
				{
					Relation: SecurityListRelationAnd,
					Checks: []external.SecurityCheck{
						{
							SchemaName: "securitySchemaName",
							Scopes: []string{
								"read",
								"write",
							},
						},
					},
				},
			},
		)
		if authErr != nil {
			return handleAuthorizationError(ctx, authErr, "PostWithAllParamsWithBody")
		}
		controller := E2EControllerImport.E2EController{}
		controller.InitController(ctx)
		var conversionErr error
		var queryParamRawPtr *string = nil
		queryParamRaw := ctx.QueryParam("queryParam")
		isqueryParamExists := ctx.Request().URL.Query().Has("queryParam")
		if isqueryParamExists {
			queryParam := queryParamRaw
			queryParamRawPtr = &queryParam
		}
		if validatorErr := validatorInstance.Var(queryParamRawPtr, "required"); validatorErr != nil {
			fieldName := "queryParam"
			validationError := wrapValidatorError(validatorErr, "PostWithAllParamsWithBody", fieldName)
			return ctx.JSON(http.StatusUnprocessableEntity, validationError)
		}
		var headerParamRawPtr *string = nil
		headerParamRaw := ctx.Request().Header.Get("headerParam")
		_, isheaderParamExists := ctx.Request().Header["headerParam"]
		if !isheaderParamExists {
			// In echo, the ctx.Request().Header["key"] is not 100% reliable, so we need other check, but only if is was not found in the first method
			headerValues := ctx.Request().Header.Values("headerParam")
			isheaderParamExists = len(headerValues) > 0
		}
		if isheaderParamExists {
			headerParam := headerParamRaw
			headerParamRawPtr = &headerParam
		}
		if validatorErr := validatorInstance.Var(headerParamRawPtr, "required"); validatorErr != nil {
			fieldName := "headerParam"
			validationError := wrapValidatorError(validatorErr, "PostWithAllParamsWithBody", fieldName)
			return ctx.JSON(http.StatusUnprocessableEntity, validationError)
		}
		var theBodyRawPtr *Param19theBody.BodyInfo = nil
		conversionErr = bindAndValidateBody(ctx, "application/json", "required", &theBodyRawPtr)
		if conversionErr != nil {
			validationError := external.Rfc7807Error{
				Type: http.StatusText(http.StatusUnprocessableEntity),
				Detail: fmt.Sprintf(
					"A request was made to operation 'PostWithAllParamsWithBody' but body parameter '%s' did not pass validation of '%s' - %s",
					"theBody",
					"BodyInfo",
					validation.ExtractValidationErrorMessage(conversionErr, nil),
				),
				Status:   http.StatusUnprocessableEntity,
				Instance: "/gleece/validation/error/PostWithAllParamsWithBody",
			}
			return ctx.JSON(http.StatusUnprocessableEntity, validationError)
		}
		value, opError := controller.PostWithAllParamsWithBody(*queryParamRawPtr, *headerParamRawPtr, *theBodyRawPtr)
		for key, value := range controller.GetHeaders() {
			ctx.Response().Header().Set(key, value)
		}
		statusCode := getStatusCode(&controller, true, opError)
		if opError != nil {
			stdError := external.Rfc7807Error{
				Type:       http.StatusText(statusCode),
				Detail:     "Encountered an error during operation 'PostWithAllParamsWithBody'",
				Status:     statusCode,
				Instance:   "/gleece/controller/error/PostWithAllParamsWithBody",
				Extensions: map[string]string{"error": opError.Error()},
			}
			return ctx.JSON(statusCode, stdError)
		}
		return ctx.JSON(statusCode, value)
	})
	engine.POST(toEchoUrl("/e2e/post-with-all-params-body-ptr"), func(ctx echo.Context) error {
		authErr := authorize(
			ctx,
			[]SecurityCheckList{
				{
					Relation: SecurityListRelationAnd,
					Checks: []external.SecurityCheck{
						{
							SchemaName: "securitySchemaName",
							Scopes: []string{
								"read",
								"write",
							},
						},
					},
				},
			},
		)
		if authErr != nil {
			return handleAuthorizationError(ctx, authErr, "PostWithAllParamsWithBodyPtr")
		}
		controller := E2EControllerImport.E2EController{}
		controller.InitController(ctx)
		var conversionErr error
		var queryParamRawPtr *string = nil
		queryParamRaw := ctx.QueryParam("queryParam")
		isqueryParamExists := ctx.Request().URL.Query().Has("queryParam")
		if isqueryParamExists {
			queryParam := queryParamRaw
			queryParamRawPtr = &queryParam
		}
		var headerParamRawPtr *string = nil
		headerParamRaw := ctx.Request().Header.Get("headerParam")
		_, isheaderParamExists := ctx.Request().Header["headerParam"]
		if !isheaderParamExists {
			// In echo, the ctx.Request().Header["key"] is not 100% reliable, so we need other check, but only if is was not found in the first method
			headerValues := ctx.Request().Header.Values("headerParam")
			isheaderParamExists = len(headerValues) > 0
		}
		if isheaderParamExists {
			headerParam := headerParamRaw
			headerParamRawPtr = &headerParam
		}
		var theBodyRawPtr *Param24theBody.BodyInfo = nil
		conversionErr = bindAndValidateBody(ctx, "application/json", "", &theBodyRawPtr)
		if conversionErr != nil {
			validationError := external.Rfc7807Error{
				Type: http.StatusText(http.StatusUnprocessableEntity),
				Detail: fmt.Sprintf(
					"A request was made to operation 'PostWithAllParamsWithBodyPtr' but body parameter '%s' did not pass validation of '%s' - %s",
					"theBody",
					"BodyInfo",
					validation.ExtractValidationErrorMessage(conversionErr, nil),
				),
				Status:   http.StatusUnprocessableEntity,
				Instance: "/gleece/validation/error/PostWithAllParamsWithBodyPtr",
			}
			return ctx.JSON(http.StatusUnprocessableEntity, validationError)
		}
		value, opError := controller.PostWithAllParamsWithBodyPtr(queryParamRawPtr, headerParamRawPtr, theBodyRawPtr)
		for key, value := range controller.GetHeaders() {
			ctx.Response().Header().Set(key, value)
		}
		statusCode := getStatusCode(&controller, true, opError)
		if opError != nil {
			stdError := external.Rfc7807Error{
				Type:       http.StatusText(statusCode),
				Detail:     "Encountered an error during operation 'PostWithAllParamsWithBodyPtr'",
				Status:     statusCode,
				Instance:   "/gleece/controller/error/PostWithAllParamsWithBodyPtr",
				Extensions: map[string]string{"error": opError.Error()},
			}
			return ctx.JSON(statusCode, stdError)
		}
		return ctx.JSON(statusCode, value)
	})
	engine.POST(toEchoUrl("/e2e/post-with-all-params-body-required-ptr"), func(ctx echo.Context) error {
		authErr := authorize(
			ctx,
			[]SecurityCheckList{
				{
					Relation: SecurityListRelationAnd,
					Checks: []external.SecurityCheck{
						{
							SchemaName: "securitySchemaName",
							Scopes: []string{
								"read",
								"write",
							},
						},
					},
				},
			},
		)
		if authErr != nil {
			return handleAuthorizationError(ctx, authErr, "PostWithAllParamsWithBodyRequiredPtr")
		}
		controller := E2EControllerImport.E2EController{}
		controller.InitController(ctx)
		var conversionErr error
		var theBodyRawPtr *Param27theBody.BodyInfo = nil
		conversionErr = bindAndValidateBody(ctx, "application/json", "required", &theBodyRawPtr)
		if conversionErr != nil {
			validationError := external.Rfc7807Error{
				Type: http.StatusText(http.StatusUnprocessableEntity),
				Detail: fmt.Sprintf(
					"A request was made to operation 'PostWithAllParamsWithBodyRequiredPtr' but body parameter '%s' did not pass validation of '%s' - %s",
					"theBody",
					"BodyInfo",
					validation.ExtractValidationErrorMessage(conversionErr, nil),
				),
				Status:   http.StatusUnprocessableEntity,
				Instance: "/gleece/validation/error/PostWithAllParamsWithBodyRequiredPtr",
			}
			return ctx.JSON(http.StatusUnprocessableEntity, validationError)
		}
		value, opError := controller.PostWithAllParamsWithBodyRequiredPtr(theBodyRawPtr)
		for key, value := range controller.GetHeaders() {
			ctx.Response().Header().Set(key, value)
		}
		statusCode := getStatusCode(&controller, true, opError)
		if opError != nil {
			stdError := external.Rfc7807Error{
				Type:       http.StatusText(statusCode),
				Detail:     "Encountered an error during operation 'PostWithAllParamsWithBodyRequiredPtr'",
				Status:     statusCode,
				Instance:   "/gleece/controller/error/PostWithAllParamsWithBodyRequiredPtr",
				Extensions: map[string]string{"error": opError.Error()},
			}
			return ctx.JSON(statusCode, stdError)
		}
		return ctx.JSON(statusCode, value)
	})
}
