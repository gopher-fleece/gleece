package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"

	"github.com/gopher-fleece/gleece/cmd"
	"github.com/gopher-fleece/gleece/cmd/arguments"
	gleeceEchoRoutes "github.com/gopher-fleece/gleece/e2e/echo/routes"
	gleeceGinRoutes "github.com/gopher-fleece/gleece/e2e/gin/routes"
	gleeceMuxRoutes "github.com/gopher-fleece/gleece/e2e/mux/routes"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/labstack/echo/v4"

	"github.com/haimkastner/unitsnet-go/units"
	"github.com/labstack/echo/v4/middleware"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGleeceE2E(t *testing.T) {
	// Disable logging to reduce clutter.
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gleece E2E Suite")
}

var ginRouter *gin.Engine
var echoRouter *echo.Echo
var muxRouter *mux.Router

var suitPrepareTimeout, _ = units.DurationFactory{}.FromMinutes(1.5)

func RegenerateRoutes() {

	// Always build routes for gin  ...
	err := cmd.GenerateSpecAndRoutes(arguments.CliArguments{ConfigPath: "./gin.e2e.gleece.config.json"})
	if err != nil {
		Fail("Failed to generate gin routes" + err.Error())
	}

	// Get from env var whenever to regenerate all routes again.
	// Use it only when modifying the templates which requires new routes for tests for all other engines too.
	generate, exists := os.LookupEnv("GENERATE_ALL_E2E_ROUTES")
	if !exists || generate != "true" {
		return
	}

	// Build routes for echo ...
	err = cmd.GenerateSpecAndRoutes(arguments.CliArguments{ConfigPath: "./echo.e2e.gleece.config.json"})
	if err != nil {
		Fail("Failed to generate echo routes" + err.Error())
	}

	// // Build routes for Gorilla mux ...
	err = cmd.GenerateSpecAndRoutes(arguments.CliArguments{ConfigPath: "./mux.e2e.gleece.config.json"})
	if err != nil {
		Fail("Failed to generate echo routes" + err.Error())
	}
}

var _ = BeforeSuite(func() {

	RegenerateRoutes()
	// Init routes

	// Set Gin
	gin.SetMode(gin.TestMode)
	ginRouter = gin.Default()
	gleeceGinRoutes.RegisterRoutes(ginRouter)

	// Set Echo
	echoRouter = echo.New()
	echoRouter.Use(middleware.Recover())
	gleeceEchoRoutes.RegisterRoutes(echoRouter)

	// Set Gorilla mux
	muxRouter = mux.NewRouter()
	gleeceMuxRoutes.RegisterRoutes(muxRouter)
})

type RouterTest struct {
	Name                string
	Path                string
	Method              string
	Body                any
	Query               map[string]string
	Headers             map[string]string
	ExpectedStatus      int
	ExpectedBody        string
	ExpectedBodyContain string
	ExpendedHeaders     map[string]string
}

type RouterTestResult struct {
	Code    int
	Body    string
	Headers map[string]string
}

func GinRouterTest(routerTest RouterTest) RouterTestResult {
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

	ginRouter.ServeHTTP(w, req)

	// Convert response headers to map[string]string
	headers := make(map[string]string)
	for k, v := range w.Header() {
		if len(v) > 0 {
			headers[strings.ToLower(k)] = v[0]
		}
	}

	return RouterTestResult{
		Code:    w.Code,
		Body:    w.Body.String(),
		Headers: headers,
	}
}

func EchoRouterTest(routerTest RouterTest) RouterTestResult {
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

	echoRouter.ServeHTTP(w, req)

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
	return RouterTestResult{
		Code:    w.Code,
		Body:    bodyRes,
		Headers: headers,
	}
}

func MuxRouterTest(routerTest RouterTest) RouterTestResult {
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
	muxRouter.ServeHTTP(w, req)

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
	return RouterTestResult{
		Code:    w.Code,
		Body:    bodyRes,
		Headers: headers,
	}
}

func VerifyResult(result RouterTestResult, routerTest RouterTest) {
	Expect(result.Code).To(Equal(routerTest.ExpectedStatus))
	if routerTest.ExpectedBodyContain != "" {
		Expect(result.Body).To(ContainSubstring(routerTest.ExpectedBodyContain))
	} else if routerTest.ExpectedBody != "" {
		Expect(result.Body).To(Equal(routerTest.ExpectedBody))
	}
	if routerTest.ExpendedHeaders != nil {
		for k, v := range routerTest.ExpendedHeaders {
			Expect(result.Headers[strings.ToLower(k)]).To(Equal(v))
		}
	}
}

func RunRouterTest(routerTest RouterTest) {
	ginResponse := GinRouterTest(routerTest)
	VerifyResult(ginResponse, routerTest)

	echoResponse := EchoRouterTest(routerTest)
	VerifyResult(echoResponse, routerTest)

	muxResponse := MuxRouterTest(routerTest)
	VerifyResult(muxResponse, routerTest)
}
