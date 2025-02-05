package gin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gopher-fleece/gleece/e2e/common"
)

var GinRouter *gin.Engine

func GinRouterTest(routerTest common.RouterTest) common.RouterTestResult {
	w := httptest.NewRecorder()
	params := url.Values{}

	path := routerTest.Path

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

	if routerTest.Headers != nil {
		for k, v := range routerTest.Headers {
			req.Header.Add(strings.ToLower(k), v)
		}
	}

	GinRouter.ServeHTTP(w, req)

	// Convert response headers to map[string]string
	headers := make(map[string]string)
	for k, v := range w.Header() {
		if len(v) > 0 {
			headers[strings.ToLower(k)] = v[0]
		}
	}

	return common.RouterTestResult{
		Code:    w.Code,
		Body:    w.Body.String(),
		Headers: headers,
	}
}
