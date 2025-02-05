package fiber

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/gopher-fleece/gleece/e2e/common"
)

var FiberRouter *fiber.App

func FiberRouterTest(routerTest common.RouterTest) common.RouterTestResult {
	// Build the query parameters
	params := url.Values{}
	path := routerTest.Path
	if routerTest.Query != nil {
		for k, v := range routerTest.Query {
			params.Add(k, v)
		}
		if encoded := params.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	// Prepare request body if available.
	var jsonDataBuffer *bytes.Buffer
	if routerTest.Body != nil {
		jsonData, _ := json.Marshal(routerTest.Body)
		jsonDataBuffer = bytes.NewBuffer(jsonData)
	}

	// Create the new HTTP request.
	var req *http.Request
	if jsonDataBuffer == nil {
		req = httptest.NewRequest(routerTest.Method, path, nil)
	} else {
		req = httptest.NewRequest(routerTest.Method, path, jsonDataBuffer)
	}

	// Add the provided headers to the request.
	if routerTest.Headers != nil {
		for k, v := range routerTest.Headers {
			req.Header.Add(strings.ToLower(k), v)
		}
	}

	// Execute the request on the Fiber app.
	// The second parameter is the timeout in milliseconds.
	resp, err := FiberRouter.Test(req, -1)
	if err != nil {
		// Handle error as needed (e.g. panic or return a default result)
		panic(err)
	}

	// Read the response body.
	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	// Convert the HTTP response headers to a map[string]string.
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[strings.ToLower(k)] = v[0]
		}
	}

	// Trim any trailing whitespace from the response body.
	bodyRes := string(bodyBytes)
	if bodyRes != "" {
		bodyRes = strings.TrimRightFunc(bodyRes, unicode.IsSpace)
	}

	return common.RouterTestResult{
		Code:    resp.StatusCode,
		Body:    bodyRes,
		Headers: headers,
	}
}
