package controller

import (
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/bmatcuk/doublestar/v4"
	MapSet "github.com/deckarep/golang-set/v2"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor"
	"github.com/gopher-fleece/gleece/extractor/arbitrators"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
)

// ControllerVisitor locates and walks over Gleece controllers to extract metadata and route information
// This construct serves as the main business logic for Gleece's context generation
type ControllerVisitor struct {
	// A facade for managing packages.Package
	packagesFacade arbitrators.PackagesFacade

	astArbitrator arbitrators.AstArbitrator

	// The project's configuration, as specified in the user's gleece.config.json
	config *definitions.GleeceConfig

	// The source files being included/processed by the visitor
	//
	// Used by AST-level functions
	sourceFiles map[string]*ast.File

	// The file set being included/processed by the visitor.
	//
	// Used by AST-level functions
	fileSet *token.FileSet

	// The file currently being worked on
	currentSourceFile *ast.File

	// The last-encountered GenDecl.
	//
	// Documentation may be placed on TypeDecl or their parent GenDecl so we track these,
	// in case we need to fetch the docs from the TypeDecl's parent.
	currentGenDecl *ast.GenDecl

	// The current Gleece Controller being processed
	currentController *definitions.ControllerMetadata

	// An import counter.
	//
	// This is a bit of a hack. Currently, we avoid merging imports due to complexity so
	// every type reference gets its own unique import ID which is used by the template to prevent import conflicts.
	// Example:
	// import Type12 "abc"
	importIdCounter uint64

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

	// A list of fully processed controller metadata, ready to be passed to the routes/spec generators
	controllers []definitions.ControllerMetadata
}

// NewControllerVisitor Instantiates a new Gleece Controller visitor
func NewControllerVisitor(config *definitions.GleeceConfig) (*ControllerVisitor, error) {
	visitor := ControllerVisitor{}
	visitor.config = config

	var globs []string
	if len(config.CommonConfig.ControllerGlobs) > 0 {
		globs = config.CommonConfig.ControllerGlobs
	} else {
		globs = []string{"./*.go", "./**/*.go"}
	}

	return &visitor, visitor.init(globs)
}

// init Prepares the visitor by enumerating files and loading their packages
func (v *ControllerVisitor) init(sourceFileGlobs []string) error {
	v.sourceFiles = make(map[string]*ast.File)
	v.fileSet = token.NewFileSet()

	packages := MapSet.NewSet[string]()

	// For each glob expression (provided via gleece.config), parse all matching files
	for _, globExpr := range sourceFileGlobs {
		sourceFiles, err := doublestar.FilepathGlob(globExpr)
		if err != nil {
			v.lastError = &err
			return err
		}

		for _, sourceFile := range sourceFiles {
			file, err := parser.ParseFile(v.fileSet, sourceFile, nil, parser.ParseComments)
			if err != nil {
				logger.Error("Error parsing file %s - %v", sourceFile, err)
				v.lastError = &err
				return err
			}
			v.sourceFiles[sourceFile] = file

			packageName, err := extractor.GetFullPackageName(file, v.fileSet)
			if err != nil {
				return err
			}
			packages.Add(packageName)
		}
	}

	v.packagesFacade = arbitrators.NewPackagesFacade()

	// Eagerly load all packages picked up by the globs
	err := v.packagesFacade.LoadPackages(packages.ToSlice())
	if err != nil {
		return err
	}

	v.astArbitrator = arbitrators.NewAstArbitrator(&v.packagesFacade, v.fileSet)

	return nil
}
