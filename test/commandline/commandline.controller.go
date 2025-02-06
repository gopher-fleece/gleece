package imports_test

import (
	"github.com/gopher-fleece/runtime"
)

// @Tag(Commandline Controller Tag)
// @Route(/test/commandline)
type CommandlineController struct {
	runtime.GleeceController
}

// @Method(POST)
// @Route(/empty-function)
func (ec *CommandlineController) EmptyFunction() error {
	return nil
}
