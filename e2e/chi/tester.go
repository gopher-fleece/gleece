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
	params := url.Values{}

	path := routerTest.Path

	// Add query parameters
	if routerTest.Query != nil {
		for k, v := range routerTest.Query {
			params.Add(k, v)
		}
		path += "?" + params.Encode()
	}

	var jsonDataBuffer *bytes.Buffer = nil
	if routerTest.Body != nil {
		jsonData, _ := json.Marshal(routerTest.Body)
		jsonDataBuffer = bytes.NewBuffer(jsonData)
	}

	var req *http.Request
	if jsonDataBuffer == nil {
		req = httptest.NewRequest(routerTest.Method, path, nil)
	} else {
		req = httptest.NewRequest(routerTest.Method, path, jsonDataBuffer)
	}

	// Add headers to the request
	if routerTest.Headers != nil {
		for k, v := range routerTest.Headers {
			req.Header.Add(strings.ToLower(k), v)
		}
	}

	// Replace echoRouter.ServeHTTP with muxRouter.ServeHTTP
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
