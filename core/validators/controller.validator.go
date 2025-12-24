package validators

import (
	"fmt"

	"github.com/gopher-fleece/gleece/v2/core/annotations"
	"github.com/gopher-fleece/gleece/v2/core/arbitrators"
	"github.com/gopher-fleece/gleece/v2/core/metadata"
	"github.com/gopher-fleece/gleece/v2/core/validators/diagnostics"
	"github.com/gopher-fleece/gleece/v2/definitions"
)

type ControllerValidator struct {
	CommonValidator

	gleeceConfig   *definitions.GleeceConfig
	packagesFacade *arbitrators.PackagesFacade
	controller     *metadata.ControllerMeta
}

func NewControllerValidator(
	gleeceConfig *definitions.GleeceConfig,
	packagesFacade *arbitrators.PackagesFacade,
	controller *metadata.ControllerMeta,
) ControllerValidator {
	return ControllerValidator{
		CommonValidator: CommonValidator{
			holder: controller.Struct.Annotations,
		},
		gleeceConfig:   gleeceConfig,
		packagesFacade: packagesFacade,
		controller:     controller,
	}
}

func (v *ControllerValidator) Validate() (diagnostics.EntityDiagnostic, error) {
	controllerDiag := diagnostics.NewEntityDiagnostic("Controller", v.controller.Struct.Name)
	controllerDiag.AddDiagnostics(v.validateSelf())

	for _, receiver := range v.controller.Receivers {
		recDiag, err := v.validateReceiver(&receiver)
		if err != nil {
			return controllerDiag, fmt.Errorf("failed to validate receiver '%s' - %w", receiver.Name, err)
		}

		if !recDiag.Empty() {
			controllerDiag.AddChild(&recDiag)
		}
	}

	return controllerDiag, nil
}

func (v *ControllerValidator) validateSelf() []diagnostics.ResolvedDiagnostic {
	diags := v.CommonValidator.Validate()
	diags = append(diags, v.validateAnnotationPresence()...)

	return diags
}

func (v *ControllerValidator) validateAnnotationPresence() []diagnostics.ResolvedDiagnostic {
	diagnosticsList := []diagnostics.ResolvedDiagnostic{}
	counts := v.controller.Struct.Annotations.AttributeCounts()

	if counts[annotations.GleeceAnnotationTag] <= 0 {
		diagnosticsList = append(
			diagnosticsList,
			diagnostics.NewWarningDiagnostic(
				v.holder.FileName(),
				fmt.Sprintf("Controller '%s' is lacking a @Tag annotation", v.controller.Struct.Name),
				diagnostics.DiagControllerLevelMissingTag,
				v.holder.Range(),
			))
	}

	return diagnosticsList
}

func (v *ControllerValidator) validateReceiver(receiver *metadata.ReceiverMeta) (diagnostics.EntityDiagnostic, error) {
	validator := NewReceiverValidator(v.gleeceConfig, v.packagesFacade, v.controller, receiver)
	return validator.Validate()
}
