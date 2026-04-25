package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gopher-fleece/gleece/v2/definitions"
)

type cliWizard struct {
	validate *validator.Validate
	reader   *bufio.Reader
}

// newCliWizard creates a new cliWizard instance with a validator and reader.
func newCliWizard(reader *bufio.Reader) *cliWizard {
	return &cliWizard{
		reader:   reader,
		validate: validator.New(),
	}
}

// askOpenEnded prompts the user for free-form input with a default value.
func (w *cliWizard) askOpenEnded(question string, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s, Default: [%s]: ", question, defaultValue)
	} else {
		fmt.Printf("%s, Default: [\"\"]: ", question)
	}

	input, _ := w.reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}
	return input
}

// askBool prompts the user for a yes/no boolean input.
func (w *cliWizard) askBool(question string, defaultValue bool) bool {
	var defaultValueFmt string
	if defaultValue {
		defaultValueFmt = "y"
	} else {
		defaultValueFmt = "n"
	}

	fmt.Printf("%s (y/n), Default: [%s]: ", question, defaultValueFmt)

	input, _ := w.reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	if input == "" {
		return defaultValue
	}
	return input == "y" || input == "yes"
}

// askSelection prompts the user to select one option from a list.
func (w *cliWizard) askSelection(
	question string,
	options []string,
	defaultValue string,
	caseSensitive bool,
) string {
	fmt.Printf("%s (%s), Default: [%s]: ", question, strings.Join(options, "|"), defaultValue)

	return w.getValidatedInput(
		&defaultValue,
		func(input string) bool {
			for _, opt := range options {
				if !caseSensitive {
					input = strings.ToLower(input)
					opt = strings.ToLower(opt)
				}
				if input == opt {
					return true
				}
			}

			return false
		},
		fmt.Sprintf("Invalid selection. Please choose from (%s): ", strings.Join(options, "|")),
	)
}

