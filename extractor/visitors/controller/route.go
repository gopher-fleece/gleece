package controller

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
)

func (v *ControllerVisitor) visitMethod(funcDecl *ast.FuncDecl) (definitions.RouteMetadata, bool, error) {
	v.enter(fmt.Sprintf("Method '%s'", funcDecl.Name.Name))
	defer v.exit()

	// Check whether there are any comments on the method - we expect all API endpoints to contain comments.
	// No comments - not an API endpoint.
	if funcDecl.Doc == nil || funcDecl.Doc.List == nil || len(funcDecl.Doc.List) <= 0 {
		return definitions.RouteMetadata{}, false, nil
	}

	comments := extractor.MapDocListToStrings(funcDecl.Doc.List)
	attributes, err := extractor.NewAttributeHolder(comments)
	if err != nil {
		return definitions.RouteMetadata{}, false, v.frozenError(err)
	}

	methodAttr := attributes.GetFirst(extractor.AttributeMethod)
	if methodAttr == nil {
		logger.Info("Method '%s' does not have a @Method attribute and will be ignored", funcDecl.Name.Name)
		return definitions.RouteMetadata{}, false, nil
	}

	routePath := attributes.GetFirstPropertyValueOrEmpty(extractor.AttributeRoute)
	if len(routePath) <= 0 {
		logger.Info("Method '%s' does not have an @Route attribute and will be ignored", funcDecl.Name.Name)
		return definitions.RouteMetadata{}, true, nil
	}

	errorResponses, err := v.getErrorResponseMetadata(&attributes)
	if err != nil {
		return definitions.RouteMetadata{}, true, v.frozenError(err)
	}

	security, err := v.getRouteSecurityWithInheritance(attributes)
	if err != nil {
		return definitions.RouteMetadata{}, true, v.frozenError(err)
	}

	if v.config.RoutesConfig.AuthorizationConfig.EnforceSecurityOnAllRoutes && len(security) <= 0 {
		return definitions.RouteMetadata{}, true, v.getFrozenError(
			"'enforceSecurityOnAllRoutes' setting is 'true'' but method '%s' on controller '%s'"+
				"does not have any explicit or implicit (inherited) security attributes",
			funcDecl.Name.Name,
			v.currentController.Name,
		)
	}

	meta := definitions.RouteMetadata{
		OperationId:         funcDecl.Name.Name,
		HttpVerb:            definitions.EnsureValidHttpVerb(methodAttr.Value),
		Description:         attributes.GetDescription(),
		Hiding:              v.getMethodHideOpts(&attributes),
		Deprecation:         v.getDeprecationOpts(&attributes),
		RestMetadata:        definitions.RestMetadata{Path: routePath},
		ErrorResponses:      errorResponses,
		RequestContentType:  definitions.ContentTypeJSON, // Hardcoded for now, should be supported via comments later
		ResponseContentType: definitions.ContentTypeJSON, // Hardcoded for now, should be supported via comments later
		Security:            security,
	}

	// Check whether the method is an API endpoint, i.e., has all the relevant metadata.
	// Methods without expected metadata are ignored (may switch to raising an error instead)
	isApiEndpoint := len(meta.HttpVerb) > 0 && len(meta.RestMetadata.Path) > 0
	if !isApiEndpoint {
		return meta, false, nil
	}

	// Retrieve parameter information
	funcParams, err := v.getFuncParams(funcDecl, comments)
	if err != nil {
		return meta, true, v.frozenError(err)
	}
	meta.FuncParams = funcParams

	// Set the function's return types
	responses, err := v.getFuncReturnValue(funcDecl)
	if err != nil {
		return meta, true, v.frozenError(err)
	}
	meta.Responses = responses
	meta.HasReturnValue = len(responses) > 1

	successResponseCode, successDescription, err := v.getResponseStatusCodeAndDescription(&attributes, meta.HasReturnValue)
	if err != nil {
		return meta, true, v.frozenError(err)
	}
	meta.ResponseSuccessCode = successResponseCode
	meta.ResponseDescription = successDescription

	return meta, isApiEndpoint, nil
}

