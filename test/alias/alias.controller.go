package generics_test

import (
	"github.com/gopher-fleece/runtime"
)

type TypedefAlias string

type AssignedAlias = string

type NestedTypedefAlias TypedefAlias

type NestedAssignedAlias = TypedefAlias

type BodyWithTypedefAlias struct {
	Ally TypedefAlias
}

type BodyWithAssignedAlias struct {
	Ally AssignedAlias
}

type BodyWithNestedTypedefAlias struct {
	Ally NestedTypedefAlias
}

type BodyWithNestedAssignedAlias struct {
	Ally NestedAssignedAlias
}

// @Tag(Alias Controller Tag)
// @Route(/test/alias)
// @Description Alias Controller
type AliasController struct {
	runtime.GleeceController
}

// Flat TypeDef

// @Method(POST)
// @Route(/td-alias-query)
// @Query(alias)
func (ec *AliasController) ReceivesTypedefAliasQuery(alias TypedefAlias) error {
	return nil
}

// @Method(POST)
// @Route(/td-alias-in-body)
// @Body(body)
func (ec *AliasController) ReceivesATypedefAliasInBody(body BodyWithTypedefAlias) error {
	return nil
}

// @Method(POST)
// @Route(/td-alias-return)
func (ec *AliasController) ReturnsATypedefAlias() (TypedefAlias, error) {
	return "", nil
}

// Flat Assigned

// @Method(POST)
// @Route(/as-alias-query)
// @Query(alias)
func (ec *AliasController) ReceivesAssignedAliasQuery(alias AssignedAlias) error {
	return nil
}

// @Method(POST)
// @Route(/as-alias-in-body)
// @Body(body)
func (ec *AliasController) ReceivesAnAssignedAliasInBody(body BodyWithAssignedAlias) error {
	return nil
}

// @Method(POST)
// @Route(/as-alias-return)
func (ec *AliasController) ReturnsAnAssignedAlias() (AssignedAlias, error) {
	return "", nil
}

// Nested TypeDef

// @Method(POST)
// @Route(/ntd-alias-query)
// @Query(alias)
func (ec *AliasController) ReceivesNestedTypedefAliasQuery(alias NestedTypedefAlias) error {
	return nil
}

// @Method(POST)
// @Route(/ntd-alias-in-body)
// @Body(body)
func (ec *AliasController) ReceivesAnNestedTypedefAliasInBody(body BodyWithNestedTypedefAlias) error {
	return nil
}

// @Method(POST)
// @Route(/ntd-alias-return)
func (ec *AliasController) ReturnsAnNestedTypedefAlias() (NestedTypedefAlias, error) {
	return "", nil
}

// Nested Assigned

// @Method(POST)
// @Route(/nas-alias-query)
// @Query(alias)
func (ec *AliasController) ReceivesNestedAssignedAliasQuery(alias NestedAssignedAlias) error {
	return nil
}

// @Method(POST)
// @Route(/nas-alias-in-body)
// @Body(body)
func (ec *AliasController) ReceivesAnNestedAssignedAliasInBody(body BodyWithNestedAssignedAlias) error {
	return nil
}

// @Method(POST)
// @Route(/nas-alias-return)
func (ec *AliasController) ReturnsAnNestedAssignedAlias() (NestedAssignedAlias, error) {
	return "", nil
}
