package pipeline

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/core/arbitrators/caching"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/visitors"
	"github.com/gopher-fleece/gleece/core/visitors/providers"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
)

type GleeceGeneratedMetadata struct {
	Flat   []definitions.ControllerMetadata
	Models definitions.Models
}

type GleecePipeline struct {
	gleeceConfig        definitions.GleeceConfig
	metadataCache       metadata.MetaCache
	arbitrationProvider providers.ArbitrationProvider
	syncedProvider      providers.SyncedProvider
}

func NewGleecePipeline(gleeceConfig definitions.GleeceConfig) (GleecePipeline, error) {
	var globs []string
	if len(gleeceConfig.CommonConfig.ControllerGlobs) > 0 {
		globs = gleeceConfig.CommonConfig.ControllerGlobs
	} else {
		globs = []string{"./*.go", "./**/*.go"}
	}

	arbProvider, err := providers.NewArbitrationProvider(globs)
	if err != nil {
		return GleecePipeline{}, err
	}

	return GleecePipeline{
		gleeceConfig:        gleeceConfig,
		metadataCache:       caching.NewMetadataCache(),
		arbitrationProvider: *arbProvider,
		syncedProvider:      providers.NewSyncedProvider(),
	}, nil
}

func (p *GleecePipeline) GenerateMetadata(config *definitions.GleeceConfig) (GleeceGeneratedMetadata, error) {
	visitor, err := visitors.NewControllerVisitor(&visitors.VisitContext{GleeceConfig: config})
	if err != nil {
		return GleeceGeneratedMetadata{}, err
	}

	for _, file := range visitor.GetAllSourceFiles() {
		ast.Walk(visitor, file)
	}

	lastErr := visitor.GetLastError()
	if lastErr != nil {
		logger.Error("Visitor encountered at-least one error. Last error:\n%v\n\t%s", lastErr, visitor.GetFormattedDiagnosticStack())
		return GleeceGeneratedMetadata{}, lastErr
	}

	controllers, err := p.reduceControllers(visitor.GetControllers())
	if err != nil {
		logger.Error("Failed to reduce controller tree to flat form: %w", err)
		return GleeceGeneratedMetadata{}, err
	}

	flatModels, hasAnyErrorTypes, err := visitor.GetModelsFlat()
	if err != nil {
		logger.Error("Failed to get models metadata: %v", err)
		return GleeceGeneratedMetadata{}, err
	}


	metadata := GleeceGeneratedMetadata{
		Flat: controllers,
	}
}

func (p *GleecePipeline) reduceControllers(controllers []metadata.ControllerMeta) ([]definitions.ControllerMetadata, error) {
	var reducedControllers []definitions.ControllerMetadata

	for _, controller := range controllers {
		reduced, err := controller.Reduce(&p.gleeceConfig, p.metadataCache, &p.syncedProvider)
		if err != nil {
			return []definitions.ControllerMetadata{}, err
		}
		reducedControllers = append(reducedControllers, reduced)
	}

	return reducedControllers, nil
}
