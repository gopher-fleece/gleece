package temp

import (
	SomeRandomName "github.com/haimkastner/gleece/controller"
	"github.com/haimkastner/gleece/temp/nested"
	. "github.com/haimkastner/gleece/temp/nested"
	CustomAlias "github.com/haimkastner/gleece/temp/nested"
)

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

type DefinedInSameFile struct {
}

// DontDoItPlease bla bla bla
// @Query fgd fdffdf
// @Method GET
// @Route /dont
// @Response 204
// @ErrorResponse 400 <p>Test Error 400</p>
// @ErrorResponse 403 <p>Test Error 403</p>
// @ErrorResponse 403 <p>Test Error 403 #2</p>
// @Query theInput the_input
func (ec *ExtendedController) DontDoItPlease(fgd uint64) error {
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
func (ec *ExtendedController) ImportedWithDefaultAliasRetType(definedElseWhere string) (nested.ImportedWithDefaultAlias, error) {
	return nested.ImportedWithDefaultAlias{}, nil
}

// A test for simple imports
// @Query definedElseWhere Testing simple type import
// @Method POST
// @Route /test2
// @Response 204
// @ErrorResponse 400 <p>Test Error 400</p>
func (ec *ExtendedController) ImportedWithCustomAliasRetType() (CustomAlias.ImportedWithCustomAlias, error) {
	return CustomAlias.ImportedWithCustomAlias{}, nil
}

// A test for simple imports
// @Method POST
// @Route /test3
// @Response 204
// @ErrorResponse 400 <p>Test Error 400</p>
func (ec *ExtendedController) ImportedWithDotRetType() (ImportedWithDot, error) {
	return ImportedWithDot{}, nil
}

// A test for simple imports
// @Method POST
// @Route /test4
// @Response 204
// @ErrorResponse 400 <p>Test Error 400</p>
func (ec *ExtendedController) DefinedInSameFileRetType() (DefinedInSameFile, error) {
	return DefinedInSameFile{}, nil
}

// A test for multiple params
// @Method POST
// @Body p1 Body test
// @Header p2 Header test
// @Query(anotherName) p3 Query test
// @Route /test4
// @Response 204
// @ErrorResponse 400 <p>Test Error 400</p>
func (ec *ExtendedController) MultipleParams(p1 string, p2 uint, p3 bool) (DefinedInSameFile, error) {
	return DefinedInSameFile{}, nil
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
