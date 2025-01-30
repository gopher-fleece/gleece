package imports_test

import (
	. "github.com/gopher-fleece/gleece/external"
	"github.com/gopher-fleece/gleece/test/types"
	. "github.com/gopher-fleece/gleece/test/types"
	alias "github.com/gopher-fleece/gleece/test/types"
)

// @Tag(Imports Controller Tag)
// @Route(/test/imports)
type ImportsController struct {
	// Importing with . to test that specific AST enumeration branch
	// which is slightly different than for function parameters/return values
	GleeceController
}

// @Method(POST)
// @Route(/imported-with-dot)
// @Body(input)
func (ec *ImportsController) ImportedWithDot(input ImportedWithDot) (ImportedWithDot, error) {
	return ImportedWithDot{}, nil
}

// @Method(POST)
// @Route(/imported-with-default-alias)
// @Body(input)
func (ec *ImportsController) ImportedWithDefaultAlias(
	input types.ImportedWithDefaultAlias,
) (types.ImportedWithDefaultAlias, error) {
	return types.ImportedWithDefaultAlias{}, nil
}

// @Method(POST)
// @Route(/imported-with-custom-alias)
// @Body(input)
func (ec *ImportsController) ImportedWithCustomAlias(
	input alias.ImportedWithCustomAlias,
) (alias.ImportedWithCustomAlias, error) {
	return alias.ImportedWithCustomAlias{}, nil
}
