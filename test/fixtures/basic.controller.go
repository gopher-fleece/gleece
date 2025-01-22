package temp

import (
	SomeRandomName "github.com/gopher-fleece/gleece/external"
	"github.com/gopher-fleece/gleece/test/types"
	. "github.com/gopher-fleece/gleece/test/types"
	CustomAlias "github.com/gopher-fleece/gleece/test/types"
)

// ExtendedController
// @Tag(Something with a space)
// @Route(/users)
// @Description This is an extended controller
type ExtendedController struct {
	SomeRandomName.GleeceController // Embedding the GleeceController to inherit its methods
}

type EmbedsAnError struct {
	error
}

// @Description Ahhhh
type DefinedInSameFile struct {
	// Some comment
	SomeField string `json:"someField" validator:"required,email"`
}

// A test for returning embedded errors
// @Method(POST)
// @Route(/test/embedded/error, {"someContext": 53553})
// @Security(SchemaX, { scopes: ["a", "b"]})
func (ec *ExtendedController) ReturnEmbedsAndError() (HoldsVeryNestedStructs, EmbedsAnError) {
	return HoldsVeryNestedStructs{}, EmbedsAnError{}
}

// A test for simple imports
// @Query(definedElseWhere, {name:'someAlias', validate:'required, email'}) Testing simple type import
// @Method(POST)
// @Route(/test)
// @Response(204)
// @Security(SchemaZ, { scopes: ["c"]})
// @ErrorResponse(400) <p>Test Error 400</p>
func (ec *ExtendedController) ImportedWithDefaultAliasRetType(definedElseWhere string) (types.ImportedWithDefaultAlias, error) {
	return types.ImportedWithDefaultAlias{}, nil
}

// A test for simple imports
// @Query definedElseWhere Testing simple type import
// @Method(POST)
// @Route(/test2)
// @Response(204)
// @ErrorResponse(400) <p>Test Error 400</p>
func (ec *ExtendedController) ImportedWithCustomAliasRetType() (CustomAlias.ImportedWithCustomAlias, error) {
	return CustomAlias.ImportedWithCustomAlias{}, nil
}

// A test for simple imports
// @Method(POST)
// @Route(/test3)
// @Response(204)
// @ErrorResponse(400) <p>Test Error 400</p>
func (ec *ExtendedController) ImportedWithDotRetType() (ImportedWithDot, error) {
	return ImportedWithDot{}, nil
}

// A test for simple imports
// @Method(POST)
// @Route(/test4)
// @Response(204)
// @ErrorResponse(400) <p>Test Error 400</p>
func (ec *ExtendedController) DefinedInSameFileRetType() (DefinedInSameFile, error) {
	return DefinedInSameFile{}, nil
}

// For simple @Hidden annotation test
//
// @Method(GET)
// @Route(/ignored-method)
// @Response(204)
// @Query(value)
// @Hidden
func (ec *ExtendedController) HiddenMethodSimple(value uint32) error {
	return nil
}

// For conditional @Hidden annotation test
//
// @Method(GET)
// @Route(/ignored-method-2)
// @Response(204)
// @Query(value)
// @Hidden($BRANCH=="master")
func (ec *ExtendedController) HiddenMethodConditional(value uint32) error {
	return nil
}

// For simple @Deprecated annotation test
//
// @Method(GET)
// @Route(/deprecated-method)
// @Response(204)
// @Query(value)
// @Deprecated
func (ec *ExtendedController) DeprecatedMethodSimple(value uint32) error {
	return nil
}

// For conditional @Deprecated annotation test
//
// @Method(GET)
// @Route(/deprecated-method-2)
// @Response(204)
// @Query(value)
// @Deprecated This method is deprecated
func (ec *ExtendedController) DeprecatedMethodConditional(value uint32) error {
	return nil
}
