package mux

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"unicode"

	"github.com/gopher-fleece/gleece/e2e/common"
	"github.com/gorilla/mux"
)

var MuxRouter *mux.Router
var MuxExExtraRouter *mux.Router

func MuxRouterTest(routerTest common.RouterTest) common.RouterTestResult {
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

	if routerTest.QueryArray != nil {
		for k, v := range routerTest.QueryArray {
			for _, vItem := range v {
				queryParams.Add(k, vItem)
			}
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

	// Use Mux router to serve the request
	switch *routerTest.RunningMode {
	case common.RunOnVanillaRoutes:
		MuxExExtraRouter.ServeHTTP(w, req)
	case common.RunOnFullyFeaturedRoutes:
		MuxRouter.ServeHTTP(w, req)
	default:
		return common.RouterTestResult{}
	}

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
