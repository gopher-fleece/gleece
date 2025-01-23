package fixtures

import (
	"github.com/google/uuid"
	SomeRandomName "github.com/gopher-fleece/gleece/external"
	"github.com/gopher-fleece/gleece/test/types"
	. "github.com/gopher-fleece/gleece/test/types"
	CustomAlias "github.com/gopher-fleece/gleece/test/types"
)

// ExtendedController
// @Tag(Something with a space)
// @Route(/soothing-with-a-space)
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
// @Security(securitySchemaName, { scopes: ["a", "b"]})
// @Security(securitySchemaName, { scopes: ["c"]})
func (ec *ExtendedController) ReturnEmbedsAndError() (HoldsVeryNestedStructs, EmbedsAnError) {
	return HoldsVeryNestedStructs{}, EmbedsAnError{}
}

// A test for simple imports
// @Query(definedElseWhere, {name:'someAlias', validate:'required, email'}) Testing simple type import
// @Method(POST)
// @Route(/test)
// @Response(204)
// @Security(securitySchemaName, { scopes: ["c"]})
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

// UsersController
// @Tag(Users) Users
// @Route(/users)
// @Description The Users API
type UsersController struct {
	SomeRandomName.GleeceController // Embedding the GleeceController to inherit its methods
}

// @Description User's domicile
type Domicile struct {
	// @Description The address
	Address string `json:"address" validate:"required"`
	// @Description The number of the house (must be at least 1)
	HouseNumber int `json:"houseNumber" validate:"gte=1"`
}

// @Method(POST) This text is not part of the OpenAPI spec
// @Route(/user/{id}/domicile)
// @Path(id)
// @Body(domicile)
// @Response(200) The user's domicile
// @ErrorResponse(404) The user not found
// @ErrorResponse(500) The error when process failed
// @Security(securitySchemaName, { scopes: ["read:users"] })
func (ec *UsersController) SetUserDomicile(id string, domicile Domicile) error {
	return nil
}

// @Description Get user's domicile
// @Method(GET) This text is not part of the OpenAPI spec
// @Route(/user/{id}/domicile)
// @Path(id)
// @Response(200) The user's domicile
// @ErrorResponse(404) The user not found
// @ErrorResponse(500) The error when process failed
// @Security(securitySchemaName, { scopes: ["read:users"] })
func (ec *UsersController) GetUserDomicile(id string) (Domicile, error) {
	return Domicile{
		Address:     "Jl. Jend. Sudirman",
		HouseNumber: 1,
	}, nil
}

// @Description Create a new user
// @Method(POST) This text is not part of the OpenAPI spec
// @Route(/user/{user_name}/{user_id}) Same here
// @Query(email, { validate: "required,email" }) The user's email
// @Path(id, { name: "user_id", validate:"gt=1" }) The user's ID
// @Path(id2, { name: "user_id_2", validate:"gt=10" }) The user's ID 2
// @Path(name, { name: "user_name" }) The user's name
// @Body(domicile) The user's domicile
// @Header(origin, { name: "x-origin" }) The request origin
// @Header(trace) The trace info
// @Response(200) The ID of the newly created user
// @ErrorResponse(500) The error when process failed
// @Security(securitySchemaName, { scopes: ["read:users", "write:users"] })
func (ec *UsersController) CreateNewUser(id int, id2 int, email string, name string, origin string, trace string, domicile Domicile) (string, error) {
	userId := uuid.New()
	return domicile.Address + " " + userId.String(), nil
}
