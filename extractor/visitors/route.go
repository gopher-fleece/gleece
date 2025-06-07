package visitors

import (
	"fmt"
	"go/ast"
	"slices"
	"strings"

	MapSet "github.com/deckarep/golang-set/v2"
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/extractor/arbitrators"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/runtime"
)

type RouteVisitor struct {
	BaseVisitor

	// The file currently being worked on
	currentSourceFile *ast.File

	parentController *definitions.ControllerMetadata
	gleeceConfig     *definitions.GleeceConfig
}

func NewRouteVisitor(
	context *VisitContext,
	parentController *definitions.ControllerMetadata,
) (*RouteVisitor, error) {
	visitor := RouteVisitor{parentController: parentController}

	err := visitor.initializeWithArbitrationProvider(context)
	return &visitor, err
}

// visitMethod Visits a controller route given as a FuncDecl and returns its metadata and whether it is an API endpoint
func (v *RouteVisitor) VisitMethod(funcDecl *ast.FuncDecl, sourceFile *ast.File) (definitions.RouteMetadata, bool, error) {
	v.enter(fmt.Sprintf("Method '%s'", funcDecl.Name.Name))
	defer v.exit()

	// Sets the context for the visit
	v.currentSourceFile = sourceFile

	// Check whether there are any comments on the method - we expect all API endpoints to contain comments.
	// No comments - not an API endpoint.
	if funcDecl.Doc == nil || funcDecl.Doc.List == nil || len(funcDecl.Doc.List) <= 0 {
		return definitions.RouteMetadata{}, false, nil
	}

	comments := gast.MapDocListToStrings(funcDecl.Doc.List)
	attributes, err := annotations.NewAnnotationHolder(comments, annotations.CommentSourceRoute)
	if err != nil {
		// Couldn't read comments. Fail.
		return definitions.RouteMetadata{}, false, v.frozenError(err)
	}

	methodAttr := attributes.GetFirst(annotations.AttributeMethod)
	if methodAttr == nil {
		logger.Info("Method '%s' does not have a @Method attribute and will be ignored", funcDecl.Name.Name)
		return definitions.RouteMetadata{}, false, nil
	}

	routePath := attributes.GetFirstValueOrEmpty(annotations.AttributeRoute)
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

	// If the route does not have any security whatsoever and the `EnforceSecurityOnAllRoutes` setting is true, we must fail here.
	//
	// This is to prevent cases where a developer forgets to declare security for a controller/route.
	if v.gleeceConfig != nil && v.gleeceConfig.RoutesConfig.AuthorizationConfig.EnforceSecurityOnAllRoutes && len(security) <= 0 {
		controllerName := "N/A"
		if v.parentController != nil {
			controllerName = v.parentController.Name
		}

		return definitions.RouteMetadata{}, true, v.getFrozenError(
			"'enforceSecurityOnAllRoutes' setting is 'true'' but method '%s' on controller '%s'"+
				"does not have any explicit or implicit (inherited) security attributes",
			funcDecl.Name.Name,
			controllerName,
		)
	}

	// Template context is optional, additional information that can be accessed at the template level.
	// This allows users to perform deep customizations on a per-route basis.
	templateContext, err := getTemplateContextMetadata(&attributes)
	if err != nil {
		return definitions.RouteMetadata{}, true, v.frozenError(err)
	}

	fVer, err := gast.NewFileVersionFromAstFile(v.currentSourceFile, v.context.ArbitrationProvider.FileSet())
	if err != nil {
		return definitions.RouteMetadata{}, true, v.frozenError(err)
	}

	meta := definitions.RouteMetadata{
		OperationId:         funcDecl.Name.Name,
		HttpVerb:            definitions.EnsureValidHttpVerb(methodAttr.Value),
		Description:         attributes.GetDescription(),
		Hiding:              getMethodHideOpts(&attributes),
		Deprecation:         getDeprecationOpts(&attributes),
		RestMetadata:        definitions.RestMetadata{Path: routePath},
		ErrorResponses:      errorResponses,
		RequestContentType:  definitions.ContentTypeJSON, // Hardcoded for now, should be supported via annotations later on
		ResponseContentType: definitions.ContentTypeJSON, // Hardcoded for now, should be supported via annotations later on
		Security:            security,
		TemplateContext:     templateContext,
		FVersion:            &fVer,
	}

	// Check whether the method is an API endpoint, i.e., has all the relevant metadata.
	// Methods without expected metadata are ignored (may switch to raising an error instead)
	isApiEndpoint := len(meta.HttpVerb) > 0 && len(meta.RestMetadata.Path) > 0
	if !isApiEndpoint {
		return meta, false, nil
	}

	// Retrieve parameter information
	funcParams, err := v.getValidatedFuncParams(funcDecl, comments)
	if err != nil {
		return meta, true, v.frozenError(err)
	}

	meta.FuncParams = []definitions.FuncParam{}
	for _, fParam := range funcParams {
		meta.FuncParams = append(meta.FuncParams, fParam.Reduce())
	}

	// Set the function's return types
	returnValues, err := v.getFuncReturnValue(funcDecl)
	if err != nil {
		return meta, true, v.frozenError(err)
	}

	meta.Responses = []definitions.FuncReturnValue{}
	for _, fRetVal := range returnValues {
		meta.Responses = append(meta.Responses, fRetVal.Reduce())
	}

	meta.HasReturnValue = len(returnValues) > 1

	successResponseCode, successDescription, err := getResponseStatusCodeAndDescription(&attributes, meta.HasReturnValue)
	if err != nil {
		return meta, true, v.frozenError(err)
	}
	meta.ResponseSuccessCode = successResponseCode
	meta.ResponseDescription = successDescription

	return meta, isApiEndpoint, nil
}

