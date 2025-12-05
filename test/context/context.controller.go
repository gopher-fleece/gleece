package sanity_test

import (
	"context"

	"github.com/gopher-fleece/runtime"
)

// @Tag(Context Controller Tag)
// @Route(/test/context)
// @Description Context Controller
type ContextController struct {
	runtime.GleeceController // Embedding the GleeceController to inherit its methods
}

// @Method(POST)
// @Route(/context-and-route-param/{id})
// @Path(id)
func (ec *ContextController) MethodWithContext(ctx context.Context, id string) error {
	return nil
}

// @Method(POST)
// @Route(/context-as-last-param/{id})
// @Path(id)
func (ec *ContextController) MethodWithLastParamContext(id string, ctx context.Context) error {
	return nil
}
