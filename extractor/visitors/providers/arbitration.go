package providers

import (
	"go/ast"
	"go/token"

	"github.com/gopher-fleece/gleece/extractor/arbitrators"
)

type ArbitrationProvider struct {
	// The source files being included/processed by the visitor
	//
	// Used by AST-level functions
	sourceFiles map[string]*ast.File

	// The file set being included/processed by the visitor.
	//
	// Used by AST-level functions
	fileSet *token.FileSet

	packagesFacade *arbitrators.PackagesFacade

	// An arbitrator providing logic for working with Go's AST
	astArbitrator *arbitrators.AstArbitrator
}

func (p *ArbitrationProvider) Pkg() *arbitrators.PackagesFacade {
	return p.packagesFacade
}

func (p *ArbitrationProvider) Ast() *arbitrators.AstArbitrator {
	return p.astArbitrator
}

func (p *ArbitrationProvider) FileSet() *token.FileSet {
	return p.fileSet
}

func NewArbitrationProvider(
	sourceFiles map[string]*ast.File,
	fileSet *token.FileSet,
) *ArbitrationProvider {
	packagesFacade := arbitrators.NewPackagesFacade()
	astArbitrator := arbitrators.NewAstArbitrator(&packagesFacade, fileSet)

	return &ArbitrationProvider{
		sourceFiles:    sourceFiles,
		fileSet:        fileSet,
		packagesFacade: &packagesFacade,
		astArbitrator:  &astArbitrator,
	}
}

func (v *ArbitrationProvider) GetAllSourceFiles() []*ast.File {
	result := []*ast.File{}
	for _, file := range v.sourceFiles {
		result = append(result, file)
	}
	return result
}
