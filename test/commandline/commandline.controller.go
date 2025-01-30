package imports_test

import (
	"github.com/gopher-fleece/gleece/external"
)

// @Tag(Commandline Controller Tag)
// @Route(/test/commandline)
type CommandlineController struct {
	external.GleeceController
}

// @Method(POST)
// @Route(/empty-function)
func (ec *CommandlineController) EmptyFunction() error {
	return nil
}
