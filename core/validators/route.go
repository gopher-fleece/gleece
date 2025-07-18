package validators

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/arbitrators"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/definitions"
)

func ValidateRoute(
	gleeceConfig *definitions.GleeceConfig,
	packagesFacade *arbitrators.PackagesFacade,
	route definitions.RouteMetadata,
) []error {

	var errorCollector common.Collector[error]

	validateParams(errorCollector, route.FuncParams)
	errorCollector.AddIfNotZero(validateReturnTypes(packagesFacade, route.Responses))
	errorCollector.AddIfNotZero(validateSecurity(gleeceConfig, route))

	return errorCollector.Items()
}

func validateParams(errorCollector common.Collector[error], params []definitions.FuncParam) {
	var processedParams []definitions.FuncParam

	for _, param := range params {
		if param.IsContext {
			continue
		}

		switch param.PassedIn {
		case definitions.PassedInBody:
			errorCollector.AddIfNotZero(validateBodyParam(param))
		default:
			errorCollector.AddIfNotZero(validatePrimitiveParam(param))
		}

		errorCollector.AddIfNotZero(validateParamsCombinations(processedParams, param.PassedIn))
		processedParams = append(processedParams, param)
	}
}

func validateBodyParam(param definitions.FuncParam) error {
	// Verify the body is a struct
	if param.TypeMeta.SymbolKind != common.SymKindStruct {
		return fmt.Errorf(
			"body parameters must be structs but '%s' (schema name '%s', type '%s') is of kind '%s'",
			param.Name,
			param.NameInSchema,
			param.TypeMeta.Name,
			param.TypeMeta.SymbolKind,
		)
	}

	return nil
}

func validatePrimitiveParam(param definitions.FuncParam) error {
	isErrType := param.TypeMeta.PkgPath == "" && param.TypeMeta.Name == "error"
	isMapType := param.TypeMeta.PkgPath == "" && strings.HasPrefix(param.TypeMeta.Name, "map[")
	isAliasType := param.TypeMeta.SymbolKind == common.SymKindAlias
	if (!param.TypeMeta.IsUniverseType && !isAliasType) || isErrType || isMapType {
		return fmt.Errorf(
			"header, path and query parameters are currently limited to primitives only but "+
				"%s parameter '%s' (schema name '%s', type '%s') is of kind '%s'",
			param.PassedIn,
			param.Name,
			param.NameInSchema,
			param.TypeMeta.Name,
			param.TypeMeta.SymbolKind,
		)
	}

	return nil
}

// This function is deprecated - no need to test here, all validation moved to the NewAnnotationHolder logic
func validateParamsCombinations(funcParams []definitions.FuncParam, newParamType definitions.ParamPassedIn) error {

	isBodyParamAlreadyExists := slices.ContainsFunc(funcParams, func(p definitions.FuncParam) bool {
		return p.PassedIn == definitions.PassedInBody
	})

	isFormParamAlreadyExists := slices.ContainsFunc(funcParams, func(p definitions.FuncParam) bool {
		return p.PassedIn == definitions.PassedInForm
	})

	// Body is a special case, only one body parameter is allowed per route
	if newParamType == definitions.PassedInBody && isBodyParamAlreadyExists {
		return fmt.Errorf("body parameter is invalid, only one body per route is allowed")
	}

	// Form is an implementation of url encoded string in the body, thus it cannot be used if the body is already in use
	if newParamType == definitions.PassedInBody && isFormParamAlreadyExists {
		return fmt.Errorf("body parameter is invalid, using body is not allowed when a form is in use")
	}

	// Form is an implementation of url encoded string in the body, thus it cannot be used if the body is already in use
	if newParamType == definitions.PassedInForm && isBodyParamAlreadyExists {
		return fmt.Errorf("form parameter is invalid, using form is not allowed when a body is in use")
	}
	return nil
}

func validateReturnTypes(packagesFacade *arbitrators.PackagesFacade, funcRetTypes []definitions.FuncReturnValue) error {
	// Note that controller methods must return and error or (any, error)

	var errorRetTypeIndex int

	switch len(funcRetTypes) {
	case 2:
		// If the method returns a 2-tuple, the error is expected to be the second value in the tuple
		errorRetTypeIndex = 1
	case 1:
		// If the method returns a single value, its expected to be an error
		errorRetTypeIndex = 0
	case 0:
		return fmt.Errorf("expected method to return an error or a value and error tuple but found void")
	default:
		typeNames := []string{}
		for _, typeMeta := range funcRetTypes {
			typeNames = append(typeNames, typeMeta.Name)
		}

		return fmt.Errorf(
			"expected method to return an error or a value and error tuple but found (%s)",
			strings.Join(typeNames, ", "),
		)
	}

	// Validate whether the method returns a proper error. This may be the first or second return type in the list
	retType := funcRetTypes[errorRetTypeIndex]
	relevantPkg, err := packagesFacade.GetPackage(retType.PkgPath)
	if err != nil {
		return err
	}

	isValidError, err := metadata.IsAnErrorEmbeddingType(retType.TypeMetadata, relevantPkg)
	if err != nil {
		return err
	}

	if !isValidError {
		return fmt.Errorf("return type '%s' expected to be an error or directly embed it", retType.Name)
	}

	return nil
}

func validateSecurity(
	gleeceConfig *definitions.GleeceConfig,
	route definitions.RouteMetadata,
) error {
	if gleeceConfig != nil && gleeceConfig.RoutesConfig.AuthorizationConfig.EnforceSecurityOnAllRoutes && len(route.Security) <= 0 {
		return fmt.Errorf(
			"'enforceSecurityOnAllRoutes' setting is 'true'' but route with operation ID '%s' "+
				"does not have any explicit or implicit (inherited) security attributes",
			route.OperationId,
		)
	}
	return nil
}
