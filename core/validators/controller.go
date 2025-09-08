package validators

import (
	"fmt"

	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/arbitrators"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/validators/diagnostics"
	"github.com/gopher-fleece/gleece/definitions"
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

	diagnosticsList = append(
		diagnosticsList,
		v.yieldDiagsForExcessiveAttr(counts, annotations.GleeceAnnotationQuery)...,
	)

	diagnosticsList = append(
		diagnosticsList,
		v.yieldDiagsForExcessiveAttr(counts, annotations.GleeceAnnotationPath)...,
	)

	diagnosticsList = append(
		diagnosticsList,
		v.yieldDiagsForExcessiveAttr(counts, annotations.GleeceAnnotationBody)...,
	)

	diagnosticsList = append(
		diagnosticsList,
		v.yieldDiagsForExcessiveAttr(counts, annotations.GleeceAnnotationFormField)...,
	)

	diagnosticsList = append(
		diagnosticsList,
		v.yieldDiagsForExcessiveAttr(counts, annotations.GleeceAnnotationHeader)...,
	)

	diagnosticsList = append(
		diagnosticsList,
		v.yieldDiagsForExcessiveAttr(counts, annotations.GleeceAnnotationMethod)...,
	)

	return diagnosticsList
}

func (v *ControllerValidator) yieldDiagsForExcessiveAttr(countMap map[string]int, attrName string) []diagnostics.ResolvedDiagnostic {
	diags := []diagnostics.ResolvedDiagnostic{}
	if countMap[annotations.GleeceAnnotationPath] <= 0 {
		return diags
	}

	unexpectedAttrInstances := v.holder.GetAll(attrName)
	for _, attrInstance := range unexpectedAttrInstances {
		diags = append(
			diags,
			v.createMayNotHaveAnnotation("Controllers", *attrInstance),
		)
	}

	return diags

}

func (v *ControllerValidator) validateReceiver(receiver *metadata.ReceiverMeta) (diagnostics.EntityDiagnostic, error) {
	validator := NewReceiverValidator(v.gleeceConfig, v.packagesFacade, v.controller, receiver)
	return validator.Validate()
}
