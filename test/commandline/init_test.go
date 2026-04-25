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
		input := getInput(Questionnaire{
			OutputFilePerms: []string{"bad-perms", "0640"},
		})

		stdout, stderr, err := runInitWithInput(input)
		Expect(err).To(BeNil())
		Expect(stderr).To(BeEmpty())
		Expect(stdout).To(ContainSubstring("Invalid file permissions format. Please use an octal number like 0644 or 0777."))
		Expect(stdout).To(ContainSubstring("Successfully created gleece.config.json"))

		cfg := readInitConfig("./gleece.config.json")
		Expect(cfg.RoutesConfig.OutputFilePerms).To(Equal("640"))
	})

	It("Uses user-provided values and validates formatted inputs", func() {
		input := getInput(Questionnaire{
			ControllerGlobs:           []string{"src/controllers/**/*.go"},
			AllowPackageLoadFailures:  []string{"y"},
			RoutingEngine:             []string{"echo"},
			RoutesPackageName:         []string{"api"},
			RoutesOutputPath:          []string{"./custom/routes.go"},
			OutputFilePerms:           []string{"0755"},
			ValidateResponsePayload:   []string{"y"},
			SkipGenerateDateComment:   []string{"n"},
			AuthMiddlewarePackageName: []string{"custom.auth"},
			EnforceSecurity:           []string{"n"},
			OpenAPIVersion:            []string{"3.1.0"},
			APITitle:                  []string{"My Custom API"},
			APIDescription:            []string{"My custom description"},
			APIVersion:                []string{"2.0.0"},
			TermsOfServiceURL:         []string{"not-a-url", "https://example.com/tos"},
			IncludeContact:            []string{"y", "Jane Doe", "not-a-url", "https://example.com", "invalid-email", "jane@example.com"},
			IncludeLicense:            []string{"y", "MIT", "not-a-url", "https://example.com/lic"},
			BaseURL:                   []string{"https://example.com/api"},
			OpenAPISpecOutputPath:     []string{"./docs/openapi.json"},
			AddSecurityScheme:         []string{"n"},
			ValidateTopLevelOnlyEnum:  []string{"y"},
			GenerateEnumValidator:     []string{"y"},
			GenerateAuthMiddleware:    []string{"n"},
		})

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

	It("Adds multiple security schemes and validates fields", func() {
		input := getInput(Questionnaire{
			AddSecurityScheme: []string{"y"},
			SecuritySchemes: []SecuritySchemeInput{
				{
					"MyApiKey",
					"API Key description",
					"apiKey",
					"header",
					"X-API-KEY",
				},
				{
					"MyHttp",
					"HTTP description",
					"http",
					"basic",
				},
				{
					"MyOAuth2",
					"OAuth2 description",
					"oauth2",
					"y",
					"https://auth.com",
					"https://token.com",
					"https://refresh.com",
					"scope1:value1",
					"n",
					"n",
					"n",
				},
				{
					"MyOIDC",
					"OIDC description",
					"openIdConnect",
					"https://oidc.com",
				},
			},
			GenerateAuthMiddleware: []string{"n"},
		})

		out, _, err := runInitWithInput(input)
		Expect(out).ToNot(BeNil())
		Expect(err).To(BeNil())

		cfg := readInitConfig("./gleece.config.json")
		Expect(cfg.OpenAPIGeneratorConfig.SecuritySchemes).To(HaveLen(4))
	})

	It("Adds apiKey security scheme and validates fields", func() {
		input := getInput(Questionnaire{
			AddSecurityScheme: []string{"y"},
			SecuritySchemes: []SecuritySchemeInput{
				{"MyApiKey", "API Key description", "apiKey", "invalid-in", "query", "MyField"},
			},
			GenerateAuthMiddleware: []string{"n"},
		})

		stdout, _, err := runInitWithInput(input)
		Expect(err).To(BeNil())
		Expect(stdout).To(ContainSubstring("Invalid selection. Please choose from (query|header|cookie)"))

		cfg := readInitConfig("./gleece.config.json")
		Expect(cfg.OpenAPIGeneratorConfig.SecuritySchemes).To(HaveLen(1))
		Expect(cfg.OpenAPIGeneratorConfig.SecuritySchemes[0].Type).To(Equal(definitions.APIKey))
		Expect(cfg.OpenAPIGeneratorConfig.SecuritySchemes[0].In).To(Equal(definitions.SecuritySchemeIn("query")))
	})

	It("Adds http security scheme and validates fields", func() {
		input := getInput(Questionnaire{
			AddSecurityScheme: []string{"y"},
			SecuritySchemes: []SecuritySchemeInput{
				{"MyHttp", "HTTP description", "http", "invalid-scheme", "basic"},
			},
			GenerateAuthMiddleware: []string{"n"},
		})

		stdout, _, err := runInitWithInput(input)
		Expect(err).To(BeNil())
		Expect(stdout).To(ContainSubstring("Invalid selection. Please choose from"))

		cfg := readInitConfig("./gleece.config.json")
		Expect(cfg.OpenAPIGeneratorConfig.SecuritySchemes).To(HaveLen(1))
		Expect(cfg.OpenAPIGeneratorConfig.SecuritySchemes[0].Type).To(Equal(definitions.HTTP))
		Expect(cfg.OpenAPIGeneratorConfig.SecuritySchemes[0].Scheme).To(Equal(definitions.HttpAuthScheme("basic")))
	})

	It("Adds oauth2 security scheme and validates fields", func() {
		input := getInput(Questionnaire{
			AddSecurityScheme: []string{"y"},
			SecuritySchemes: []SecuritySchemeInput{
				{"MyOAuth2", "OAuth2 description", "oauth2", "n", "n", "n", "n"},
			},
			GenerateAuthMiddleware: []string{"n"},
		})

		_, _, err := runInitWithInput(input)
		Expect(err).To(BeNil())

		cfg := readInitConfig("./gleece.config.json")
		Expect(cfg.OpenAPIGeneratorConfig.SecuritySchemes).To(HaveLen(1))
		Expect(cfg.OpenAPIGeneratorConfig.SecuritySchemes[0].Type).To(Equal(definitions.OAuth2))
	})

	It("Adds oauth2 security scheme with invalid URLs", func() {
		input := getInput(Questionnaire{
			AddSecurityScheme: []string{"y"},
			SecuritySchemes: []SecuritySchemeInput{
				{"MyOAuth2", "OAuth2 description", "oauth2", "y", "not-a-url", "https://auth.com", "not-a-url", "https://token.com", "not-a-url", "https://refresh.com", "n"},
			},
			GenerateAuthMiddleware: []string{"n"},
		})

		stdout, _, err := runInitWithInput(input)
		Expect(err).To(BeNil())
		Expect(stdout).To(ContainSubstring("Invalid URL format."))

		cfg := readInitConfig("./gleece.config.json")
		Expect(cfg.OpenAPIGeneratorConfig.SecuritySchemes).To(HaveLen(1))
		Expect(cfg.OpenAPIGeneratorConfig.SecuritySchemes[0].Type).To(Equal(definitions.OAuth2))
		Expect(cfg.OpenAPIGeneratorConfig.SecuritySchemes[0].Flows.Implicit.AuthorizationURL).To(Equal("https://auth.com"))
		Expect(cfg.OpenAPIGeneratorConfig.SecuritySchemes[0].Flows.Implicit.TokenURL).To(Equal("https://token.com"))
		Expect(cfg.OpenAPIGeneratorConfig.SecuritySchemes[0].Flows.Implicit.RefreshURL).To(Equal("https://refresh.com"))
	})

	It("Adds openIdConnect security scheme and validates fields", func() {
		input := getInput(Questionnaire{
			AddSecurityScheme: []string{"y"},
			SecuritySchemes: []SecuritySchemeInput{
				{"MyOIDC", "OIDC description", "openIdConnect", "not-a-url", "https://oidc.com"},
			},
			GenerateAuthMiddleware: []string{"n"},
		})

		stdout, _, err := runInitWithInput(input)
		Expect(err).To(BeNil())
		Expect(stdout).To(ContainSubstring("Invalid URL format."))

		cfg := readInitConfig("./gleece.config.json")
		Expect(cfg.OpenAPIGeneratorConfig.SecuritySchemes).To(HaveLen(1))
		Expect(cfg.OpenAPIGeneratorConfig.SecuritySchemes[0].Type).To(Equal(definitions.OpenIDConnect))
		Expect(cfg.OpenAPIGeneratorConfig.SecuritySchemes[0].OpenIdConnectUrl).To(Equal("https://oidc.com"))
	})

	It("Prompts an error when providing invalid overwrite confirmation and rejecting", func() {
		// First run to create the config file.
		createStdout, createStderr, createErr := runInitWithInput(getLineForDefaultConfig())
		Expect(createErr).To(BeNil())
		Expect(createStderr).To(BeEmpty())
		Expect(createStdout).To(ContainSubstring("Successfully created gleece.config.json"))

		// Second run: invalid confirmation first, then reject with 'n'.
		rejectInput := getInput(Questionnaire{
			OverwriteConfirmation:  []string{"maybe", "n"},
			GenerateAuthMiddleware: []string{"n"},
		})

		stdout, stderr, err := runInitWithInput(rejectInput)
		Expect(err).To(BeNil())
		Expect(stderr).To(BeEmpty())
		Expect(stdout).To(ContainSubstring("File 'gleece.config.json' already exists. Overwrite? (y/n): "))
		Expect(stdout).To(ContainSubstring("Please confirm or reject by typing 'y' or 'n' and pressing enter"))
		Expect(stdout).To(ContainSubstring("Configuration overwrite aborted"))
	})

	It("Successfully overwrites configuration when accepting overwrite", func() {
		// First run to create the config file.
		createStdout, createStderr, createErr := runInitWithInput(getLineForDefaultConfig())
		Expect(createErr).To(BeNil())
		Expect(createStderr).To(BeEmpty())
		Expect(createStdout).To(ContainSubstring("Successfully created gleece.config.json"))

		// Second run: accept with 'y'.
		acceptInput := getInput(Questionnaire{
			OverwriteConfirmation:  []string{"y"},
			GenerateAuthMiddleware: []string{"n"},
		})

		stdout, stderr, err := runInitWithInput(acceptInput)
		Expect(err).To(BeNil())
		Expect(stderr).To(BeEmpty())
		Expect(stdout).To(ContainSubstring("File 'gleece.config.json' already exists. Overwrite? (y/n): "))
		Expect(stdout).To(ContainSubstring("Successfully created gleece.config.json"))
	})

	It("Generates authentication middleware skeleton code", func() {
		input := getInput(Questionnaire{
			AuthMiddlewarePackageName: []string{"auth/pkg"},
			GenerateAuthMiddleware:    []string{"y"},
		})

		_, _, err := runInitWithInput(input)
		Expect(err).To(BeNil())

		// Verify authentication.go exists in the expected directory
		// The code suggests:
		// splitPkgPath := strings.Split(config.RoutesConfig.AuthorizationConfig.AuthFileFullPackageName, "/")
		// concatenatedPath := append(splitPkgPath[1:], "authentication.go")
		// suggestedOutputPath := filepath.Join(concatenatedPath...)
		// If input is "auth/pkg", split is ["auth", "pkg"].
		// splitPkgPath[1:] is ["pkg"].
		// Output is "pkg/authentication.go".

		expectedPath := "pkg/authentication.go"
		_, err = os.Stat(expectedPath)
		Expect(err).To(BeNil(), "Authentication middleware file should exist at %s", expectedPath)

		content, err := os.ReadFile(expectedPath)
		Expect(err).To(BeNil())
		Expect(string(content)).To(ContainSubstring("package pkg"))
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

type SecuritySchemeInput []string

type Questionnaire struct {
	ControllerGlobs           []string
	AllowPackageLoadFailures  []string
	RoutingEngine             []string
	RoutesPackageName         []string
	RoutesOutputPath          []string
	OutputFilePerms           []string
	ValidateResponsePayload   []string
	SkipGenerateDateComment   []string
	AuthMiddlewarePackageName []string
	EnforceSecurity           []string
	OpenAPIVersion            []string
	APITitle                  []string
	APIDescription            []string
	APIVersion                []string
	TermsOfServiceURL         []string
	IncludeContact            []string
	IncludeLicense            []string
	BaseURL                   []string
	OpenAPISpecOutputPath     []string
	AddSecurityScheme         []string
	ValidateTopLevelOnlyEnum  []string
	GenerateEnumValidator     []string
	SecuritySchemes           []SecuritySchemeInput
	OverwriteConfirmation     []string
	GenerateAuthMiddleware    []string
	AuthMiddlewareOutputPath  []string
}

func getLineForDefaultConfig() string {
	return getInput(Questionnaire{
		OverwriteConfirmation: []string{"n"},
	})
}

func getInput(q Questionnaire) string {
	var inputs []string

	// Helper to append a slice of strings if it exists, otherwise a default value
	appendOrEmpty := func(field []string, defaultValue string) {
		if len(field) == 0 {
			inputs = append(inputs, defaultValue)
		} else {
			inputs = append(inputs, field...)
		}
	}

	// This order must match the prompt sequence in init.go/cliWizard
	appendOrEmpty(q.ControllerGlobs, "")
	appendOrEmpty(q.AllowPackageLoadFailures, "")
	appendOrEmpty(q.RoutingEngine, "")
	appendOrEmpty(q.RoutesPackageName, "")
	appendOrEmpty(q.RoutesOutputPath, "")
	appendOrEmpty(q.OutputFilePerms, "")
	appendOrEmpty(q.ValidateResponsePayload, "")
	appendOrEmpty(q.SkipGenerateDateComment, "")
	appendOrEmpty(q.AuthMiddlewarePackageName, "")
	appendOrEmpty(q.EnforceSecurity, "")
	appendOrEmpty(q.OpenAPIVersion, "")
	appendOrEmpty(q.APITitle, "")
	appendOrEmpty(q.APIDescription, "")
	appendOrEmpty(q.APIVersion, "")
	appendOrEmpty(q.TermsOfServiceURL, "")
	appendOrEmpty(q.IncludeContact, "")
	appendOrEmpty(q.IncludeLicense, "")
	appendOrEmpty(q.BaseURL, "")
	appendOrEmpty(q.OpenAPISpecOutputPath, "")
	appendOrEmpty(q.AddSecurityScheme, "")

	// Only append SecuritySchemes if AddSecurityScheme effectively resulted in 'y'
	shouldAddSecurity := false
	if len(q.AddSecurityScheme) > 0 {
		last := q.AddSecurityScheme[len(q.AddSecurityScheme)-1]
		if last != "" {
			shouldAddSecurity = last == "y" || last == "yes"
		}
	}
	if shouldAddSecurity {
		inputs = append(inputs, getSecuritySchemesInput(q.SecuritySchemes)...)
	}

	appendOrEmpty(q.ValidateTopLevelOnlyEnum, "")
	appendOrEmpty(q.GenerateEnumValidator, "")

	if len(q.OverwriteConfirmation) > 0 {
		inputs = append(inputs, q.OverwriteConfirmation...)
	}

	appendOrEmpty(q.GenerateAuthMiddleware, "")
	appendOrEmpty(q.AuthMiddlewareOutputPath, "")

	return strings.Join(inputs, "\n") + "\n"
}

func getSecuritySchemesInput(schemes []SecuritySchemeInput) []string {
	var inputs []string
	for i, scheme := range schemes {
		inputs = append(inputs, scheme...)
		if i < len(schemes)-1 {
			inputs = append(inputs, "y")
		} else {
			inputs = append(inputs, "n")
		}
	}
	return inputs
}
