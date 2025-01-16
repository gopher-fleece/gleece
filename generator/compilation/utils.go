package compilation

import (
	"fmt"
	"go/format"

	"golang.org/x/tools/imports"
)

func OptimizeImportsAndFormat(sourceCode string) (string, error) {
	// Use imports.Process to optimize imports and format the code
	optSource, err := imports.Process("", []byte(sourceCode), nil)
	if err != nil {
		return "", fmt.Errorf("failed to optimize imports: %w", err)
	}

	// Ensure the code is properly formatted
	formattedSource, err := format.Source(optSource)
	if err != nil {
		return "", fmt.Errorf("failed to format source code: %w", err)
	}

	return string(formattedSource), nil
}
