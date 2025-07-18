package providers

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/core/arbitrators"
)

type ArbitrationProvider struct {
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

func NewArbitrationProvider(globs []string) (*ArbitrationProvider, error) {
	packagesFacade, err := arbitrators.NewPackagesFacade(globs)
	if err != nil {
		return nil, err
	}

	astArbitrator := arbitrators.NewAstArbitrator(&packagesFacade)

	return &ArbitrationProvider{
		packagesFacade: &packagesFacade,
		astArbitrator:  &astArbitrator,
	}, nil
}

func (v *ArbitrationProvider) GetAllSourceFiles() []*ast.File {
	return v.Pkg().GetAllSourceFiles()
}
