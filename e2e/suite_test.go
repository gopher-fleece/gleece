package e2e

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/mux"

	"github.com/gopher-fleece/runtime"

	"github.com/gopher-fleece/gleece/cmd"
	"github.com/gopher-fleece/gleece/cmd/arguments"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/e2e/common"

	gleeceChiRoutesExExtra "github.com/gopher-fleece/gleece/e2e/chi/ex_extra_routes"
	gleeceChiRoutes "github.com/gopher-fleece/gleece/e2e/chi/routes"

	gleeceEchoRoutesExExtra "github.com/gopher-fleece/gleece/e2e/echo/ex_extra_routes"
	gleeceEchoRoutes "github.com/gopher-fleece/gleece/e2e/echo/routes"

	gleeceFiberRoutesExExtra "github.com/gopher-fleece/gleece/e2e/fiber/ex_extra_routes"
	gleeceFiberRoutes "github.com/gopher-fleece/gleece/e2e/fiber/routes"

	gleeceGinRoutesExExtra "github.com/gopher-fleece/gleece/e2e/gin/ex_extra_routes"
	gleeceGinRoutes "github.com/gopher-fleece/gleece/e2e/gin/routes"

	gleeceMuxRoutesExExtra "github.com/gopher-fleece/gleece/e2e/mux/ex_extra_routes"
	gleeceMuxRoutes "github.com/gopher-fleece/gleece/e2e/mux/routes"

	chiTester "github.com/gopher-fleece/gleece/e2e/chi"
	echoTester "github.com/gopher-fleece/gleece/e2e/echo"
	fiberTester "github.com/gopher-fleece/gleece/e2e/fiber"
	ginTester "github.com/gopher-fleece/gleece/e2e/gin"
	muxTester "github.com/gopher-fleece/gleece/e2e/mux"

	e2eAssets "github.com/gopher-fleece/gleece/e2e/assets"
	chiMiddlewares "github.com/gopher-fleece/gleece/e2e/chi/middlewares"
	echoMiddlewares "github.com/gopher-fleece/gleece/e2e/echo/middlewares"
	fiberMiddlewares "github.com/gopher-fleece/gleece/e2e/fiber/middlewares"
	ginMiddlewares "github.com/gopher-fleece/gleece/e2e/gin/middlewares"
	muxMiddlewares "github.com/gopher-fleece/gleece/e2e/mux/middlewares"

	"github.com/gofiber/fiber/v2"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gopher-fleece/gleece/generator/routes"
	"github.com/gopher-fleece/gleece/generator/swagen"
)

var exExtraRouting = common.RunOnVanillaRoutes
var fullyFeaturedRouting = common.RunOnFullyFeaturedRoutes
var allRouting = common.RunOnAllRoutes

func init() {
	// Force-set the test timeout via environment variable
	// This affects the Go test runtime directly
	os.Setenv("GO_TEST_TIMEOUT", "5m")

	// Set up a failsafe timer that will force-exit if tests take too long
	go func() {
		time.Sleep(5 * time.Minute)
		fmt.Println("⚠️ Test exceeded 5-minute timeout - force terminating")
		os.Exit(1)
	}()
}

func TestGleeceE2E(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)

	// Configure Ginkgo with a longer default timeout
	suiteConfig, reporterConfig := GinkgoConfiguration()
	suiteConfig.Timeout = 5 * time.Minute

	RegisterFailHandler(Fail)
	RunSpecs(t, "Gleece E2E Suite", suiteConfig, reporterConfig)
}

