package validators

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/arbitrators"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/validators/diagnostics"
	"github.com/gopher-fleece/gleece/definitions"
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

		passedIn, diag, err := v.getPassedInValue(param)
		if err != nil {
			return diags, err
		}

		common.AppendIfNotNil(diags, diag)
		if passedIn == nil {
			// If we couldn't process the passedIn portion, no reason in continuing.
			// An error or error diagnostic will have been added, at this point.
			continue
		}

		switch *passedIn {
		case definitions.PassedInBody:
			common.AppendIfNotNil(diags, v.validateBodyParam(receiver, param))
		default:
			common.AppendIfNotNil(diags, v.validatePrimitiveParam(receiver, param, *passedIn))
		}

		common.AppendIfNotNil(diags, v.validateParamsCombinations(processedParams, param, *passedIn))

		processedParams = append(processedParams, funcParamEx{FuncParam: param, PassedIn: *passedIn})
	}

	return diags, nil
}

func (v ReceiverValidator) getPassedInValue(param metadata.FuncParam) (
	*definitions.ParamPassedIn,
	*diagnostics.ResolvedDiagnostic,
	error,
) {
	// This function gets the parameter's passed-in value (e.g. passed-in-body or passed-in-header)
	// If it fails, it may return a standard error or an InvalidAnnotation error.
	passedIn, err := metadata.GetParamPassedIn(param.Name, param.Annotations)
	if err == nil {
		return &passedIn, nil, nil
	}

	// If we didn't get a value, it generally means a missing annotation which is a 'diagnostic'
	// or an outright malformed one which we consider an 'error'.
	// InvalidAnnotation error is the former.
	if _, isInvalidAnnotationErr := err.(metadata.InvalidAnnotationError); isInvalidAnnotationErr {
		diag := diagnostics.NewErrorDiagnostic(
			v.receiver.FVersion.Path,
			fmt.Sprintf(
				"Parameter '%s' in receiver '%s' is not referenced by any annotation",
				param.Name,
				v.receiver.Name,
			),
			diagnostics.DiagLinkerUnreferencedParameter,
			v.receiver.RetValsRange(),
		)
		return nil, &diag, nil
	}

	// A true error or a grossly malformed annotation. Regardless, this is a flow-terminating error.
	return nil, nil, fmt.Errorf(
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
	// Verify the body is a struct
	if param.SymbolKind != common.SymKindStruct {
		nameInSchema, err := metadata.GetParameterSchemaName(param.Name, param.Annotations)
		if err != nil {
			nameInSchema = "unknown"
		}

		diag := diagnostics.NewErrorDiagnostic(
			receiver.FVersion.Path,
			fmt.Sprintf(
				"body parameters must be structs but '%s' (schema name '%s', type '%s') is of kind '%s'",
				param.Name,
				nameInSchema,
				param.Type.Name,
				param.Type.SymbolKind,
			),
			diagnostics.DiagReceiverInvalidBody,
			param.Range,
		)
		return &diag
	}

	return nil
}

func (v ReceiverValidator) validatePrimitiveParam(
	receiver *metadata.ReceiverMeta,
	param metadata.FuncParam,
	passedIn definitions.ParamPassedIn,
) *diagnostics.ResolvedDiagnostic {
	isErrType := param.Type.PkgPath == "" && param.Type.Name == "error"
	isMapType := param.Type.PkgPath == "" && strings.HasPrefix(param.Type.Name, "map[")
	isAnEnum := param.Type.SymbolKind == common.SymKindEnum

	if (param.Type.IsUniverseType() || isAnEnum) && !isErrType && !isMapType {
		return nil
	}

	nameInSchema, err := metadata.GetParameterSchemaName(param.Name, param.Annotations)
	if err != nil {
		nameInSchema = "unknown"
	}
	diag := diagnostics.NewErrorDiagnostic(
		receiver.FVersion.Path,
		fmt.Sprintf(
			"header, path and query parameters are currently limited to primitives only but "+
				"%s parameter '%s' (schema name '%s', type '%s') is of kind '%s'",
			passedIn,
			param.Name,
			nameInSchema,
			param.Type.Name,
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

	isFormParamAlreadyExists := slices.ContainsFunc(funcParams, func(p funcParamEx) bool {
		return p.PassedIn == definitions.PassedInForm
	})

	var errMsg string

	switch newParamType {
	case definitions.PassedInBody:
		if doesBodyParamAlreadyExists {
			// Body is a special case, only one body parameter is allowed per route
			errMsg = "Body parameter is invalid, only one body per route is allowed"
		} else if isFormParamAlreadyExists {
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
			receiver.FVersion.Path,
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
		receiver.FVersion.Path,
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
				receiver.FVersion.Path,
				fmt.Sprintf(
					"expected method '%s' to return an error or a value and error tuple but found void",
					receiver.Name,
				),
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
				receiver.FVersion.Path,
				fmt.Sprintf(
					"expected method '%s' to return an error or a value and error tuple but found (%s)",
					receiver.Name,
					strings.Join(typeNames, ", "),
				),
				diagnostics.DiagReceiverRetValsInvalidSignature,
				receiver.RetValsRange(),
			),
		)
	}

	return errorRetTypeIndex, nil
}