// askVarArgs prompts the user for a comma-separated list of values.
func (w *cliWizard) askVarArgs(question string, defaultValues []string) []string {
	fmt.Printf("%s (comma separated), Default: [%s]: ", question, strings.Join(defaultValues, ","))

	input, _ := w.reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValues
	}

	parts := strings.Split(input, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// askSecuritySchemes prompts the user to configure OpenAPI security schemes.
func (w *cliWizard) askSecuritySchemes() []definitions.SecuritySchemeConfig {
	var schemes []definitions.SecuritySchemeConfig
	for {
		scheme := definitions.SecuritySchemeConfig{}
		scheme.SecurityName = w.askOpenEnded("Scheme name (e.g. mySchema)", "")
		if scheme.SecurityName == "" {
			break
		}
		scheme.Description = w.askOpenEnded("Description", "")
		scheme.Type = definitions.SecuritySchemeType(w.askSelection(
			"Type",
			[]string{"apiKey", "http", "oauth2", "openIdConnect"},
			"http",
			true,
		))

		switch scheme.Type {
		case definitions.APIKey:
			w.askApiKeySchema(&scheme)
		case definitions.HTTP:
			w.askHttpSchema(&scheme)
		case definitions.OAuth2:
			w.askOAuth2Schema(&scheme)
		case definitions.OpenIDConnect:
			w.askOpenIDConnectSchema(&scheme)
		}

		schemes = append(schemes, scheme)

		if !w.askBool("Add another security scheme?", false) {
			break
		}
	}
	return schemes
}

// askOAuthFlow prompts the user to configure OAuth2 flow.
func (w *cliWizard) askOAuthFlow() *definitions.OAuthFlow {
	flow := &definitions.OAuthFlow{}
	flow.AuthorizationURL = w.askUrl("Authorization URL", "")
	flow.TokenURL = w.askUrl("Token URL", "")
	flow.RefreshURL = w.askUrl("Refresh URL", "")

	for {
		scopes := w.askVarArgs("Scopes (key:value pairs separated by commas)", []string{})
		flow.Scopes = make(map[string]string)

		if len(scopes) == 0 {
			// An empty 'scopes' object is valid under OAPI 3/3.1
			return flow
		}

		valid := true

		for _, s := range scopes {
			key, value, ok := strings.Cut(s, ":")
			if !ok || strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
				fmt.Printf("Invalid scope %q. Expected format: key:value\n", s)
				valid = false
				break
			}

			flow.Scopes[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}

		if valid {
			return flow
		}
	}
}

// askApiKeySchema prompts the user for API key schema details.
func (w *cliWizard) askApiKeySchema(scheme *definitions.SecuritySchemeConfig) {
	scheme.In = definitions.SecuritySchemeIn(w.askSelection(
		"In",
		[]string{"query", "header", "cookie"},
		"header",
		true,
	),
	)
	scheme.FieldName = w.askOpenEnded("Field name", "X-API-KEY")
}

// askHttpSchema prompts the user for HTTP security scheme details.
func (w *cliWizard) askHttpSchema(scheme *definitions.SecuritySchemeConfig) {
	scheme.Scheme = definitions.HttpAuthScheme(w.askSelection(
		"Scheme",
		[]string{
			"basic",
			"bearer",
			"digest",
			"hoba",
			"mutual",
			"negotiate",
			"oauth",
			"scram-sha-1",
			"scram-sha-256",
			"vapid",
		},
		"bearer",
		true,
	),
	)
}

// askOAuth2Schema prompts the user for OAuth2 security scheme details.
func (w *cliWizard) askOAuth2Schema(scheme *definitions.SecuritySchemeConfig) {
	scheme.Flows = &definitions.OAuthFlows{}
	if w.askBool("Configure implicit flow?", false) {
		scheme.Flows.Implicit = w.askOAuthFlow()
	}
	if w.askBool("Configure password flow?", false) {
		scheme.Flows.Password = w.askOAuthFlow()
	}
	if w.askBool("Configure client credentials flow?", false) {
		scheme.Flows.ClientCredentials = w.askOAuthFlow()
	}
	if w.askBool("Configure authorization code flow?", false) {
		scheme.Flows.AuthorizationCode = w.askOAuthFlow()
	}
}

// askOpenIDConnectSchema prompts the user for OpenID Connect schema details.
func (w *cliWizard) askOpenIDConnectSchema(scheme *definitions.SecuritySchemeConfig) {
	scheme.OpenIdConnectUrl = w.askOpenEnded("OpenID Connect URL", "")
}

// askFilePerms prompts the user for file permission octal string input and validates it.
func (w *cliWizard) askFilePerms(question string, defaultValue string) string {
	for {
		input := w.askOpenEnded(question, defaultValue)
		fileMod, err := definitions.PermissionStringToFileMod(input)
		if err == nil {
			return fmt.Sprintf("%o", fileMod)
		}
		fmt.Println("Invalid file permissions format. Please use an octal number like 0644 or 0777.")
	}
}

// askEmail prompts the user for an email address and validates it.
func (w *cliWizard) askEmail(question string, defaultValue string) string {
	for {
		input := w.askOpenEnded(question, defaultValue)
		if input == "" || w.validate.Var(input, "email") == nil {
			return input
		}
		fmt.Println("Invalid email format.")
	}
}

// askUrl prompts the user for a URL and validates it.
func (w *cliWizard) askUrl(question string, defaultValue string) string {
	for {
		input := w.askOpenEnded(question, defaultValue)
		if input == "" || w.validate.Var(input, "url") == nil {
			return input
		}
		fmt.Println("Invalid URL format.")
	}
}

// askFilePath prompts the user for a file path, validates it, and optionally enforces an extension.
func (w *cliWizard) askFilePath(question string, defaultValue string, forceExtension string) string {
	for {
		input := w.askOpenEnded(question, defaultValue)
		if input == "" {
			return input
		}

		// Validate first
		if w.validate.Var(input, "filepath") != nil {
			fmt.Println("Invalid file path.")
			continue
		}

		// Normalize the file extension, if required
		if forceExtension != "" {
			normalizedReqExt := "." + strings.TrimPrefix(forceExtension, ".")

			fileExt := filepath.Ext(input)
			if fileExt == "" {
				input = input + normalizedReqExt
			} else if fileExt != normalizedReqExt {
				input = strings.TrimSuffix(input, fileExt)
				input = input + normalizedReqExt
			}
		}

		// If the path is relative, we can use it as is
		if !filepath.IsAbs(input) {
			return input
		}

		// Otherwise we need to convert it to a relative path.
		// This step is not strictly necessary but is more 'hygienic' than leaving in absolute paths in the configuration.
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Failed to get current working directory.")
			return ""
		}

		rel, err := filepath.Rel(cwd, input)
		if err != nil {
			fmt.Println(fmt.Sprintf("Failed to convert path '%s' to a relative path. Using absolute instead", input))
			return input
		}

		return rel
	}
}

// askBlockingConfirm prompts the user for confirmation.
func (w *cliWizard) askBlockingConfirm(question string) bool {
	fmt.Printf("%s (y/n): ", question)

	validated := w.getValidatedInput(
		nil,
		func(input string) bool {
			lowercase := strings.ToLower(input)
			return lowercase == "y" || lowercase == "n"
		},
		"Please confirm or reject by typing 'y' or 'n' and pressing enter",
	)

	return validated == "y"
}

// getValidatedInput reads user input until the validator function returns true.
func (w *cliWizard) getValidatedInput(
	defaultValue *string,
	validator func(input string) bool,
	invalidInputMsg string,
) string {
	for {
		input, _ := w.reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))

		if input == "" {
			if defaultValue != nil {
				return *defaultValue
			}

			fmt.Println(invalidInputMsg)
			continue
		}

		if validator(input) {
			return input
		}

		fmt.Println(invalidInputMsg)
	}
}
