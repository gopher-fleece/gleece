package pipeline

import (
	"errors"
	"fmt"
	"go/ast"
	"slices"
	"strings"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/arbitrators/caching"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/validators"
	"github.com/gopher-fleece/gleece/core/visitors"
	"github.com/gopher-fleece/gleece/core/visitors/providers"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
)

type GleeceFlattenedMetadata struct {
	Imports           map[string][]string
	Flat              []definitions.ControllerMetadata
	Models            definitions.Models
	PlainErrorPresent bool
}

type GleecePipeline struct {
	gleeceConfig        *definitions.GleeceConfig
	metadataCache       metadata.MetaCache
	arbitrationProvider providers.ArbitrationProvider
	syncedProvider      providers.SyncedProvider

	symGraph    symboldg.SymbolGraphBuilder
	rootVisitor *visitors.ControllerVisitor
}

func NewGleecePipeline(gleeceConfig *definitions.GleeceConfig) (GleecePipeline, error) {
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

	metaCache := caching.NewMetadataCache()
	symGraph := symboldg.NewSymbolGraph()

	visitor, err := visitors.NewControllerVisitor(&visitors.VisitContext{
		GleeceConfig:        gleeceConfig,
		ArbitrationProvider: arbProvider,
		MetadataCache:       metaCache,
		GraphBuilder:        &symGraph,
	})
	if err != nil {
		return GleecePipeline{}, err
	}

	return GleecePipeline{
		rootVisitor:         visitor,
		symGraph:            &symGraph,
		gleeceConfig:        gleeceConfig,
		metadataCache:       metaCache,
		arbitrationProvider: *arbProvider,
		syncedProvider:      providers.NewSyncedProvider(),
	}, nil
}

func (p *GleecePipeline) Graph() symboldg.SymbolGraphBuilder {
	return p.symGraph
}

func (p *GleecePipeline) Run() (GleeceFlattenedMetadata, error) {
	err := p.GenerateGraph()
	if err != nil {
		return GleeceFlattenedMetadata{}, err
	}

	intermediate, err := p.GenerateIntermediate()
	if err != nil {
		return GleeceFlattenedMetadata{}, err
	}

	err = p.ValidateIntermediate(&intermediate)

	return intermediate, err
}

func (p *GleecePipeline) GenerateGraph() error {
	for _, file := range p.rootVisitor.GetAllSourceFiles() {
		ast.Walk(p.rootVisitor, file)
	}

	lastErr := p.rootVisitor.GetLastError()
	if lastErr != nil {
		logger.Error("Visitor encountered at-least one error. Last error:\n%v\n\t%s", lastErr, p.rootVisitor.GetFormattedDiagnosticStack())
		return lastErr
	}

	return nil
}

func (p *GleecePipeline) GenerateIntermediate() (GleeceFlattenedMetadata, error) {
	controllers, err := p.reduceControllers(p.rootVisitor.GetControllers())
	if err != nil {
		logger.Error("Failed to reduce controller tree to flat form: %w", err)
		return GleeceFlattenedMetadata{}, err
	}

	return GleeceFlattenedMetadata{
		Imports:           p.getImports(controllers),
		Flat:              controllers,
		Models:            p.getModels(),
		PlainErrorPresent: p.symGraph.IsSpecialPresent(symboldg.SpecialTypeError),
	}, nil
}

func (p *GleecePipeline) getImports(controllers []definitions.ControllerMetadata) map[string][]string {

	imports := make(map[string][]string)

	for _, controller := range controllers {
		imports[controller.PkgPath] = append(imports[controller.PkgPath], controller.Name)
		for _, route := range controller.Routes {
			p.appendRouteImports(imports, route)
		}
	}

	return imports
}

func (p *GleecePipeline) appendRouteImports(imports map[string][]string, route definitions.RouteMetadata) {
	for _, param := range route.FuncParams {
		paramPkgPath := param.TypeMeta.PkgPath

		if paramPkgPath != "" {
			paramImportName := fmt.Sprintf(
				"Param%d%s",
				param.UniqueImportSerial,
				common.UnwrapArrayTypeString(param.Name),
			)
			imports[paramPkgPath] = append(imports[paramPkgPath], paramImportName)
		}
	}

	for _, retVal := range route.Responses {
		retValPkgPath := retVal.PkgPath

		if retValPkgPath != "" {
			retValImportName := fmt.Sprintf(
				"Response%d%s",
				retVal.UniqueImportSerial,
				common.UnwrapArrayTypeString(retVal.Name),
			)
			imports[retValPkgPath] = append(imports[retValPkgPath], retValImportName)
		}
	}
}

func (p *GleecePipeline) ValidateIntermediate(meta *GleeceFlattenedMetadata) error {
	controllerValidationErrors := common.Map(meta.Flat, func(controller definitions.ControllerMetadata) error {
		return validators.ValidateController(p.gleeceConfig, p.arbitrationProvider.Pkg(), controller)
	})

	return errors.Join(controllerValidationErrors...)
}

func (p *GleecePipeline) reduceControllers(controllers []metadata.ControllerMeta) ([]definitions.ControllerMetadata, error) {
	var reducedControllers []definitions.ControllerMetadata

	for _, controller := range controllers {
		reduced, err := controller.Reduce(p.gleeceConfig, p.metadataCache, &p.syncedProvider)
		if err != nil {
			return []definitions.ControllerMetadata{}, err
		}
		reducedControllers = append(reducedControllers, reduced)
	}

	return reducedControllers, nil
}

func (p *GleecePipeline) getModels() definitions.Models {
	structs := p.symGraph.Structs()
	reducedStructs := common.Map(structs, func(s metadata.StructMeta) definitions.StructMetadata {
		return s.Reduce()
	})

	enums := p.symGraph.Enums()
	reducedEnums := common.Map(enums, func(e metadata.EnumMeta) definitions.EnumMetadata {
		return e.Reduce()
	})

	slices.SortFunc(reducedStructs, func(a, b definitions.StructMetadata) int {
		return strings.Compare(a.Name, b.Name)
	})

	slices.SortFunc(reducedEnums, func(a, b definitions.EnumMetadata) int {
		return strings.Compare(a.Name, b.Name)
	})

	return definitions.Models{
		Structs: reducedStructs,
		Enums:   reducedEnums,
	}
}
