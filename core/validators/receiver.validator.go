package validators

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gopher-fleece/gleece/v2/common"
	"github.com/gopher-fleece/gleece/v2/core/arbitrators"
	"github.com/gopher-fleece/gleece/v2/core/metadata"
	"github.com/gopher-fleece/gleece/v2/core/validators/diagnostics"
	"github.com/gopher-fleece/gleece/v2/definitions"
)

type funcParamEx struct {
	metadata.FuncParam
	PassedIn definitions.ParamPassedIn
}

type ReceiverValidator struct {
	CommonValidator
	gleeceConfig     *definitions.GleeceConfig
	packagesFacade   *arbitrators.PackagesFacade
	parentController *metadata.ControllerMeta
	receiver         *metadata.ReceiverMeta
}

func NewReceiverValidator(
	gleeceConfig *definitions.GleeceConfig,
	packagesFacade *arbitrators.PackagesFacade,
	controller *metadata.ControllerMeta,
	receiver *metadata.ReceiverMeta,
) ReceiverValidator {
	return ReceiverValidator{
		CommonValidator: CommonValidator{
			holder: receiver.Annotations,
		},
		gleeceConfig:     gleeceConfig,
		packagesFacade:   packagesFacade,
		parentController: controller,
		receiver:         receiver,
	}
}

func (v ReceiverValidator) Validate() (diagnostics.EntityDiagnostic, error) {
	receiverDiag := diagnostics.NewEntityDiagnostic("Receiver", v.receiver.Name)

	// Run base/common validations first
	receiverDiag.AddDiagnostics(v.CommonValidator.Validate())

	paramDiags, err := v.validateParams(v.receiver)
	if err != nil {
		return receiverDiag, fmt.Errorf("could not validate parameter for receiver '%s' - %w", v.receiver.Name, err)
	}

	receiverDiag.AddDiagnostics(paramDiags)

	retTypeDiag, err := v.validateReturnTypes(v.receiver)
	if err != nil {
		return receiverDiag, fmt.Errorf("could not validate return types for receiver '%s' - %w", v.receiver.Name, err)
	}
	receiverDiag.AddDiagnosticIfNotNil(retTypeDiag)

	secDiag, err := v.validateSecurity(v.receiver)
	if err != nil {
		return receiverDiag, fmt.Errorf("could not validate security for receiver '%s' - %w", v.receiver.Name, err)
	}
	receiverDiag.AddDiagnosticIfNotNil(secDiag)

	linkValidator, err := NewAnnotationLinkValidator(v.receiver)
	if err != nil {
		return receiverDiag, fmt.Errorf("failed to construct an annotation link validator - %v", err)
	}

	receiverDiag.AddDiagnostics(linkValidator.Validate())

	return receiverDiag, nil
}

func (v ReceiverValidator) validateParams(receiver *metadata.ReceiverMeta) ([]diagnostics.ResolvedDiagnostic, error) {
	var processedParams []funcParamEx
	diags := []diagnostics.ResolvedDiagnostic{}

	for _, param := range receiver.Params {
		if param.Type.IsContext() {
			continue
		}

		passedIn, err := v.getPassedInValue(param)
		if err != nil {
			return diags, err
		}

		if passedIn == nil {
			// If we couldn't process the passedIn portion, no reason in continuing.
			// An error or error diagnostic will have been added, at this point
			// or later on by the annotation link validator.
			continue
		}

		switch *passedIn {
		case definitions.PassedInBody:
			diags = common.AppendIfNotNil(diags, v.validateBodyParam(receiver, param))
		default:
			diags = common.AppendIfNotNil(diags, v.validateNonBodyParam(receiver, param, *passedIn))
		}

		diags = common.AppendIfNotNil(diags, v.validateParamsCombinations(processedParams, param, *passedIn))

		processedParams = append(processedParams, funcParamEx{FuncParam: param, PassedIn: *passedIn})
	}

	return diags, nil
}

func (v ReceiverValidator) getPassedInValue(param metadata.FuncParam) (*definitions.ParamPassedIn, error) {
	// This function gets the parameter's passed-in value (e.g. passed-in-body or passed-in-header)
	// If it fails, it may return a standard error or an InvalidAnnotation error.
	passedIn, err := metadata.GetParamPassedIn(param.Name, param.Annotations)
	if err == nil {
		return &passedIn, nil
	}

	// If we didn't get a value, it generally means a missing annotation which is a 'diagnostic'
	// or an outright malformed one which we consider an 'error'.
	// InvalidAnnotation error is the former.
	// In such a case, we return a nil here to halt further checks.
	//
	// The relevant diagnostic will be emitted by a subsequent call to validators.AnnotationLinkValidator
	if _, isInvalidAnnotationErr := err.(metadata.InvalidAnnotationError); isInvalidAnnotationErr {
		return nil, nil
	}

	// A true error or a grossly malformed annotation. Regardless, this is a flow-terminating error.
	return nil, fmt.Errorf(
		"failed to determine 'passed-in' type for parameter '%s' in receiver '%s' - %w",
		param.Name,
		v.receiver.Name,
		err,
	)
}

