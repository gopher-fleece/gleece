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

func newCliWizard(reader *bufio.Reader) *cliWizard {
	return &cliWizard{
		reader:   reader,
		validate: validator.New(),
	}
}

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

func (w *cliWizard) askSelection(question string, options []string, defaultValue string) string {
	fmt.Printf("%s (%s), Default: [%s]: ", question, strings.Join(options, "|"), defaultValue)

	for {
		input, _ := w.reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			return defaultValue
		}

		for _, opt := range options {
			if input == opt {
				return opt
			}
		}
		fmt.Printf("Invalid selection. Please choose from (%s): ", strings.Join(options, "|"))
	}
}

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

func (w *cliWizard) askSecuritySchemes() []definitions.SecuritySchemeConfig {
	var schemes []definitions.SecuritySchemeConfig
	for {
		scheme := definitions.SecuritySchemeConfig{}
		scheme.SecurityName = w.askOpenEnded("Scheme name (e.g. mySchema)", "")
		if scheme.SecurityName == "" {
			break
		}
		scheme.Description = w.askOpenEnded("Description", "")
		scheme.Type = definitions.SecuritySchemeType(w.askSelection("Type", []string{"apiKey", "http", "oauth2", "openIdConnect"}, "http"))

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

func (w *cliWizard) askOAuthFlow() *definitions.OAuthFlow {
	flow := &definitions.OAuthFlow{}
	flow.AuthorizationURL = w.askOpenEnded("Authorization URL", "")
	flow.TokenURL = w.askOpenEnded("Token URL", "")
	flow.RefreshURL = w.askOpenEnded("Refresh URL", "")

	scopes := w.askVarArgs("Scopes (key:value pairs separated by commas)", []string{})
	if len(scopes) > 0 {
		flow.Scopes = make(map[string]string)
		for _, s := range scopes {
			parts := strings.Split(s, ":")
			if len(parts) == 2 {
				flow.Scopes[parts[0]] = parts[1]
			}
		}
	}

	return flow
}

func (w *cliWizard) askApiKeySchema(scheme *definitions.SecuritySchemeConfig) {
	scheme.In = definitions.SecuritySchemeIn(w.askSelection("In", []string{"query", "header", "cookie"}, "header"))
	scheme.FieldName = w.askOpenEnded("Field name", "X-API-KEY")
}

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
		"bearer"),
	)
}

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

func (w *cliWizard) askOpenIDConnectSchema(scheme *definitions.SecuritySchemeConfig) {
	scheme.OpenIdConnectUrl = w.askOpenEnded("OpenID Connect URL", "")
}

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

func (w *cliWizard) askEmail(question string, defaultValue string) string {
	for {
		input := w.askOpenEnded(question, defaultValue)
		if input == "" || w.validate.Var(input, "email") == nil {
			return input
		}
		fmt.Println("Invalid email format.")
	}
}

func (w *cliWizard) askUrl(question string, defaultValue string) string {
	for {
		input := w.askOpenEnded(question, defaultValue)
		if input == "" || w.validate.Var(input, "url") == nil {
			return input
		}
		fmt.Println("Invalid URL format.")
	}
}

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
			if fileExt != "" && fileExt != normalizedReqExt {
				input = strings.TrimSuffix(input, fileExt)
			}
			input = input + normalizedReqExt
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

func (w *cliWizard) askBlockingConfirm(question string) bool {
	fmt.Printf("%s (y/n): ", question)

	invalidInputMsg := "Please confirm or reject by typing 'y' or 'n' and pressing enter"
	for {
		input, _ := w.reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))

		if input == "" {
			fmt.Println(invalidInputMsg)
			continue
		}

		switch strings.ToLower(input) {
		case "y":

			return true
		case "n":
			return false
		default:
			fmt.Println(invalidInputMsg)
		}
	}
}
