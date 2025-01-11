package temp

import (
	SomeRandomName "github.com/haimkastner/gleece/controller"
)

// Extended2Controller
// @Tag Users2
// @Route /users2
// @Description This is an extended controller222222222
type Extended2Controller struct {
	SomeRandomName.GleeceController // Embedding the GleeceController to inherit its methods
}

// ExtendedController
// @Tag Users
// @Route /users
// @Description This is an extended controller
type ExtendedController struct {
	SomeRandomName.GleeceController // Embedding the GleeceController to inherit its methods
}

type GetUserInput2 struct {
	UserID string `query:"userId"`
}

type GetUserInput struct {
	UserID GetUserInput2
}

//// DoItPlease bla bla bla
//// @Method GET
//// @Route /
//// @Query(the_input) theInput ParameterId kaki
//// @Body(theBody)     theBody      The body of the request YAP
//// @Header data
//// @ResponseCode 200
//func (ec *ExtendedController) DoItPlease(theInput GetUserInput2, theBody CreateUserInput, data string) (string, error) {
//
//	return "OK?", nil
//}

// DontDoItPlease bla bla bla
// @Query fgd fdffdf
// @Method GET
// @Route /dont
// @Response 204
// @ErrorResponse 400 <p>Test Error 400</p>
// @ErrorResponse 403 <p>Test Error 403</p>
// @ErrorResponse 403 <p>Test Error 403 #2</p>
// @Query theInput the_input
func (ec *ExtendedController) DontDoItPlease(fgd GetUserInput) error {
	// Print  fgd as a string
	println(fgd)
	return nil
}

// A test for simple imports
// @Query definedElseWhere Testing simple type import
// @Method POST
// @Route /test
// @Response 204
// @ErrorResponse 400 <p>Test Error 400</p>
func (ec *ExtendedController) ExternalTypeTest(definedElseWhere SimpleType) (string, error) {
	return "", nil
}

//// DoItPlease2 bla bla bla
//// @Method GET
//// @Route /
//// @ResponseCode 204
//// @Query theInput the_input
//func (ec *Extended2Controller) DoItPlease2(fgd string) error {
//	return nil
//}

// SkipIt bla bla bla
// @Method GET
// @Route /dont
// @Response 204
// @Query theInput the_input
func SkipIt(fgd string) error {
	return nil
}

type CreateUserInput struct {
	Name string `json:"name"`
	// @Ignore
	// @Name email
	Email string `json:"email"`
}
