package providers

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/v2/core/arbitrators"
	"github.com/gopher-fleece/gleece/v2/definitions"
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

func NewArbitrationProviderFromGleeceConfig(gleeceConfig *definitions.GleeceConfig) (*ArbitrationProvider, error) {
	return NewArbitrationProvider(NewArbitrationProviderConfig(gleeceConfig))
}

func NewArbitrationProvider(config ArbitrationProviderConfig) (*ArbitrationProvider, error) {
	packagesFacade, err := arbitrators.NewPackagesFacade(config.PackageFacadeConfig)
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
