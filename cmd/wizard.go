package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/gopher-fleece/gleece/v2/definitions"
)

type cliWizard struct {
	reader *bufio.Reader
}

func (w *cliWizard) askOpenEnded(question string, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s, Default: [%s]: ", question, defaultValue)
	} else {
		fmt.Printf("%s: ", question)
	}

	input, _ := w.reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}
	return input
}

func (w *cliWizard) askBool(question string, defaultValue bool) bool {
	fmt.Printf("%s (%s), Default: [%t]: ", question, "y/n", defaultValue)

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