func (v *RouteVisitor) getValidatedFuncParams(funcDecl *ast.FuncDecl, comments []string) ([]arbitrators.FuncParamWithAst, error) {
	funcParams, err := v.getFuncParams(funcDecl, comments)
	if err != nil {
		return funcParams, v.frozenError(err)
	}

	for _, param := range funcParams {
		if param.IsContext {
			// Context parameters do not require any validation
			continue
		}

		var validationErr error

		switch param.PassedIn {
		case definitions.PassedInBody:
			validationErr = v.validateBodyParam(param)
		default:
			validationErr = v.validatePrimitiveParam(param)
		}

		if validationErr != nil {
			return funcParams, v.frozenError(validationErr)
		}
	}

	return funcParams, nil
}

func (v *RouteVisitor) validateBodyParam(param arbitrators.FuncParamWithAst) error {
	// Verify the body is a struct
	if param.ParamMetaWithAst.TypeMetadata.SymbolKind != common.SymKindStruct {
		return v.getFrozenError(
			"body parameters must be structs but '%s' (schema name '%s', type '%s') is of kind '%s'",
			param.Name,
			param.NameInSchema,
			param.TypeMetadata.Name,
			param.TypeMetadata.SymbolKind,
		)
	}

	return nil
}

func (v *RouteVisitor) validatePrimitiveParam(param arbitrators.FuncParamWithAst) error {
	// Currently, we're limited to primitive header, path and query parameters.
	// This is a simple and silly check for those.
	// need to fully integrate the SymbolKind field..
	isErrType := param.TypeMetadata.PkgPath == "" && param.TypeMetadata.Name == "error"
	isMapType := param.TypeMetadata.PkgPath == "" && strings.HasPrefix(param.TypeMetadata.Name, "map[")
	isAliasType := param.TypeMetadata.SymbolKind == common.SymKindAlias
	if (!param.TypeMetadata.IsUniverseType && !isAliasType) || isErrType || isMapType {
		return v.getFrozenError(
			"header, path and query parameters are currently limited to primitives only but "+
				"%s parameter '%s' (schema name '%s', type '%s') is of kind '%s'",
			param.PassedIn,
			param.Name,
			param.NameInSchema,
			param.TypeMetadata.Name,
			param.TypeMetadata.SymbolKind,
		)
	}

	return nil
}

