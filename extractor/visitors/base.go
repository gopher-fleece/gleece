package visitors

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"runtime"
	"slices"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	MapSet "github.com/deckarep/golang-set/v2"
	"github.com/gopher-fleece/gleece/extractor/visitors/providers"
	"github.com/gopher-fleece/gleece/gast"
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
	lastError *error
}

func (v *BaseVisitor) initialize(context *VisitContext) error {
	err := contextInitGuard(context)
	if err != nil {
		return err
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
		return err
	}

	sourceFiles := make(map[string]*ast.File)
	fileSet := token.NewFileSet()

	packages := MapSet.NewSet[string]()

	var globs []string
	if len(context.GleeceConfig.CommonConfig.ControllerGlobs) > 0 {
		globs = context.GleeceConfig.CommonConfig.ControllerGlobs
	} else {
		globs = []string{"./*.go", "./**/*.go"}
	}

	// For each glob expression (provided via gleece.config), parse all matching files
	for _, globExpr := range globs {
		globbedSources, err := doublestar.FilepathGlob(globExpr)
		if err != nil {
			return err
		}

		for _, sourceFile := range globbedSources {
			file, err := parser.ParseFile(fileSet, sourceFile, nil, parser.ParseComments)
			if err != nil {
				logger.Error("Error parsing file %s - %v", sourceFile, err)
				return err
			}

			sourceFiles[sourceFile] = file

			packageName, err := gast.GetFullPackageName(file, fileSet)
			if err != nil {
				return err
			}
			packages.Add(packageName)
		}
	}

	v.context = context
	v.context.ArbitrationProvider = providers.NewArbitrationProvider(sourceFiles, fileSet)

	if v.context.SyncedProvider == nil {
		v.context.SyncedProvider = &providers.SyncedProvider{}
	}

	return nil
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

func (v BaseVisitor) GetFormattedDiagnosticStack() string {
	stack := slices.Clone(v.diagnosticStack)
	slices.Reverse(stack)
	return strings.Join(stack, "\n\t")
}

func (v *BaseVisitor) GetLastError() *error {
	return v.lastError
}

func (v *BaseVisitor) GetAllSourceFiles() []*ast.File {
	return v.context.ArbitrationProvider.GetAllSourceFiles()
}

func contextInitGuard(context *VisitContext) error {
	if context == nil {
		return fmt.Errorf("nil context was given to initializeWithGlobs")
	}

	return nil
}
