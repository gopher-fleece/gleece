package visitors

import (
	"fmt"
	"go/ast"
	"runtime"
	"slices"
	"strings"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/arbitrators/caching"
	"github.com/gopher-fleece/gleece/core/visitors/providers"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
)

type Visitor interface {
	GetFormattedDiagnosticStack() string
	GetLastError() *error
	GetFiles() []*ast.File
}

type BaseVisitor struct {
	context *VisitContext

	// Determines whether the diagnostic stack is currently frozen.
	stackFrozen bool

	// The diagnostic stack is effectively just a simplified stacktrace with some additional
	// context (such as controller/receiver names) being added by the functions themselves.
	//
	// When an error occurs, the stack is 'frozen' which allows returning a proper error from the visitor.
	// This is used instead of a panic/recover as we need might need the extra context to actually understand what went wrong and where
	diagnosticStack []string

	// The last error encountered by the visitor.
	//
	// When setting this value, we also freeze the stack so we don't loose context
	lastError error
}

func (v *BaseVisitor) initialize(context *VisitContext) error {
	err := contextInitGuard(context)
	if err != nil {
		return v.getFrozenError("failed to initialize visitor - %v", err)
	}

	// If an arbitration provider was given, we use that, otherwise, we enumerate the configuration globs and create
	// the rest of the context ourselves.
	if context.ArbitrationProvider != nil {
		return v.initializeWithArbitrationProvider(context)
	}

	return v.initializeWithGlobs(context)
}

// init Prepares the visitor by enumerating files and loading their packages
func (v *BaseVisitor) initializeWithArbitrationProvider(context *VisitContext) error {
	err := contextInitGuard(context)
	if err != nil {
		return err
	}

	v.context = context
	return nil
}

// init Prepares the visitor by enumerating files and loading their packages
func (v *BaseVisitor) initializeWithGlobs(context *VisitContext) error {
	err := contextInitGuard(context)
	if err != nil {
		// This can actually never happen, so long as this method is called strictly from within 'initialize'
		return err
	}

	var globs []string
	if len(context.GleeceConfig.CommonConfig.ControllerGlobs) > 0 {
		globs = context.GleeceConfig.CommonConfig.ControllerGlobs
	} else {
		globs = []string{"./*.go", "./**/*.go"}
	}

	v.context = context

	arbProvider, err := providers.NewArbitrationProvider(globs)
	if err != nil {
		return err
	}

	v.context.ArbitrationProvider = arbProvider

	if v.context.SyncedProvider == nil {
		v.context.SyncedProvider = common.Ptr(providers.NewSyncedProvider())
	}

	if v.context.MetadataCache == nil {
		v.context.MetadataCache = caching.NewMetadataCache()
	}

	return nil
}

// enter records a stack frame and message in the diagnostic stack that can be used for
// debugging and/or printing failures in a clear way.
//
// Note that this method does nothing if the diagnostic stack is frozen
func (v *BaseVisitor) enterFmt(format string, args ...any) {
	v.enter(fmt.Sprintf(format, args...))
}

// enter records a stack frame and message in the diagnostic stack that can be used for
// debugging and/or printing failures in a clear way.
//
// Note that this method does nothing if the diagnostic stack is frozen
func (v *BaseVisitor) enter(message string) {
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
func (v *BaseVisitor) exit() {
	if !v.stackFrozen && len(v.diagnosticStack) > 0 {
		v.diagnosticStack = v.diagnosticStack[:len(v.diagnosticStack)-1]
	}
}

// getFrozenError is an idempotent function that creates an error from the given format string and arguments
// and freezes the visitor's diagnostic stack.
func (v *BaseVisitor) getFrozenError(format string, args ...any) error {
	v.stackFrozen = true
	return fmt.Errorf(format, args...)
}

// frozenError is an idempotent function that returns the provided error unmodified
// and freezes the visitor's diagnostic stack.
func (v *BaseVisitor) frozenError(err error) error {
	// Just a convenient way to freeze the diagnostic stack while returning the same error
	v.stackFrozen = true
	return err
}

// frozenError is an idempotent function that returns the provided error unmodified
// and freezes the visitor's diagnostic stack if the error was not nil
func (v *BaseVisitor) frozenIfError(err error) error {
	if err != nil {
		// Just a convenient way to freeze the diagnostic stack while returning the same error
		v.stackFrozen = true
	}
	return err
}

func (v BaseVisitor) GetFormattedDiagnosticStack() string {
	stack := slices.Clone(v.diagnosticStack)
	slices.Reverse(stack)
	return strings.Join(stack, "\n\t")
}

func (v *BaseVisitor) GetLastError() error {
	return v.lastError
}

func (v *BaseVisitor) GetAllSourceFiles() []*ast.File {
	return v.context.ArbitrationProvider.GetAllSourceFiles()
}

func (v *BaseVisitor) setLastError(err error) {
	v.lastError = err
	v.stackFrozen = true
}

func contextInitGuard(context *VisitContext) error {
	if context == nil {
		return fmt.Errorf("nil context was given to contextInitGuard")
	}

	return nil
}
