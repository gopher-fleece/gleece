package validators

import (
	"fmt"
	"slices"

	"github.com/gopher-fleece/gleece/v2/core/annotations"
	"github.com/gopher-fleece/gleece/v2/core/arbitrators"
	"github.com/gopher-fleece/gleece/v2/core/metadata"
	"github.com/gopher-fleece/gleece/v2/core/validators/diagnostics"
	"github.com/gopher-fleece/gleece/v2/core/validators/paths"
	"github.com/gopher-fleece/gleece/v2/definitions"
)

type ApiValidator struct {
	gleeceConfig   *definitions.GleeceConfig
	packagesFacade *arbitrators.PackagesFacade
	controllers    []metadata.ControllerMeta
}

func NewApiValidator(
	gleeceConfig *definitions.GleeceConfig,
	packagesFacade *arbitrators.PackagesFacade,
	controllers []metadata.ControllerMeta,
) ApiValidator {
	return ApiValidator{
		gleeceConfig:   gleeceConfig,
		packagesFacade: packagesFacade,
		controllers:    controllers,
	}
}

func (v *ApiValidator) Validate() ([]diagnostics.EntityDiagnostic, error) {
	controllerDiags, routeEntries, err := v.validateControllers()
	if err != nil {
		return controllerDiags, fmt.Errorf("failed to validate one or more controllers - %w", err)
	}

	conflicts, err := v.inPlaceAppendPathConflictDiagnostics(controllerDiags, routeEntries)
	if err != nil {
		return controllerDiags, err
	}

	return conflicts, nil
}

func (v *ApiValidator) validateControllers() ([]diagnostics.EntityDiagnostic, []paths.RouteEntry, error) {
	controllerDiags := []diagnostics.EntityDiagnostic{}

	routeEntries := []paths.RouteEntry{}

	for _, ctrl := range v.controllers {
		validator := NewControllerValidator(v.gleeceConfig, v.packagesFacade, &ctrl)
		ctrlDiag, err := validator.Validate()
		if err != nil {
			return controllerDiags, routeEntries, fmt.Errorf(
				"failed to validate controller '%s' due to an error - %w",
				ctrl.Struct.Name,
				err,
			)
		}

		if !ctrlDiag.Empty() {
			controllerDiags = append(controllerDiags, ctrlDiag)
		}

		routeEntries = slices.Concat(routeEntries, v.getRouteEntries(&ctrl))
	}

	return controllerDiags, routeEntries, nil
}

func (v *ApiValidator) getRouteEntries(controller *metadata.ControllerMeta) []paths.RouteEntry {
	entries := make([]paths.RouteEntry, 0, len(controller.Receivers))

	for _, route := range controller.Receivers {
		entries = append(
			entries,
			paths.RouteEntry{
				Path:   route.Annotations.GetFirstValueOrEmpty(annotations.GleeceAnnotationRoute),
				Method: route.Annotations.GetFirstValueOrEmpty(annotations.GleeceAnnotationMethod),
				Meta: paths.RouteEntryMeta{
					Controller: controller,
					Receiver:   &route,
				},
			},
		)
	}

	return entries
}

func (v *ApiValidator) inPlaceAppendPathConflictDiagnostics(
	controllerDiags []diagnostics.EntityDiagnostic,
	routeEntries []paths.RouteEntry,
) ([]diagnostics.EntityDiagnostic, error) {

	conflicts := paths.FindConflicts(routeEntries)

	if len(conflicts) <= 0 {
		return controllerDiags, nil
	}

	for _, conflict := range conflicts {
		controllerDiags = v.adjustDiagsForConflictingEntry(controllerDiags, conflict.A, conflict.Reason)
		controllerDiags = v.adjustDiagsForConflictingEntry(controllerDiags, conflict.B, conflict.Reason)
	}

	return controllerDiags, nil
}

// adjustDiagsForConflictingEntry merges the given controller-level diagnostics with the given
// global path conflict entry
func (v *ApiValidator) adjustDiagsForConflictingEntry(
	controllerDiags []diagnostics.EntityDiagnostic,
	entry paths.RouteEntry,
	conflictReason string,
) []diagnostics.EntityDiagnostic {
	// First, see if there are existing diagnostics for the controller referenced
	// by the route entry
	relevantDiagIdx := slices.IndexFunc(controllerDiags, func(diag diagnostics.EntityDiagnostic) bool {
		return diag.BaseKey() == diagnostics.CreateEntityDiagKey(controllerDiagKind, entry.Meta.Controller.Struct.Name)
	})

	var relevantDiag diagnostics.EntityDiagnostic

	if relevantDiagIdx < 0 {
		// This is the first diagnostic for this controller - need to create a new diagnostic entity
		relevantDiag = diagnostics.NewEntityDiagnostic(
			controllerDiagKind,
			entry.Meta.Controller.Struct.Name,
		)
	} else {
		// This controller already has diagnostics - we just need to append to them
		relevantDiag = controllerDiags[relevantDiagIdx]
	}

	// Get the route for the receiver (API endpoint) mentioned by the conflict entry
	routeAnnotation := entry.Meta.Receiver.Annotations.GetFirst(annotations.GleeceAnnotationRoute)
	receiverResolvedDiag := diagnostics.NewWarningDiagnostic(
		entry.Meta.Receiver.Annotations.FileName(),
		fmt.Sprintf("Path conflict - %s", conflictReason),
		diagnostics.DiagRouteConflict,
		routeAnnotation.GetValueRange(),
	)

	// Same logic as before - if the controller already has a diagnostic for the referenced receiver (route)
	// use that, otherwise, create a new one and append
	receiverDiag := relevantDiag.GetChild(receiverDiagKind, entry.Meta.Receiver.Name)
	if receiverDiag == nil {
		diag := diagnostics.NewEntityDiagnostic(receiverDiagKind, entry.Meta.Receiver.Name)
		diag.AddDiagnostic(receiverResolvedDiag)
		relevantDiag.AddChild(&diag)
	} else {
		receiverDiag.AddDiagnostic(receiverResolvedDiag)
	}

	if relevantDiagIdx >= 0 {
		controllerDiags[relevantDiagIdx] = relevantDiag
	} else {
		controllerDiags = append(controllerDiags, relevantDiag)
	}

	return controllerDiags
}
