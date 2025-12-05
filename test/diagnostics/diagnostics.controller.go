package diagnostics

import "github.com/gopher-fleece/runtime"

// Missing a Tag annotation to trigger a warning diagnostic
// @Route(/test/diagnostics)
type DiagnosticsController struct {
	runtime.GleeceController
}

// @Route(/non-standard-status-code)
// @Method(POST)
// @Response(641)
func (c DiagnosticsController) MethodWithNonStandardStatusCode() error {
	return nil
}

// @Route(/unannotated-param)
// @Method(POST)
func (c DiagnosticsController) MethodWithUnAnnotatedParam(id string) error {
	return nil
}

// @Route(/unlinked-param/{id})
// @Method(POST)
func (c DiagnosticsController) MethodWithUnlinkedParam(id string) error {
	return nil
}

// @Route(/unlinked-alias/{aliased})
// @Method(POST)
// @Path(id)
func (c DiagnosticsController) MethodWithUnlinkedAlias(id string) error {
	return nil
}
