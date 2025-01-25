package controller

import (
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/bmatcuk/doublestar/v4"
	MapSet "github.com/deckarep/golang-set/v2"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"golang.org/x/tools/go/packages"
)

type ControllerVisitor struct {
	// Data
	config      *definitions.GleeceConfig
	sourceFiles map[string]*ast.File
	fileSet     *token.FileSet
	packages    []*packages.Package

	// Context
	currentSourceFile *ast.File
	currentGenDecl    *ast.GenDecl
	currentController *definitions.ControllerMetadata
	importIdCounter   uint64

	// Diagnostics
	stackFrozen     bool
	diagnosticStack []string
	lastError       *error

	// Output
	controllers []definitions.ControllerMetadata
}

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

func (v *ControllerVisitor) init(sourceFileGlobs []string) error {
	v.sourceFiles = make(map[string]*ast.File)
	v.fileSet = token.NewFileSet()

	packages := MapSet.NewSet[string]()

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

	info, err := extractor.GetPackagesFromExpressions(packages.ToSlice())
	if err != nil {
		return err
	}
	v.packages = info

	return nil
}
