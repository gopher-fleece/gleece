package gin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gopher-fleece/gleece/v2/e2e/common"
)

var GinRouter *gin.Engine
var GinExExtraRouter *gin.Engine

func GinRouterTest(routerTest common.RouterTest) common.RouterTestResult {
	w := httptest.NewRecorder()
	queryParams := url.Values{}
	formParams := url.Values{}

	path := routerTest.Path

	// Handle query parameters
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

	// Add custom headers
	if routerTest.Headers != nil {
		for k, v := range routerTest.Headers {
			req.Header.Add(strings.ToLower(k), v)
		}
	}

	switch *routerTest.RunningMode {
	case common.RunOnVanillaRoutes:
		GinExExtraRouter.ServeHTTP(w, req)
	case common.RunOnFullyFeaturedRoutes:
		GinRouter.ServeHTTP(w, req)
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

	return common.RouterTestResult{
		Code:    w.Code,
		Body:    w.Body.String(),
		Headers: headers,
	}
}
