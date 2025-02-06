package e2e

import (
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/mux"

	"github.com/gopher-fleece/gleece/cmd"
	"github.com/gopher-fleece/gleece/cmd/arguments"
	gleeceChiRoutes "github.com/gopher-fleece/gleece/e2e/chi/routes"
	gleeceEchoRoutes "github.com/gopher-fleece/gleece/e2e/echo/routes"
	gleeceFiberRoutes "github.com/gopher-fleece/gleece/e2e/fiber/routes"
	gleeceGinRoutes "github.com/gopher-fleece/gleece/e2e/gin/routes"
	gleeceMuxRoutes "github.com/gopher-fleece/gleece/e2e/mux/routes"

	chiTester "github.com/gopher-fleece/gleece/e2e/chi"
	echoTester "github.com/gopher-fleece/gleece/e2e/echo"
	fiberTester "github.com/gopher-fleece/gleece/e2e/fiber"
	ginTester "github.com/gopher-fleece/gleece/e2e/gin"
	muxTester "github.com/gopher-fleece/gleece/e2e/mux"

	"github.com/gopher-fleece/gleece/e2e/common"

	"github.com/gofiber/fiber/v2"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/labstack/echo/v4"
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

func RegenerateRoutes() {

	// Always build routes for gin  ...
	err := cmd.GenerateSpecAndRoutes(arguments.CliArguments{ConfigPath: "./e2e.gin.gleece.config.json"})
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
	err = cmd.GenerateSpecAndRoutes(arguments.CliArguments{ConfigPath: "./e2e.echo.gleece.config.json"})
	if err != nil {
		Fail("Failed to generate echo routes" + err.Error())
	}

	// Build routes for Gorilla mux ...
	err = cmd.GenerateSpecAndRoutes(arguments.CliArguments{ConfigPath: "./e2e.mux.gleece.config.json"})
	if err != nil {
		Fail("Failed to generate echo routes" + err.Error())
	}

	// Build routes for chi ...
	err = cmd.GenerateSpecAndRoutes(arguments.CliArguments{ConfigPath: "./e2e.chi.gleece.config.json"})
	if err != nil {
		Fail("Failed to generate echo routes" + err.Error())
	}

	// Build routes for Fiber ...
	err = cmd.GenerateSpecAndRoutes(arguments.CliArguments{ConfigPath: "./e2e.fiber.gleece.config.json"})
	if err != nil {
		Fail("Failed to generate echo routes" + err.Error())
	}
}

var _ = BeforeSuite(func() {
	RegenerateRoutes()
	// Init routes

	// Set Gin
	gin.SetMode(gin.TestMode)
	ginTester.GinRouter = gin.Default()
	gleeceGinRoutes.RegisterRoutes(ginTester.GinRouter)

	// Set Echo
	echoTester.EchoRouter = echo.New()
	echoTester.EchoRouter.Use(middleware.Recover())
	gleeceEchoRoutes.RegisterRoutes(echoTester.EchoRouter)

	// Set Gorilla mux
	muxTester.MuxRouter = mux.NewRouter()
	gleeceMuxRoutes.RegisterRoutes(muxTester.MuxRouter)

	// Set Chi
	chiTester.ChiRouter = chi.NewRouter()
	gleeceChiRoutes.RegisterRoutes(chiTester.ChiRouter)

	// Set Fiber
	fiberTester.FiberRouter = fiber.New()
	gleeceFiberRoutes.RegisterRoutes(fiberTester.FiberRouter)
})

func VerifyResult(result common.RouterTestResult, routerTest common.RouterTest) {
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

func RunRouterTest(routerTest common.RouterTest) {
	ginResponse := ginTester.GinRouterTest(routerTest)
	VerifyResult(ginResponse, routerTest)

	echoResponse := echoTester.EchoRouterTest(routerTest)
	VerifyResult(echoResponse, routerTest)

	muxResponse := muxTester.MuxRouterTest(routerTest)
	VerifyResult(muxResponse, routerTest)

	chiResponse := chiTester.ChiRouterTest(routerTest)
	VerifyResult(chiResponse, routerTest)

	fiberResponse := fiberTester.FiberRouterTest(routerTest)
	VerifyResult(fiberResponse, routerTest)
}