func GenerateE2ERoutes(args arguments.CliArguments, engineName string) error {
	logger.Info("Generating spec and routes")
	config, meta, err := cmd.GetConfigAndMetadata(args)
	if err != nil {
		return err
	}

	// // Generate the routes first
	if err := routes.GenerateRoutes(config, meta); err != nil {
		logger.Fatal("Failed to generate routes - %v", err)
		return err
	}

	// Off experimental feature
	config.ExperimentalConfig = definitions.ExperimentalConfig{}

	config.RoutesConfig.TemplateOverrides = nil
	config.RoutesConfig.TemplateExtensions = nil
	config.RoutesConfig.AuthorizationConfig.EnforceSecurityOnAllRoutes = false
	config.RoutesConfig.ValidateResponsePayload = false
	config.RoutesConfig.SkipGenerateDateComment = false

	config.RoutesConfig.OutputPath = fmt.Sprintf("./%s/ex_extra_routes/%s.e2e.ex_extra.gleece.go", engineName, engineName)
	config.RoutesConfig.PackageName = "ex_extra_routes"

	// Generate the routes for ex_extra - a vanilla version of the routes without any customizations amd as default as possible
	if err := routes.GenerateRoutes(config, meta); err != nil {
		logger.Fatal("Failed to generate ex_extra_routes - %v", err)
		return err
	}

	// Generate the OpenAPI 3.0.0 spec
	config.OpenAPIGeneratorConfig.SpecGeneratorConfig.OutputPath = fmt.Sprintf("./%s/openapi/openapi3.0.0.json", engineName)
	if err := swagen.GenerateAndOutputSpec(
		&config.OpenAPIGeneratorConfig,
		meta.Flat,
		&meta.Models,
		meta.PlainErrorPresent,
	); err != nil {
		logger.Fatal("Failed to generate OpenAPI spec - %v", err)
		return err
	}

	// Generate the OpenAPI 3.1.0 spec
	config.OpenAPIGeneratorConfig.Info.Version = "3.1.0"
	config.OpenAPIGeneratorConfig.SpecGeneratorConfig.OutputPath = fmt.Sprintf("./%s/openapi/openapi3.1.0.json", engineName)
	if err := swagen.GenerateAndOutputSpec(
		&config.OpenAPIGeneratorConfig,
		meta.Flat,
		&meta.Models,
		meta.PlainErrorPresent,
	); err != nil {
		logger.Fatal("Failed to generate OpenAPI spec - %v", err)
		return err
	}

	logger.Info("Spec and routes successfully generated")
	return nil
}

func RegenerateRoutes() {

	var err error
	// Always build routes for gin  ...
	err = GenerateE2ERoutes(arguments.CliArguments{ConfigPath: "./e2e.gin.gleece.config.json"}, "gin")
	if err != nil {
		Fail("Failed to generate gin routes - \n" + err.Error())
	}

	// Get from env var whenever to regenerate all routes again.
	// Use it only when modifying the templates which requires new routes for tests for all other engines too.
	generate, exists := os.LookupEnv("GENERATE_ALL_E2E_ROUTES")
	if !exists || generate != "true" {
		return
	}

	// Build routes for echo ...
	err = GenerateE2ERoutes(arguments.CliArguments{ConfigPath: "./e2e.echo.gleece.config.json"}, "echo")
	if err != nil {
		Fail("Failed to generate echo routes - " + err.Error())
	}

	// Build routes for Gorilla mux ...
	err = GenerateE2ERoutes(arguments.CliArguments{ConfigPath: "./e2e.mux.gleece.config.json"}, "mux")
	if err != nil {
		Fail("Failed to generate mux routes - " + err.Error())
	}

	// Build routes for chi ...
	err = GenerateE2ERoutes(arguments.CliArguments{ConfigPath: "./e2e.chi.gleece.config.json"}, "chi")
	if err != nil {
		Fail("Failed to generate chi routes - " + err.Error())
	}

	// Build routes for Fiber ...
	err = GenerateE2ERoutes(arguments.CliArguments{ConfigPath: "./e2e.fiber.gleece.config.json"}, "fiber")
	if err != nil {
		Fail("Failed to generate fiber routes - " + err.Error())
	}
}