func (v ReceiverValidator) validateBodyParam(
	receiver *metadata.ReceiverMeta,
	param metadata.FuncParam,
) *diagnostics.ResolvedDiagnostic {
	// Currently, body parameters cannot be a non-array/slice primitive/built-in special
	if param.Type.SymbolKind.IsBuiltin() && !param.Type.IsIterable() {
		diag := diagnostics.NewErrorDiagnostic(
			receiver.Annotations.FileName(),
			fmt.Sprintf(
				"body parameter '%s' (schema name '%s', type '%s') is a built-in primitive or special (e.g. time.Time) which is not allowed",
				param.Name,
				getParamSchemaNameOrFallback(param, "unknown"),
				param.Type.Name,
			),
			diagnostics.DiagReceiverInvalidBody,
			param.Range,
		)
		return &diag
	}

	return nil
}

func (v ReceiverValidator) validateNonBodyParam(
	receiver *metadata.ReceiverMeta,
	param metadata.FuncParam,
	passedIn definitions.ParamPassedIn,
) *diagnostics.ResolvedDiagnostic {
	// First - any arrays in anything other than query
	// (remember that this validates non-body parameters)
	// is automatically invalid - Neither URL parameters
	if param.Type.IsIterable() && passedIn != definitions.PassedInQuery && passedIn != definitions.PassedInBody {
		diag := diagnostics.NewErrorDiagnostic(
			receiver.Annotations.FileName(),
			fmt.Sprintf(
				"parameter '%s' (schema name '%s', type '%s') is an array/slice and can only be passed in a query or a body",
				param.Name,
				getParamSchemaNameOrFallback(param, "unknown"),
				param.Type.Name,
			),
			diagnostics.DiagReceiverParamNotPrimitive,
			param.Range,
		)
		return &diag
	}

	isErrType := param.Type.PkgPath == "" && param.Type.Name == "error"
	isMapType := param.Type.PkgPath == "" && strings.HasPrefix(param.Type.Name, "map[")
	isAnEnum := param.Type.SymbolKind == common.SymKindEnum

	isAnAlias, isAPrimitiveAlias := isPrimitiveAlias(param)

	if (param.Type.IsUniverseType() || isAnEnum || (isAnAlias && isAPrimitiveAlias)) && !isErrType && !isMapType {
		return nil
	}

	isIterableMsg := ""
	if param.Type.IsIterable() {
		isIterableMsg = "an iterable "
	}

	diag := diagnostics.NewErrorDiagnostic(
		receiver.Annotations.FileName(),
		fmt.Sprintf(
			"header/path/query parameters may only be primitives but "+
				"%s parameter '%s' (schema name '%s', type '%s') is %sof kind '%s'",
			passedIn,
			param.Name,
			getParamSchemaNameOrFallback(param, "unknown"),
			param.Type.Name,
			isIterableMsg,
			param.Type.SymbolKind,
		),
		diagnostics.DiagReceiverParamNotPrimitive,
		param.Range,
	)

	return &diag
}

// This function is deprecated - no need to test here, all validation moved to the NewAnnotationHolder logic
func (v ReceiverValidator) validateParamsCombinations(
	funcParams []funcParamEx,
	newParam metadata.FuncParam,
	newParamType definitions.ParamPassedIn,
) *diagnostics.ResolvedDiagnostic {

	doesBodyParamAlreadyExists := slices.ContainsFunc(funcParams, func(p funcParamEx) bool {
		return p.PassedIn == definitions.PassedInBody
	})

	doesFormParamAlreadyExists := slices.ContainsFunc(funcParams, func(p funcParamEx) bool {
		return p.PassedIn == definitions.PassedInForm
	})

	var errMsg string

	switch newParamType {
	case definitions.PassedInBody:
		if doesBodyParamAlreadyExists {
			// Body is a special case, only one body parameter is allowed per route
			errMsg = "Body parameter is invalid, only one body per route is allowed"
		} else if doesFormParamAlreadyExists {
			// Form is an implementation of url encoded string in the body, thus it cannot be used if the body is already in use
			errMsg = "Body parameter is invalid, using body is not allowed when a form is in use"
		}
	case definitions.PassedInForm:
		if doesBodyParamAlreadyExists {
			// Form is an implementation of url encoded string in the body, thus it cannot be used if the body is already in use
			errMsg = "Form parameter is invalid, using form is not allowed when a body is in use"
		}
	}

	if errMsg != "" {
		diag := diagnostics.NewErrorDiagnostic(
			newParam.FVersion.Path,
			errMsg,
			diagnostics.DiagReceiverRetValsInvalidSignature,
			newParam.Range,
		)
		return &diag
	}

	return nil
}

