package chi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"unicode"

	"github.com/go-chi/chi/v5"

	"github.com/gopher-fleece/gleece/e2e/common"
)

var ChiRouter *chi.Mux

func ChiRouterTest(routerTest common.RouterTest) common.RouterTestResult {
	// Create a response recorder
	w := httptest.NewRecorder()
	queryParams := url.Values{}
	formParams := url.Values{}

	path := routerTest.Path

	// Add query parameters
	if routerTest.Query != nil {
		for k, v := range routerTest.Query {
			queryParams.Add(k, v)
		}
		path += "?" + queryParams.Encode()
	}

	var req *http.Request

	// Handle form data
	if routerTest.Form != nil {
		// Convert form data to url.Values
		for k, v := range routerTest.Form {
			formParams.Add(k, v)
		}
		// Create request with form data
		req = httptest.NewRequest(routerTest.Method, path, strings.NewReader(formParams.Encode()))
		// Set content type for form data
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if routerTest.Body != nil {
		// Handle JSON body
		jsonData, _ := json.Marshal(routerTest.Body)
		req = httptest.NewRequest(routerTest.Method, path, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
	} else {
		// No body or form data
		req = httptest.NewRequest(routerTest.Method, path, nil)
	}

	// Add headers to the request
	if routerTest.Headers != nil {
		for k, v := range routerTest.Headers {
			req.Header.Add(strings.ToLower(k), v)
		}
	}

	// Use Chi router to serve the request
	ChiRouter.ServeHTTP(w, req)

	// Convert response headers to map[string]string
	headers := make(map[string]string)
	for k, v := range w.Header() {
		if len(v) > 0 {
			headers[strings.ToLower(k)] = v[0]
		}
	}

	bodyRes := w.Body.String()
	if bodyRes != "" {
		bodyRes = strings.TrimRightFunc(bodyRes, unicode.IsSpace)
	}
	return common.RouterTestResult{
		Code:    w.Code,
		Body:    bodyRes,
		Headers: headers,
	}
}