var _ = BeforeSuite(func() {
	// Generate routes
	RegenerateRoutes()

	// Init routers

	// Set Gin
	gin.SetMode(gin.TestMode)
	ginTester.GinRouter = gin.Default()
	gleeceGinRoutes.RegisterRoutes(ginTester.GinRouter)
	gleeceGinRoutes.RegisterMiddleware(runtime.BeforeOperation, ginMiddlewares.MiddlewareBeforeOperation)
	gleeceGinRoutes.RegisterMiddleware(runtime.AfterOperationSuccess, ginMiddlewares.MiddlewareAfterOperationSuccess)
	gleeceGinRoutes.RegisterErrorMiddleware(runtime.OnOperationError, ginMiddlewares.MiddlewareOnError)
	gleeceGinRoutes.RegisterErrorMiddleware(runtime.OnOperationError, ginMiddlewares.MiddlewareOnError2)
	gleeceGinRoutes.RegisterErrorMiddleware(runtime.OnInputValidationError, ginMiddlewares.MiddlewareOnValidationError)
	gleeceGinRoutes.RegisterErrorMiddleware(runtime.OnOutputValidationError, ginMiddlewares.MiddlewareOnOutputValidationError)
	gleeceGinRoutes.RegisterCustomValidator("validate_starts_with_letter", e2eAssets.ValidateStartsWithLetter)

	ginTester.GinExExtraRouter = gin.Default()
	gleeceGinRoutesExExtra.RegisterRoutes(ginTester.GinExExtraRouter)

	// Set Echo
	echoTester.EchoRouter = echo.New()
	echoTester.EchoRouter.Use(middleware.Recover())
	gleeceEchoRoutes.RegisterMiddleware(runtime.BeforeOperation, echoMiddlewares.MiddlewareBeforeOperation)
	gleeceEchoRoutes.RegisterMiddleware(runtime.AfterOperationSuccess, echoMiddlewares.MiddlewareAfterOperationSuccess)
	gleeceEchoRoutes.RegisterErrorMiddleware(runtime.OnOperationError, echoMiddlewares.MiddlewareOnError)
	gleeceEchoRoutes.RegisterErrorMiddleware(runtime.OnOperationError, echoMiddlewares.MiddlewareOnError2)
	gleeceEchoRoutes.RegisterErrorMiddleware(runtime.OnInputValidationError, echoMiddlewares.MiddlewareOnValidationError)
	gleeceEchoRoutes.RegisterErrorMiddleware(runtime.OnOutputValidationError, echoMiddlewares.MiddlewareOnOutputValidationError)
	gleeceEchoRoutes.RegisterCustomValidator("validate_starts_with_letter", e2eAssets.ValidateStartsWithLetter)
	gleeceEchoRoutes.RegisterRoutes(echoTester.EchoRouter)

	echoTester.EchoExExtraRouter = echo.New()
	gleeceEchoRoutesExExtra.RegisterRoutes(echoTester.EchoExExtraRouter)

	// Set Gorilla mux
	muxTester.MuxRouter = mux.NewRouter()
	gleeceMuxRoutes.RegisterMiddleware(runtime.BeforeOperation, muxMiddlewares.MiddlewareBeforeOperation)
	gleeceMuxRoutes.RegisterMiddleware(runtime.AfterOperationSuccess, muxMiddlewares.MiddlewareAfterOperationSuccess)
	gleeceMuxRoutes.RegisterErrorMiddleware(runtime.OnOperationError, muxMiddlewares.MiddlewareOnError)
	gleeceMuxRoutes.RegisterErrorMiddleware(runtime.OnOperationError, muxMiddlewares.MiddlewareOnError2)
	gleeceMuxRoutes.RegisterErrorMiddleware(runtime.OnInputValidationError, muxMiddlewares.MiddlewareOnValidationError)
	gleeceMuxRoutes.RegisterErrorMiddleware(runtime.OnOutputValidationError, muxMiddlewares.MiddlewareOnOutputValidationError)
	gleeceMuxRoutes.RegisterCustomValidator("validate_starts_with_letter", e2eAssets.ValidateStartsWithLetter)
	gleeceMuxRoutes.RegisterRoutes(muxTester.MuxRouter)

	muxTester.MuxExExtraRouter = mux.NewRouter()
	gleeceMuxRoutesExExtra.RegisterRoutes(muxTester.MuxExExtraRouter)

	// Set Chi
	chiTester.ChiRouter = chi.NewRouter()
	gleeceChiRoutes.RegisterMiddleware(runtime.BeforeOperation, chiMiddlewares.MiddlewareBeforeOperation)
	gleeceChiRoutes.RegisterMiddleware(runtime.AfterOperationSuccess, chiMiddlewares.MiddlewareAfterOperationSuccess)
	gleeceChiRoutes.RegisterErrorMiddleware(runtime.OnOperationError, chiMiddlewares.MiddlewareOnError)
	gleeceChiRoutes.RegisterErrorMiddleware(runtime.OnOperationError, chiMiddlewares.MiddlewareOnError2)
	gleeceChiRoutes.RegisterErrorMiddleware(runtime.OnInputValidationError, chiMiddlewares.MiddlewareOnValidationError)
	gleeceChiRoutes.RegisterErrorMiddleware(runtime.OnOutputValidationError, chiMiddlewares.MiddlewareOnOutputValidationError)
	gleeceChiRoutes.RegisterCustomValidator("validate_starts_with_letter", e2eAssets.ValidateStartsWithLetter)
	gleeceChiRoutes.RegisterRoutes(chiTester.ChiRouter)

	chiTester.ChiExExtraRouter = chi.NewRouter()
	gleeceChiRoutesExExtra.RegisterRoutes(chiTester.ChiExExtraRouter)

	// Set Fiber
	fiberTester.FiberRouter = fiber.New()
	gleeceFiberRoutes.RegisterMiddleware(runtime.BeforeOperation, fiberMiddlewares.MiddlewareBeforeOperation)
	gleeceFiberRoutes.RegisterMiddleware(runtime.AfterOperationSuccess, fiberMiddlewares.MiddlewareAfterOperationSuccess)
	gleeceFiberRoutes.RegisterErrorMiddleware(runtime.OnOperationError, fiberMiddlewares.MiddlewareOnError)
	gleeceFiberRoutes.RegisterErrorMiddleware(runtime.OnOperationError, fiberMiddlewares.MiddlewareOnError2)
	gleeceFiberRoutes.RegisterErrorMiddleware(runtime.OnInputValidationError, fiberMiddlewares.MiddlewareOnValidationError)
	gleeceFiberRoutes.RegisterErrorMiddleware(runtime.OnOutputValidationError, fiberMiddlewares.MiddlewareOnOutputValidationError)
	gleeceFiberRoutes.RegisterCustomValidator("validate_starts_with_letter", e2eAssets.ValidateStartsWithLetter)
	gleeceFiberRoutes.RegisterRoutes(fiberTester.FiberRouter)

	fiberTester.FiberExExtraRouter = fiber.New()
	gleeceFiberRoutesExExtra.RegisterRoutes(fiberTester.FiberExExtraRouter)
})

