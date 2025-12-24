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
	"github.com/gopher-fleece/gleece/v2/e2e/common"
)

var FiberRouter *fiber.App
var FiberExExtraRouter *fiber.App

func FiberRouterTest(routerTest common.RouterTest) common.RouterTestResult {
	// Build the query parameters
	queryParams := url.Values{}
	formParams := url.Values{}
	path := routerTest.Path

	if routerTest.Query != nil {
		for k, v := range routerTest.Query {
			queryParams.Add(k, v)
		}
		if encoded := queryParams.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	if routerTest.QueryArray != nil {
		for k, v := range routerTest.QueryArray {
			for _, vItem := range v {
				queryParams.Add(k, vItem)
			}
		}
		if encoded := queryParams.Encode(); encoded != "" {
			path += "?" + encoded
		}
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

	// Add the provided headers to the request
	if routerTest.Headers != nil {
		for k, v := range routerTest.Headers {
			req.Header.Add(strings.ToLower(k), v)
		}
	}

	// Execute the request on the Fiber app
	// The second parameter is the timeout in milliseconds
	var resp *http.Response
	var err error

	switch *routerTest.RunningMode {
	case common.RunOnVanillaRoutes:
		resp, err = FiberExExtraRouter.Test(req, -1)
	case common.RunOnFullyFeaturedRoutes:
		resp, err = FiberRouter.Test(req, -1)
	default:
		return common.RouterTestResult{}
	}

	if err != nil {
		// Handle error as needed (e.g. panic or return a default result)
		panic(err)
	}

	// Read the response body
	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	// Convert the HTTP response headers to a map[string]string
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[strings.ToLower(k)] = v[0]
		}
	}

	// Trim any trailing whitespace from the response body
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