// This function is deprecated - no need to test here, all validation moved to the NewAnnotationHolder logic
func (v *RouteVisitor) validateParamsCombinations(funcParams []arbitrators.FuncParamWithAst, newParamType definitions.ParamPassedIn) error {

	isBodyParamAlreadyExists := slices.ContainsFunc(funcParams, func(p arbitrators.FuncParamWithAst) bool {
		return p.PassedIn == definitions.PassedInBody
	})

	isFormParamAlreadyExists := slices.ContainsFunc(funcParams, func(p arbitrators.FuncParamWithAst) bool {
		return p.PassedIn == definitions.PassedInForm
	})

	// Body is a special case, only one body parameter is allowed per route
	if newParamType == definitions.PassedInBody && isBodyParamAlreadyExists {
		return v.getFrozenError("body parameter is invalid, only one body per route is allowed")
	}

	// Form is an implementation of url encoded string in the body, thus it cannot be used if the body is already in use
	if newParamType == definitions.PassedInBody && isFormParamAlreadyExists {
		return v.getFrozenError("body parameter is invalid, using body is not allowed when a form is in use")
	}

	// Form is an implementation of url encoded string in the body, thus it cannot be used if the body is already in use
	if newParamType == definitions.PassedInForm && isBodyParamAlreadyExists {
		return v.getFrozenError("form parameter is invalid, using form is not allowed when a body is in use")
	}
	return nil
}

func (v *RouteVisitor) getFuncParams(funcDecl *ast.FuncDecl, comments []string) ([]arbitrators.FuncParamWithAst, error) {
	v.enter("")
	defer v.exit()

	funcParams := []arbitrators.FuncParamWithAst{}

	paramTypes, err := v.context.ArbitrationProvider.Ast().GetFuncParameterTypeList(v.currentSourceFile, funcDecl)
	if err != nil {
		return funcParams, err
	}

	holder, err := annotations.NewAnnotationHolder(comments, annotations.CommentSourceRoute)
	if err != nil {
		return funcParams, err
	}

	for _, param := range paramTypes {
		// Record state for diagnostics
		v.enter(fmt.Sprintf("Param %s", param.Name))
		defer v.exit()

		var finalParamMeta arbitrators.FuncParamWithAst

		if param.IsContext {
			// Special handling for contexts - a context parameter doesn't have nor need much of the metadata or validations
			finalParamMeta = arbitrators.FuncParamWithAst{ParamMetaWithAst: param.ParamMetaWithAst}
		} else {
			// Normal params do require further processing (comments, validator strings) and validations
			if finalParamMeta, err = v.processFuncParameter(holder, param.ParamMetaWithAst); err != nil {
				return funcParams, v.frozenError(err)
			}

			if err := v.validateParamsCombinations(funcParams, finalParamMeta.PassedIn); err != nil {
				return funcParams, v.frozenError(err)
			}
		}

		funcParams = append(funcParams, finalParamMeta)
	}

	return funcParams, nil
}