func VerifyResult(result common.RouterTestResult, routerTest common.RouterTest) {
	Expect(routerTest.ExpectedStatus).To(Equal(result.Code))
	if routerTest.ExpectedBodyContain != "" {
		Expect(result.Body).To(ContainSubstring(routerTest.ExpectedBodyContain))
	} else if routerTest.ExpectedBody != "" {
		Expect(result.Body).To(Equal(routerTest.ExpectedBody))
	}
	if routerTest.ExpendedHeaders != nil {
		for k, v := range routerTest.ExpendedHeaders {
			rawValue := result.Headers[strings.ToLower(k)]
			// Split by ; and check if all values are present in the response header
			value := strings.Split(rawValue, ";")[0]
			Expect(value).To(Equal(v))
		}
	}
}

func runTest(routerTest common.RouterTest) {
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

func RunRouterTest(routerTest common.RouterTest) {
	routesFlavors := []common.RunningMode{}

	if routerTest.RunningMode == nil {
		routesFlavors = []common.RunningMode{common.RunOnVanillaRoutes, common.RunOnFullyFeaturedRoutes}
	} else {
		switch *routerTest.RunningMode {
		case common.RunOnFullyFeaturedRoutes:
			routesFlavors = []common.RunningMode{common.RunOnFullyFeaturedRoutes}
		case common.RunOnVanillaRoutes:
			routesFlavors = []common.RunningMode{common.RunOnVanillaRoutes}
		case common.RunOnAllRoutes:
			routesFlavors = []common.RunningMode{common.RunOnFullyFeaturedRoutes, common.RunOnVanillaRoutes}
		}
	}

	for _, flavor := range routesFlavors {
		routerTest.RunningMode = &flavor
		runTest(routerTest)
	}
}