func (v *ControllerVisitor) getFuncParams(funcDecl *ast.FuncDecl, comments []string) ([]definitions.FuncParam, error) {
	v.enter("")
	defer v.exit()

	funcParams := []definitions.FuncParam{}

	paramTypes, err := extractor.GetFuncParameterTypeList(v.currentSourceFile, v.fileSet, v.packages, funcDecl)
	if err != nil {
		return funcParams, err
	}

	for _, param := range paramTypes {
		// Record state for diagnostics
		v.enter(fmt.Sprintf("Param %s", param.Name))
		defer v.exit()

		holder, _ := extractor.NewAttributeHolder(comments)
		paramAttrib := holder.FindFirstByValue(param.Name)
		if paramAttrib == nil {
			return funcParams, v.getFrozenError("parameter '%s' does not have a matching documentation attribute", param.Name)
		}

		castValidator, err := extractor.GetCastProperty[string](paramAttrib, extractor.PropertyValidatorString)
		if err != nil {
			return funcParams, v.frozenError(err)
		}

		validatorString := ""
		if castValidator != nil && len(*castValidator) > 0 {
			validatorString = *castValidator
		}

		castName, err := extractor.GetCastProperty[string](paramAttrib, extractor.PropertyName)
		if err != nil {
			return funcParams, v.frozenError(err)
		}

		nameString := param.Name
		if castName != nil && len(*castName) > 0 {
			nameString = *castName
		}

		finalParamMeta := definitions.FuncParam{
			NameInSchema:       nameString,
			ParamMeta:          param,
			Description:        paramAttrib.Description,
			Validator:          appendParamRequiredValidation(&validatorString),
			UniqueImportSerial: v.getNextImportId(),
		}

		// Currently, only body param can be an object type
		switch strings.ToLower(paramAttrib.Name) {
		case "query":
			finalParamMeta.PassedIn = definitions.PassedInQuery
			if !finalParamMeta.TypeMeta.IsUniverseType {
				return funcParams, v.frozenError(createInvalidParamUsageError(finalParamMeta))
			}
		case "header":
			finalParamMeta.PassedIn = definitions.PassedInHeader
			if !finalParamMeta.TypeMeta.IsUniverseType {
				return funcParams, v.frozenError(createInvalidParamUsageError(finalParamMeta))
			}
		case "path":
			finalParamMeta.PassedIn = definitions.PassedInPath
			if !finalParamMeta.TypeMeta.IsUniverseType {
				return funcParams, v.frozenError(createInvalidParamUsageError(finalParamMeta))
			}
		case "body":
			finalParamMeta.PassedIn = definitions.PassedInBody
		}

		funcParams = append(funcParams, finalParamMeta)
	}

	return funcParams, nil
}

func (v *ControllerVisitor) getFuncReturnValue(funcDecl *ast.FuncDecl) ([]definitions.FuncReturnValue, error) {
	v.enter("")
	defer v.exit()

	values := []definitions.FuncReturnValue{}
	var errorRetTypeIndex int

	returnTypes, err := extractor.GetFuncReturnTypeList(v.currentSourceFile, v.fileSet, v.packages, funcDecl)
	if err != nil {
		return values, err
	}

	// Note that controller methods must return and error or (any, error)

	switch len(returnTypes) {
	case 2:
		// If the method returns a 2-tuple, the error is expected to be the second value in the tuple
		errorRetTypeIndex = 1
	case 1:
		// If the method returns a single value, its expected to be an error
		errorRetTypeIndex = 0
	case 0:
		return values, v.getFrozenError("expected method to return an error or a value and error tuple but found void")
	default:
		typeNames := []string{}
		for _, typeMeta := range returnTypes {
			typeNames = append(typeNames, typeMeta.Name)
		}
		return values, v.getFrozenError(
			"expected method to return an error or a value and error tuple but found (%s)",
			strings.Join(typeNames, ", "),
		)
	}

	// Validate whether the method returns a proper error. This may be the first or second return type in the list
	retType := returnTypes[errorRetTypeIndex]
	isValidError, err := v.isAnErrorEmbeddingType(retType)
	if err != nil {
		return values, v.frozenError(err)
	}

	if !isValidError {
		return values, v.getFrozenError("return type '%s' expected to be an error or directly embed it", retType.Name)
	}

	for _, value := range returnTypes {
		values = append(
			values,
			definitions.FuncReturnValue{
				TypeMetadata:       value,
				UniqueImportSerial: v.getNextImportId(),
			},
		)
	}

	return values, nil
}