func (v *RouteVisitor) processFuncParameter(
	annotationHolder annotations.AnnotationHolder,
	param arbitrators.ParamMetaWithAst,
) (arbitrators.FuncParamWithAst, error) {
	paramAttrib := annotationHolder.FindFirstByValue(param.Name)
	if paramAttrib == nil {
		return arbitrators.FuncParamWithAst{}, v.getFrozenError("parameter '%s' does not have a matching documentation attribute", param.Name)
	}

	castValidator, err := annotations.GetCastProperty[string](paramAttrib, annotations.PropertyValidatorString)
	if err != nil {
		return arbitrators.FuncParamWithAst{}, v.frozenError(err)
	}

	validatorString := ""
	if castValidator != nil && len(*castValidator) > 0 {
		validatorString = *castValidator
	}

	castName, err := annotations.GetCastProperty[string](paramAttrib, annotations.PropertyName)
	if err != nil {
		return arbitrators.FuncParamWithAst{}, v.frozenError(err)
	}

	nameString := param.Name
	if castName != nil && len(*castName) > 0 {
		nameString = *castName
	}

	var paramPassedIn definitions.ParamPassedIn

	// Currently, only body param can be an object type
	switch strings.ToLower(paramAttrib.Name) {
	case "query":
		paramPassedIn = definitions.PassedInQuery
	case "header":
		paramPassedIn = definitions.PassedInHeader
	case "path":
		paramPassedIn = definitions.PassedInPath
	case "body":
		paramPassedIn = definitions.PassedInBody
	case "formfield": // Currently, form fields are the only supported form of form parameters, in the future, a full form object may be supported too
		paramPassedIn = definitions.PassedInForm
	}

	return arbitrators.FuncParamWithAst{
		NameInSchema:       nameString,
		ParamMetaWithAst:   param,
		PassedIn:           paramPassedIn,
		Description:        paramAttrib.Description,
		Validator:          appendParamRequiredValidation(&validatorString, param.TypeMetadata.IsByAddress, paramPassedIn),
		UniqueImportSerial: v.context.SyncedProvider.GetNextImportId(),
	}, nil
}

func (v *RouteVisitor) getFuncReturnValue(funcDecl *ast.FuncDecl) ([]arbitrators.FuncReturnValueWithAst, error) {
	v.enter("")
	defer v.exit()

	values := []arbitrators.FuncReturnValueWithAst{}
	var errorRetTypeIndex int

	returnTypes, err := v.context.ArbitrationProvider.Ast().GetFuncReturnTypeList(v.currentSourceFile, funcDecl)
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
	isValidError, err := isAnErrorEmbeddingType(v.context.ArbitrationProvider.Pkg(), retType.TypeMetadata)
	if err != nil {
		return values, v.frozenError(err)
	}

	if !isValidError {
		return values, v.getFrozenError("return type '%s' expected to be an error or directly embed it", retType.Name)
	}

	for _, value := range returnTypes {
		values = append(
			values,
			arbitrators.FuncReturnValueWithAst{
				TypeMetadataWithAst: value,
				UniqueImportSerial:  v.context.SyncedProvider.GetNextImportId(),
			},
		)
	}

	return values, nil
}

func (v RouteVisitor) getErrorResponseMetadata(attributes *annotations.AnnotationHolder) ([]definitions.ErrorResponse, error) {
	responseAttributes := attributes.GetAll(annotations.AttributeErrorResponse)

	responses := []definitions.ErrorResponse{}
	encounteredCodes := MapSet.NewSet[runtime.HttpStatusCode]()

	for _, attr := range responseAttributes {
		code, err := definitions.ConvertToHttpStatus(attr.Value)
		if err != nil {
			return responses, err
		}

		if encounteredCodes.ContainsOne(code) {
			logger.Warn(
				"Status code '%d' appears multiple time on a controller receiver. Ignoring. Original Comment: %s",
				code,
				attr,
			)
			continue
		}
		responses = append(responses, definitions.ErrorResponse{HttpStatusCode: code, Description: attr.Description})
		encounteredCodes.Add(code)
	}

	return responses, nil
}

// getRouteSecurityWithInheritance Gets the securities to be associated with the route annotated by the given AnnotationHolder.
// Security is hierarchial and uses a 'first-match' approach:
//
// 1. Explicit, receiver level annotations
// 2. Explicit, controller level annotations
// 3. Default securities in Gleece configuration file.
func (v *RouteVisitor) getRouteSecurityWithInheritance(attributes annotations.AnnotationHolder) ([]definitions.RouteSecurity, error) {
	explicitSec, err := getSecurityFromContext(attributes)
	if err != nil {
		return []definitions.RouteSecurity{}, v.frozenError(err)
	}

	if len(explicitSec) > 0 {
		return explicitSec, nil
	}

	if v.parentController != nil {
		return v.parentController.Security, nil
	}

	return []definitions.RouteSecurity{}, nil
}