func (v ReceiverValidator) validateReturnTypes(receiver *metadata.ReceiverMeta) (*diagnostics.ResolvedDiagnostic, error) {
	// Validate return sig first
	errorRetTypeIndex, diagErr := getDiagForRetSig(receiver)
	if diagErr != nil {
		return diagErr, nil
	}

	// Validate whether the method returns a proper error. This may be the first or second return type in the list
	retType := receiver.RetVals[errorRetTypeIndex]
	relevantPkg, err := v.packagesFacade.GetPackage(retType.PkgPath)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to obtain package object for return value '%s' in receiver '%s' - %w",
			retType.Type.Name,
			receiver.Name,
			err,
		)
	}

	isValidError, err := metadata.IsAnErrorEmbeddingTypeUsage(retType.Type, relevantPkg)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to determine whether return value '%s' in receiver '%s' embeds an error - %w",
			retType.Name,
			receiver.Name,
			err,
		)
	}

	if isValidError {
		return nil, nil
	}

	// 'error' return value is not actually an error
	return common.Ptr(
		diagnostics.NewErrorDiagnostic(
			receiver.Annotations.FileName(),
			fmt.Sprintf(
				"return type '%s' in receiver '%s' expected to be an error or directly embed it",
				retType.Name,
				receiver.Name,
			),
			diagnostics.DiagReceiverRetValsIsNotError,
			receiver.RetValsRange(),
		),
	), nil
}

func (v ReceiverValidator) validateSecurity(receiver *metadata.ReceiverMeta) (*diagnostics.ResolvedDiagnostic, error) {
	// First, check if we're enforcing security on all routes
	if v.gleeceConfig == nil || !v.gleeceConfig.RoutesConfig.AuthorizationConfig.EnforceSecurityOnAllRoutes {
		return nil, nil
	}

	// Get any security defined on the parent controller
	controllerSec, err := metadata.GetSecurityFromContext(v.parentController.Struct.Annotations)
	if err != nil {
		return nil, fmt.Errorf(
			"could not determine controller '%s' security for while processing receiver '%s' - %w",
			v.parentController.Struct.Name,
			receiver.Name,
			err,
		)
	}

	// Get the security for the route - either explicit or inherited from parent controller
	security, err := metadata.GetRouteSecurityWithInheritance(receiver.Annotations, controllerSec)
	if err != nil {
		return nil, fmt.Errorf(
			"could not determine route security for receiver '%s' - %w",
			receiver.Name,
			err,
		)
	}

	if len(security) > 0 {
		// An explicit/inherited security exists
		return nil, nil
	}

	// Look for a default security in the config
	security = metadata.GetDefaultSecurity(v.gleeceConfig)
	if len(security) > 0 {
		return nil, nil
	}

	// Finally, emit an error diagnostic - the receiver has no security
	diag := diagnostics.NewErrorDiagnostic(
		receiver.Annotations.FileName(),
		fmt.Sprintf(
			"'enforceSecurityOnAllRoutes' setting is 'true'' but route with operation ID '%s' "+
				"does not have any explicit or implicit (inherited) security attributes",
			receiver.Name,
		),
		diagnostics.DiagReceiverMissingSecurity,
		receiver.RetValsRange(),
	)

	return &diag, nil
}

func getDiagForRetSig(receiver *metadata.ReceiverMeta) (int, *diagnostics.ResolvedDiagnostic) {
	// Note that controller methods must return and error or (any, error)
	var errorRetTypeIndex int

	switch len(receiver.RetVals) {
	case 2:
		// If the method returns a 2-tuple, the error is expected to be the second value in the tuple
		errorRetTypeIndex = 1
	case 1:
		// If the method returns a single value, its expected to be an error
		errorRetTypeIndex = 0
	case 0:
		return errorRetTypeIndex, common.Ptr(
			diagnostics.NewErrorDiagnostic(
				receiver.Annotations.FileName(),
				"Expected method to return an error or a value and error tuple but found void",
				diagnostics.DiagReceiverRetValsInvalidSignature,
				receiver.RetValsRange(),
			),
		)
	default:
		typeNames := []string{}
		for _, typeMeta := range receiver.RetVals {
			typeNames = append(typeNames, typeMeta.Name)
		}

		return errorRetTypeIndex, common.Ptr(
			diagnostics.NewErrorDiagnostic(
				receiver.Annotations.FileName(),
				fmt.Sprintf(
					"Expected method to return an error or a value and error tuple but found (%s)",
					strings.Join(typeNames, ", "),
				),
				diagnostics.DiagReceiverRetValsInvalidSignature,
				receiver.RetValsRange(),
			),
		)
	}

	return errorRetTypeIndex, nil
}

func getParamSchemaNameOrFallback(param metadata.FuncParam, fallback string) string {
	nameInSchema, err := metadata.GetParameterSchemaName(param.Name, param.Annotations)
	if err != nil {
		return fallback
	}
	return nameInSchema
}

func isPrimitiveAlias(param metadata.FuncParam) (bool, bool) {
	if param.Type.SymbolKind != common.SymKindAlias {
		return false, false
	}

	flattenedTypeRef := param.Type.Root.Flatten()

	switch len(flattenedTypeRef) {
	case 0:
		return true, true
	case 1:
		return true, flattenedTypeRef[0].Kind() == metadata.TypeRefKindNamed
	default:
		return true, flattenedTypeRef[len(flattenedTypeRef)-1].Kind() == metadata.TypeRefKindNamed
	}
}
