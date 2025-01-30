package e2e

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/gopher-fleece/gleece/cmd"
	"github.com/gopher-fleece/gleece/cmd/arguments"

	"github.com/gopher-fleece/gleece/e2e/assets"
	gleeceRoutes "github.com/gopher-fleece/gleece/e2e/gin/routes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var r *gin.Engine

var _ = BeforeSuite(func() {
	// Preparation phase...

	// Build routes...
	err := cmd.GenerateSpecAndRoutes(arguments.CliArguments{ConfigPath: "./gin.e2e.gleece.config.json"})

	if err != nil {
		Fail("Failed to generate routes" + err.Error())
	}
	// Init routes
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Setup your router
	r = gin.Default()

	gleeceRoutes.RegisterRoutes(r)
})

var _ = Describe("Gin E2E Spec", func() {

	It("Should return status code 200 for simple get", func() {
		// Create a response recorder
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/e2e/simple-get", nil)
		r.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(200))
		Expect(w.Body.String()).To(Equal("\"works\""))
	})

	It("Should set custom header", func() {
		// Create a response recorder
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/e2e/simple-get", nil)
		r.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(200))
		Expect(w.Body.String()).To(Equal("\"works\""))
		Expect(w.Header().Get("X-Test-Header")).To(Equal("test"))
	})

	It("Should set custom template header", func() {
		// Create a response recorder
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/e2e/simple-get", nil)
		r.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(200))
		Expect(w.Body.String()).To(Equal("\"works\""))
		Expect(w.Header().Get("X-Test-Header")).To(Equal("test"))
		Expect(w.Header().Get("x-inject")).To(Equal("true"))
	})

	It("Should return status code 204 for explicit set status", func() {
		// Create a response recorder
		w := httptest.NewRecorder()
		params := url.Values{}
		params.Add("queryParam", "204")
		req := httptest.NewRequest("GET", "/e2e/get-with-all-params/pathParam"+"?"+params.Encode(), nil)
		req.Header.Add("headerParam", "headerParam")
		r.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(204))
		Expect(w.Body.String()).To(Equal("")) // In gin, the body is not written when status is 204
	})

	It("Should return status code 200 for get with all params in use", func() {
		// Create a response recorder
		w := httptest.NewRecorder()
		params := url.Values{}
		params.Add("queryParam", "queryParam")
		req := httptest.NewRequest("GET", "/e2e/get-with-all-params/pathParam"+"?"+params.Encode(), nil)
		req.Header.Add("headerParam", "headerParam")
		r.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(200))
		Expect(w.Body.String()).To(Equal("\"pathParamqueryParamheaderParam\""))
	})

	It("Should return status code 200 for get with all params ptr", func() {
		// Create a response recorder
		w := httptest.NewRecorder()
		params := url.Values{}
		params.Add("queryParam", "queryParam")
		req := httptest.NewRequest("GET", "/e2e/get-with-all-params-ptr/pathParam"+"?"+params.Encode(), nil)
		req.Header.Add("headerParam", "headerParam")
		r.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(200))
		Expect(w.Body.String()).To(Equal("\"pathParamqueryParamheaderParam\""))
	})

	It("Should return status code 200 for get with all params empty ptr", func() {
		// Create a response recorder
		w := httptest.NewRecorder()
		params := url.Values{}
		req := httptest.NewRequest("GET", "/e2e/get-with-all-params-ptr/pathParam"+"?"+params.Encode(), nil)
		r.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(200))
		Expect(w.Body.String()).To(Equal("\"pathParam\""))
	})

	It("Should return status code 422 for get with all params empty ptr", func() {
		// Create a response recorder
		w := httptest.NewRecorder()
		params := url.Values{}
		params.Add("queryParam", "queryParam")
		req := httptest.NewRequest("GET", "/e2e/get-with-all-params-required-ptr/pathParam"+"?"+params.Encode(), nil)
		r.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(422))
		Expect(w.Body.String()).To(ContainSubstring("A request was made to operation 'GetWithAllParamsRequiredPtr' but parameter 'headerParam' did not pass validation - Field 'headerParam' failed validation with tag 'required'"))

		params = url.Values{}
		req = httptest.NewRequest("GET", "/e2e/get-with-all-params-required-ptr/pathParam"+"?"+params.Encode(), nil)
		req.Header.Add("headerParam", "headerParam")
		r.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(422))
		Expect(w.Body.String()).To(ContainSubstring("A request was made to operation 'GetWithAllParamsRequiredPtr' but parameter 'queryParam' did not pass validation - Field 'queryParam' failed validation with tag 'required'"))
	})

	It("Should return status code 200 for get with all params with body", func() {
		// Create a response recorder
		w := httptest.NewRecorder()
		params := url.Values{}
		params.Add("queryParam", "queryParam")
		jsonData, _ := json.Marshal(assets.BodyInfo{BodyParam: "thebody"})
		req := httptest.NewRequest("POST", "/e2e/post-with-all-params-body"+"?"+params.Encode(), bytes.NewBuffer(jsonData))
		req.Header.Add("headerParam", "headerParam")
		r.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(200))
		Expect(w.Body.String()).To(ContainSubstring("queryParamheaderParamthebody"))
	})

	It("Should return status code 422 for missing body", func() {
		// Create a response recorder
		w := httptest.NewRecorder()
		params := url.Values{}
		params.Add("queryParam", "queryParam")
		req := httptest.NewRequest("POST", "/e2e/post-with-all-params-body"+"?"+params.Encode(), nil)
		req.Header.Add("headerParam", "headerParam")
		r.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(422))
		Expect(w.Body.String()).To(ContainSubstring("A request was made to operation 'PostWithAllParamsWithBody' but body parameter 'theBody' did not pass validation of 'BodyInfo' - body is required but was not provided"))
	})

	It("Should return status code 200 for get with all params with body ptr", func() {
		// Create a response recorder
		w := httptest.NewRecorder()
		params := url.Values{}
		params.Add("queryParam", "queryParam")
		jsonData, _ := json.Marshal(assets.BodyInfo{BodyParam: "thebody"})
		req := httptest.NewRequest("POST", "/e2e/post-with-all-params-body-ptr"+"?"+params.Encode(), bytes.NewBuffer(jsonData))
		req.Header.Add("headerParam", "headerParam")
		r.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(200))
		Expect(w.Body.String()).To(ContainSubstring("queryParamheaderParamthebody"))

		params = url.Values{}
		params.Add("queryParam", "queryParam")
		req = httptest.NewRequest("POST", "/e2e/post-with-all-params-body-ptr"+"?"+params.Encode(), nil)
		req.Header.Add("headerParam", "headerParam")
		r.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(200))
		Expect(w.Body.String()).To(ContainSubstring("empty"))
	})

	It("Should return status code 200 for get with all params with body required ptr", func() {
		// Create a response recorder
		w := httptest.NewRecorder()
		params := url.Values{}
		params.Add("queryParam", "queryParam")
		req := httptest.NewRequest("POST", "/e2e/post-with-all-params-body-required-ptr"+"?"+params.Encode(), nil)
		req.Header.Add("headerParam", "headerParam")
		r.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(422))
		Expect(w.Body.String()).To(ContainSubstring("A request was made to operation 'PostWithAllParamsWithBodyRequiredPtr' but body parameter 'theBody' did not pass validation of 'BodyInfo' - body is required but was not provided"))

		params = url.Values{}
		params.Add("queryParam", "queryParam")
		jsonData, _ := json.Marshal(assets.BodyInfo{})
		req = httptest.NewRequest("POST", "/e2e/post-with-all-params-body-required-ptr"+"?"+params.Encode(), bytes.NewBuffer(jsonData))
		req.Header.Add("headerParam", "headerParam")
		r.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(422))
		Expect(w.Body.String()).To(ContainSubstring("A request was made to operation 'PostWithAllParamsWithBodyRequiredPtr' but body parameter 'theBody' did not pass validation of 'BodyInfo' - body is required but was not provided"))

		params = url.Values{}
		params.Add("queryParam", "queryParam")
		jsonData, _ = json.Marshal(assets.BodyInfo2{BodyParam: 1})
		req = httptest.NewRequest("POST", "/e2e/post-with-all-params-body-required-ptr"+"?"+params.Encode(), bytes.NewBuffer(jsonData))
		req.Header.Add("headerParam", "headerParam")
		r.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(422))
		Expect(w.Body.String()).To(ContainSubstring("cannot unmarshal number into Go struct field BodyInfo.bodyParam of type string"))

		params = url.Values{}
		params.Add("queryParam", "queryParam")
		jsonData, _ = json.Marshal(assets.BodyInfo{})
		req = httptest.NewRequest("POST", "/e2e/post-with-all-params-body-required-ptr"+"?"+params.Encode(), bytes.NewBuffer(jsonData))
		req.Header.Add("headerParam", "headerParam")
		r.ServeHTTP(w, req)
		Expect(w.Code).To(Equal(422))
		Expect(w.Body.String()).To(ContainSubstring("Field 'BodyParam' failed validation with tag 'required'"))

	})
})
