package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/gopher-fleece/gleece/v2/definitions"
)

type wizard struct {
	reader *bufio.Reader
}

func (w *wizard) askOpenEnded(question string, defaultValue string) string {
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

func (w *wizard) askBool(question string, defaultValue bool) bool {
	fmt.Printf("%s (%s), Default: [%t]", question, "y/n", defaultValue)

	input, _ := w.reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	if input == "" {
		return defaultValue
	}
	return input == "y" || input == "yes"
}

func (w *wizard) askSelection(question string, options []string, defaultValue string) string {
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

func (w *wizard) askVarArgs(question string, defaultValues []string) []string {
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

func (w *wizard) askSecuritySchemes() []definitions.SecuritySchemeConfig {
	var schemes []definitions.SecuritySchemeConfig
	for {
		scheme := definitions.SecuritySchemeConfig{}
		scheme.SecurityName = w.askOpenEnded("Scheme name (e.g. someSchemaName)", "")
		if scheme.SecurityName == "" {
			break
		}
		scheme.Description = w.askOpenEnded("Description", "")
		scheme.Type = definitions.SecuritySchemeType(w.askSelection("Type", []string{"apiKey", "http", "oauth2", "openIdConnect"}, "http"))

		if scheme.Type == definitions.APIKey {
			scheme.In = definitions.SecuritySchemeIn(w.askSelection("In", []string{"query", "header", "cookie"}, "header"))
			scheme.FieldName = w.askOpenEnded("Field name", "X-API-KEY")
		} else if scheme.Type == definitions.HTTP {
			scheme.Scheme = definitions.HttpAuthScheme(w.askOpenEnded("Scheme (e.g. bearer, basic)", "bearer"))
		}

		schemes = append(schemes, scheme)

		if !w.askBool("Add another security scheme?", false) {
			break
		}
	}
	return schemes
}
