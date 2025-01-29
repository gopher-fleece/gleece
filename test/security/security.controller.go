package security_test

import (
	"github.com/gopher-fleece/gleece/external"
)

// @Tag(Sanity Controller Tag)
// @Route(/test/sanity)
type SecurityController struct {
	external.GleeceController // Embedding the GleeceController to inherit its methods
}

// A sanity test controller method
// @Method(POST)
// @Route(/security)
// @Security(secSchema1, { scopes: ["scope1"] })
// @Security(secSchema2, { scopes: ["2", "3"] })
func (ec *SecurityController) ValidMethodWithComplexSecurity() error {
	return nil
}
