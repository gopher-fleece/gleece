package assets

import (
	"github.com/gopher-fleece/gleece/external"
)

// @Route(/e2e)
type E2EController struct {
	external.GleeceController // Embedding the GleeceController to inherit its methods
}

// @Method(GET) This text is not part of the OpenAPI spec
// @Route(/simple-get)
func (ec *E2EController) SimpleGet() (string, error) {
	ec.SetHeader("X-Test-Header", "test")
	return "works", nil
}

// @Method(GET) This text is not part of the OpenAPI spec
// @Route(/get-with-all-params/{pathParam})
// @Query(queryParam)
// @Path(pathParam)
// @Header(headerParam)
func (ec *E2EController) GetWithAllParams(queryParam string, pathParam string, headerParam string) (string, error) {
	if queryParam == "204" {
		ec.SetStatus(external.StatusNoContent)
	}
	return pathParam + queryParam + headerParam, nil
}

// @Method(GET) This text is not part of the OpenAPI spec
// @Route(/get-with-all-params-ptr/{pathParam})
// @Query(queryParam)
// @Path(pathParam)
// @Header(headerParam)
func (ec *E2EController) GetWithAllParamsPtr(queryParam *string, pathParam *string, headerParam *string) (string, error) {
	if queryParam == nil {
		queryParam = new(string)
	}
	if headerParam == nil {
		headerParam = new(string)
	}
	return *pathParam + *queryParam + *headerParam, nil
}

// @Method(GET) This text is not part of the OpenAPI spec
// @Route(/get-with-all-params-required-ptr/{pathParam})
// @Query(queryParam, { validate: "required" })
// @Path(pathParam, { validate: "required" })
// @Header(headerParam, { validate: "required" })
func (ec *E2EController) GetWithAllParamsRequiredPtr(queryParam *string, pathParam *string, headerParam *string) (string, error) {
	if queryParam == nil {
		queryParam = new(string)
	}
	if headerParam == nil {
		headerParam = new(string)
	}
	return *pathParam + *queryParam + *headerParam, nil
}

type BodyInfo struct {
	BodyParam string `json:"bodyParam" validate:"required"`
}

type BodyInfo2 struct {
	BodyParam int `json:"bodyParam"`
}

// @Method(POST) This text is not part of the OpenAPI spec
// @Route(/post-with-all-params-body)
// @Query(queryParam)
// @Body(theBody)
// @Header(headerParam)
func (ec *E2EController) PostWithAllParamsWithBody(queryParam string, headerParam string, theBody BodyInfo) (BodyInfo, error) {
	return BodyInfo{
		BodyParam: queryParam + headerParam + theBody.BodyParam,
	}, nil
}

// @Method(POST) This text is not part of the OpenAPI spec
// @Route(/post-with-all-params-body-ptr)
// @Query(queryParam)
// @Body(theBody)
// @Header(headerParam)
func (ec *E2EController) PostWithAllParamsWithBodyPtr(queryParam *string, headerParam *string, theBody *BodyInfo) (*BodyInfo, error) {
	if queryParam == nil {
		queryParam = new(string)
	}
	if headerParam == nil {
		headerParam = new(string)
	}
	if theBody == nil {
		theBody = new(BodyInfo)
		theBody.BodyParam = "empty"
	}
	return &BodyInfo{
		BodyParam: *queryParam + *headerParam + theBody.BodyParam,
	}, nil
}

// @Method(POST) This text is not part of the OpenAPI spec
// @Route(/post-with-all-params-body-required-ptr)
// @Body(theBody, { validate: "required" })
func (ec *E2EController) PostWithAllParamsWithBodyRequiredPtr(theBody *BodyInfo) (*BodyInfo, error) {
	return &BodyInfo{
		BodyParam: theBody.BodyParam,
	}, nil
}

// @Method(GET) This text is not part of the OpenAPI spec
// @Route(/get-header-start-with-letter)
// @Header(headerParam, { validate: "required,validate_starts_with_letter" })
func (ec *E2EController) GetHeaderStartWithLetter(headerParam string) (string, error) {
	return headerParam, nil
}
