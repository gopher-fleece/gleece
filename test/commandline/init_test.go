package commandline_test

import (
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/gopher-fleece/gleece/v2/cmd"
	"github.com/gopher-fleece/gleece/v2/definitions"
	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
	"github.com/gopher-fleece/gleece/v2/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Init Command", Serial, func() {
	var originalWd string

	BeforeEach(func() {
		var err error
		originalWd, err = os.Getwd()
		Expect(err).To(BeNil())

		Expect(os.MkdirAll("./dist", 0o755)).To(Succeed())
		Expect(os.Chdir("./dist")).To(Succeed())
	})

	AfterEach(func() {
		Expect(os.Chdir(originalWd)).To(Succeed())
		utils.DeleteDistInCurrentFolderOrFail()
		logger.SetLogLevel(logger.LogLevelNone)
		cmd.Reset()
	})

	It("Creates a default configuration file with defaults and confirms generation", func() {
		stdout, stderr, err := runInitWithInput(getLineForDefaultConfig())
		Expect(err).To(BeNil())
		Expect(stderr).To(BeEmpty())
		Expect(stdout).To(ContainSubstring("Successfully created gleece.config.json"))

		cfg := readInitConfig("./gleece.config.json")
		Expect(cfg.CommonConfig.ControllerGlobs).To(Equal([]string{"./**/controllers/**/*.go"}))
		Expect(cfg.CommonConfig.AllowPackageLoadFailures).To(BeFalse())
		Expect(cfg.RoutesConfig.Engine).To(Equal(definitions.RoutingEngineGin))
		Expect(cfg.RoutesConfig.PackageName).To(Equal("routes"))
		Expect(cfg.RoutesConfig.OutputPath).To(Equal("./routes/gleece.go"))
		Expect(cfg.RoutesConfig.OutputFilePerms).To(Equal("644"))
		Expect(cfg.RoutesConfig.ValidateResponsePayload).To(BeFalse())
		Expect(cfg.RoutesConfig.SkipGenerateDateComment).To(BeTrue())
		Expect(cfg.RoutesConfig.AuthorizationConfig.AuthFileFullPackageName).To(Equal("authentication"))
		Expect(cfg.RoutesConfig.AuthorizationConfig.EnforceSecurityOnAllRoutes).To(BeTrue())
		Expect(cfg.OpenAPIGeneratorConfig.OpenAPI).To(Equal("3.0.0"))
		Expect(cfg.OpenAPIGeneratorConfig.Info.Title).To(Equal("My API"))
		Expect(cfg.OpenAPIGeneratorConfig.Info.Version).To(Equal("1.0.0"))
		Expect(cfg.OpenAPIGeneratorConfig.BaseURL).To(Equal("http://localhost:8080"))
		Expect(cfg.OpenAPIGeneratorConfig.SpecGeneratorConfig.OutputPath).To(Equal("./docs/swagger.json"))
		Expect(cfg.ExperimentalConfig.ValidateTopLevelOnlyEnum).To(BeFalse())
		Expect(cfg.ExperimentalConfig.GenerateEnumValidator).To(BeFalse())
	})

	It("Rejects invalid values and re-prompts until valid input is provided", func() {
		input := lines(
			"",          // Controller globs
			"",          // Allow package load failures
			"",          // Routing engine
			"",          // Routes package name
			"",          // Routes output path
			"bad-perms", // Invalid file permissions
			"0640",      // Valid file permissions
			"",          // Validate response payload
			"",          // Skip generation date comment
			"",          // Auth middleware package name
			"",          // Enforce security on all routes
			"",          // OpenAPI version
			"",          // API title
			"",          // API description
			"",          // API version
			"",          // Terms of service URL
			"",          // Include contact information?
			"",          // Include license information?
			"",          // Base URL
			"",          // OpenAPI spec output path
			"",          // Add a security scheme?
			"",          // Validate top-level only enums?
			"",          // Generate enum validator?
			"n",         // Generate authentication middleware skeleton code?
		)

		stdout, stderr, err := runInitWithInput(input)
		Expect(err).To(BeNil())
		Expect(stderr).To(BeEmpty())
		Expect(stdout).To(ContainSubstring("Invalid file permissions format. Please use an octal number like 0644 or 0777."))
		Expect(stdout).To(ContainSubstring("Successfully created gleece.config.json"))

		cfg := readInitConfig("./gleece.config.json")
		Expect(cfg.RoutesConfig.OutputFilePerms).To(Equal("640"))
	})

	It("Uses user-provided values and validates formatted inputs", func() {
		input := lines(
			"src/controllers/**/*.go",
			"y",
			"echo",
			"api",
			"./custom/routes.go",
			"0755",
			"y",
			"n",
			"custom.auth",
			"n",
			"3.1.0",
			"My Custom API",
			"My custom description",
			"2.0.0",
			"not-a-url",
			"https://example.com/tos",
			"y",
			"Jane Doe",
			"not-a-url",
			"https://example.com",
			"invalid-email",
			"jane@example.com",
			"y",
			"MIT",
			"not-a-url",
			"https://example.com/lic",
			"https://example.com/api",
			"./docs/openapi.json",
			"n",
			"y",
			"y",
			"n",
		)

		stdout, stderr, err := runInitWithInput(input)
		Expect(err).To(BeNil())
		Expect(stderr).To(BeEmpty())
		Expect(stdout).To(ContainSubstring("Invalid URL format."))
		Expect(stdout).To(ContainSubstring("Invalid email format."))

		cfg := readInitConfig("./gleece.config.json")
		Expect(cfg.CommonConfig.ControllerGlobs).To(Equal([]string{"src/controllers/**/*.go"}))
		Expect(cfg.CommonConfig.AllowPackageLoadFailures).To(BeTrue())
		Expect(cfg.RoutesConfig.Engine).To(Equal(definitions.RoutingEngineEcho))
		Expect(cfg.RoutesConfig.PackageName).To(Equal("api"))
		Expect(cfg.RoutesConfig.OutputPath).To(Equal("./custom/routes.go"))
		Expect(cfg.RoutesConfig.OutputFilePerms).To(Equal("755"))
		Expect(cfg.RoutesConfig.ValidateResponsePayload).To(BeTrue())
		Expect(cfg.RoutesConfig.SkipGenerateDateComment).To(BeFalse())
		Expect(cfg.RoutesConfig.AuthorizationConfig.AuthFileFullPackageName).To(Equal("custom.auth"))
		Expect(cfg.RoutesConfig.AuthorizationConfig.EnforceSecurityOnAllRoutes).To(BeFalse())
		Expect(cfg.OpenAPIGeneratorConfig.OpenAPI).To(Equal("3.1.0"))
		Expect(cfg.OpenAPIGeneratorConfig.Info.Title).To(Equal("My Custom API"))
		Expect(cfg.OpenAPIGeneratorConfig.Info.Description).To(Equal("My custom description"))
		Expect(cfg.OpenAPIGeneratorConfig.Info.Version).To(Equal("2.0.0"))
		Expect(cfg.OpenAPIGeneratorConfig.Info.TermsOfService).To(Equal("https://example.com/tos"))
		Expect(cfg.OpenAPIGeneratorConfig.Info.Contact).ToNot(BeNil())
		Expect(cfg.OpenAPIGeneratorConfig.Info.Contact.Name).To(Equal("Jane Doe"))
		Expect(cfg.OpenAPIGeneratorConfig.Info.Contact.URL).To(Equal("https://example.com"))
		Expect(cfg.OpenAPIGeneratorConfig.Info.Contact.Email).To(Equal("jane@example.com"))
		Expect(cfg.OpenAPIGeneratorConfig.Info.License).ToNot(BeNil())
		Expect(cfg.OpenAPIGeneratorConfig.Info.License.Name).To(Equal("MIT"))
		Expect(cfg.OpenAPIGeneratorConfig.Info.License.URL).To(Equal("https://example.com/lic"))
		Expect(cfg.OpenAPIGeneratorConfig.BaseURL).To(Equal("https://example.com/api"))
		Expect(cfg.OpenAPIGeneratorConfig.SpecGeneratorConfig.OutputPath).To(Equal("./docs/openapi.json"))
		Expect(cfg.ExperimentalConfig.ValidateTopLevelOnlyEnum).To(BeTrue())
		Expect(cfg.ExperimentalConfig.GenerateEnumValidator).To(BeTrue())
	})

	It("Exercises overwrite confirmation invalid, reject, and accept branches", func() {
		// First run to create the config file.
		createStdout, createStderr, createErr := runInitWithInput(getLineForDefaultConfig())
		Expect(createErr).To(BeNil())
		Expect(createStderr).To(BeEmpty())
		Expect(createStdout).To(ContainSubstring("Successfully created gleece.config.json"))

		// Second run: invalid confirmation first, then reject with 'n'.
		rejectInput := lines(
			"",      // Controller globs
			"",      // Allow package load failures
			"",      // Routing engine
			"",      // Routes package name
			"",      // Routes output path
			"",      // Output file permissions
			"",      // Validate response payload
			"",      // Skip generation date comment
			"",      // Auth middleware package name
			"",      // Enforce security on all routes
			"",      // OpenAPI version
			"",      // API title
			"",      // API description
			"",      // API version
			"",      // Terms of service URL
			"",      // Include contact information?
			"",      // Include license information?
			"",      // Base URL
			"",      // OpenAPI spec output path
			"",      // Add a security scheme?
			"",      // Validate top-level only enums?
			"",      // Generate enum validator?
			"maybe", // Invalid overwrite confirmation
			"n",     // Reject overwrite
			"n",     // Generate authentication middleware skeleton code?
		)

		rejectStdout, rejectStderr, rejectErr := runInitWithInput(rejectInput)
		Expect(rejectErr).To(BeNil())
		Expect(rejectStderr).To(BeEmpty())
		Expect(rejectStdout).To(ContainSubstring("File 'gleece.config.json' already exists. Overwrite? (y/n): "))
		Expect(rejectStdout).To(ContainSubstring("Please confirm or reject by typing 'y' or 'n' and pressing enter"))
		Expect(rejectStdout).To(ContainSubstring("Configuration overwrite aborted"))

		// Third run: invalid confirmation first, then accept with 'y'.
		acceptInput := lines(
			"",      // Controller globs
			"",      // Allow package load failures
			"",      // Routing engine
			"",      // Routes package name
			"",      // Routes output path
			"",      // Output file permissions
			"",      // Validate response payload
			"",      // Skip generation date comment
			"",      // Auth middleware package name
			"",      // Enforce security on all routes
			"",      // OpenAPI version
			"",      // API title
			"",      // API description
			"",      // API version
			"",      // Terms of service URL
			"",      // Include contact information?
			"",      // Include license information?
			"",      // Base URL
			"",      // OpenAPI spec output path
			"",      // Add a security scheme?
			"",      // Validate top-level only enums?
			"",      // Generate enum validator?
			"maybe", // Invalid overwrite confirmation
			"y",     // Accept overwrite
			"n",     // Generate authentication middleware skeleton code?
		)

		acceptStdout, acceptStderr, acceptErr := runInitWithInput(acceptInput)
		Expect(acceptErr).To(BeNil())
		Expect(acceptStderr).To(BeEmpty())
		Expect(acceptStdout).To(ContainSubstring("File 'gleece.config.json' already exists. Overwrite? (y/n): "))
		Expect(acceptStdout).To(ContainSubstring("Please confirm or reject by typing 'y' or 'n' and pressing enter"))
		Expect(acceptStdout).To(ContainSubstring("Successfully created gleece.config.json"))
	})
})

