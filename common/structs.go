package common

import (
	"fmt"
	"strings"
)

type Collector[T comparable] struct {
	items []T
}

func (c *Collector[T]) AddIfNotZero(item T) {
	var zero T
	if item != zero {
		c.items = append(c.items, item)
	}
}

func (c *Collector[T]) Add(item T) {
	c.items = append(c.items, item)
}

func (c *Collector[T]) AddMany(items []T) {
	for _, item := range items {
		c.Add(item)
	}
}

func (c *Collector[T]) Items() []T {
	return c.items
}

func (c *Collector[T]) HasAny() bool {
	return len(c.items) > 0
}

type ContextualError struct {
	Context string
	Errors  []error
}

func (e *ContextualError) Error() string {
	if len(e.Errors) == 0 {
		return fmt.Sprintf("[%s] no errors", e.Context)
	}
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("[%s] encountered %d error(s):", e.Context, len(e.Errors)))
	for _, err := range e.Errors {
		if err == nil {
			continue
		}

		lines := strings.Split(err.Error(), "\n")
		sb.WriteString(fmt.Sprintf("\n\t- %s", lines[0])) // First line with dash

		// Indent any remaining lines
		for _, line := range lines[1:] {
			sb.WriteString(fmt.Sprintf("\n\t  %s", line))
		}
	}
	return sb.String()
}

func (e *ContextualError) Unwrap() []error {
	return e.Errors
}
