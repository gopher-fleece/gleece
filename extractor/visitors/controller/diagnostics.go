package controller

import (
	"fmt"
	"runtime"

	"github.com/gopher-fleece/gleece/infrastructure/logger"
)

// enter records a stackframe and message in the diagnostic stack that can be used for
// debugging and/or printing failures in a clear way.
//
// Note that this method does nothing if the diagnostic stack is frozen
func (v *ControllerVisitor) enter(message string) {
	if v.stackFrozen {
		return
	}

	var formattedMessage string
	if len(message) > 0 {
		formattedMessage = fmt.Sprintf("- (%s)", message)
	}

	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		v.diagnosticStack = append(v.diagnosticStack, fmt.Sprintf("<unknown>.<unknown> - %s", formattedMessage))
		logger.Warn("Could not determine caller for diagnostic message %s", formattedMessage)
		return
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		v.diagnosticStack = append(v.diagnosticStack, fmt.Sprintf("%s:%d - %s", file, line, formattedMessage))
		logger.Warn("Could not determine caller function for diagnostic message %s", formattedMessage)
		return
	}

	v.diagnosticStack = append(v.diagnosticStack, fmt.Sprintf("%s\n\t\t%s:%d%s", fn.Name(), file, line, formattedMessage))
}

// exit pops the diagnostic stack to indicate the function returned successfully.
//
// Note that this method does nothing if the diagnostic stack is frozen
func (v *ControllerVisitor) exit() {
	if !v.stackFrozen && len(v.diagnosticStack) > 0 {
		v.diagnosticStack = v.diagnosticStack[:len(v.diagnosticStack)-1]
	}
}

// getFrozenError is an idempotent function that creates an error from the given format string and arguments
// and freezes the visitor's diagnostic stack.
func (v *ControllerVisitor) getFrozenError(format string, args ...any) error {
	v.stackFrozen = true
	return fmt.Errorf(format, args...)
}

// frozenError is an idempotent function that returns the provided error unmodified
// and freezes the visitor's diagnostic stack.
func (v *ControllerVisitor) frozenError(err error) error {
	// Just a convenient way to freeze the diagnostic stack while returning the same error
	v.stackFrozen = true
	return err
}