func readInitConfig(path string) definitions.GleeceConfig {
	data, err := os.ReadFile(path)
	Expect(err).To(BeNil())

	var cfg definitions.GleeceConfig
	Expect(json.Unmarshal(data, &cfg)).To(Succeed())

	return cfg
}

func runInitWithInput(input string) (stdout string, stderr string, err error) {
	originalStdin := os.Stdin
	originalStdout := os.Stdout

	stdinR, stdinW, err := os.Pipe()
	Expect(err).To(BeNil())

	stdoutR, stdoutW, err := os.Pipe()
	Expect(err).To(BeNil())

	var stdoutBuffer strings.Builder
	stdoutDone := make(chan error, 1)

	go func() {
		_, copyErr := io.Copy(&stdoutBuffer, stdoutR)
		stdoutDone <- copyErr
	}()

	_, err = io.WriteString(stdinW, input)
	Expect(err).To(BeNil())
	Expect(stdinW.Close()).To(BeNil())

	os.Stdin = stdinR
	os.Stdout = stdoutW

	result := cmd.ExecuteWithArgs([]string{"init", "--no-banner"})

	_ = stdoutW.Close()
	Expect(<-stdoutDone).To(BeNil())

	os.Stdin = originalStdin
	os.Stdout = originalStdout

	_ = stdinR.Close()
	_ = stdoutR.Close()

	return stdoutBuffer.String(), result.StdErr, result.Error
}

func lines(values ...string) string {
	return strings.Join(values, "\n") + "\n"
}

func getLineForDefaultConfig() string {
	return lines(
		"",  // Controller globs
		"",  // Allow package load failures
		"",  // Routing engine
		"",  // Routes package name
		"",  // Routes output path
		"",  // Output file permissions
		"",  // Validate response payload
		"",  // Skip generation date comment
		"",  // Auth middleware package name
		"",  // Enforce security on all routes
		"",  // OpenAPI version
		"",  // API title
		"",  // API description
		"",  // API version
		"",  // Terms of service URL
		"",  // Include contact information?
		"",  // Include license information?
		"",  // Base URL
		"",  // OpenAPI spec output path
		"",  // Add a security scheme?
		"",  // Validate top-level only enums?
		"",  // Generate enum validator?
		"n", // Generate authentication middleware skeleton code?
	)
}
