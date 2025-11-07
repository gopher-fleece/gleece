package pipeline

import (
	"fmt"
	"go/ast"
	"slices"
	"strings"

	MapSet "github.com/deckarep/golang-set/v2"
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/common/linq"
	"github.com/gopher-fleece/gleece/core/arbitrators/caching"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/validators"
	"github.com/gopher-fleece/gleece/core/validators/diagnostics"
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

	symGraph            symboldg.SymbolGraphBuilder
	visitorOrchestrator *visitors.VisitorOrchestrator
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
	syncedProvider := providers.NewSyncedProvider()

	visitCtx := &visitors.VisitContext{
		GleeceConfig:        gleeceConfig,
		ArbitrationProvider: arbProvider,
		MetadataCache:       metaCache,
		GraphBuilder:        &symGraph,
		SyncedProvider:      &syncedProvider,
	}

	visitorOrchestrator, err := visitors.NewVisitorOrchestrator(visitCtx)
	if err != nil {
		return GleecePipeline{}, fmt.Errorf("the GleecePipeline could not construct an instance of VisitorOrchestrator - %v", err)
	}

	return GleecePipeline{
		visitorOrchestrator: visitorOrchestrator,
		symGraph:            &symGraph,
		gleeceConfig:        gleeceConfig,
		metadataCache:       metaCache,
		arbitrationProvider: *arbProvider,
		syncedProvider:      syncedProvider,
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

	diags, err := p.Validate()
	if err != nil {
		return GleeceFlattenedMetadata{}, err
	}

	errDiagEntities := diagnostics.GetDiagnosticsWithSeverity(
		diags,
		[]diagnostics.DiagnosticSeverity{diagnostics.DiagnosticError},
	)

	// Check if validators returned any errors
	if len(errDiagEntities) > 0 {
		// If so, return a formatted list of diagnostics
		return GleeceFlattenedMetadata{}, diagnostics.DiagnosticsToError(errDiagEntities)
	}

	intermediate, err := p.GenerateIntermediate()
	if err != nil {
		return GleeceFlattenedMetadata{}, err
	}
	return intermediate, err
}

func (p *GleecePipeline) GenerateGraph() error {
	for _, file := range p.arbitrationProvider.GetAllSourceFiles() {
		ast.Walk(p.visitorOrchestrator, file)
	}

	lastErr := p.visitorOrchestrator.GetLastError()
	if lastErr != nil {
		logger.Error(
			"Visitor encountered at-least one error. Last error:\n%v\n\t%s",
			lastErr,
			p.visitorOrchestrator.GetFormattedDiagnosticStack(),
		)
		return lastErr
	}

	return nil
}

func (p *GleecePipeline) GenerateIntermediate() (GleeceFlattenedMetadata, error) {
	controllers, err := p.getReducedControllers()
	if err != nil {
		logger.Error("Pipeline failed to obtain the controller list - %w", err)
		return GleeceFlattenedMetadata{}, err
	}

	models, err := p.getModels()
	if err != nil {
		logger.Error("Pipeline failed to obtain models list - %w", err)
		return GleeceFlattenedMetadata{}, err
	}

	return GleeceFlattenedMetadata{
		Imports:           p.getImports(controllers),
		Flat:              controllers,
		Models:            models,
		PlainErrorPresent: p.symGraph.IsSpecialPresent(symboldg.SpecialTypeError),
	}, nil
}

func (p *GleecePipeline) getReducedControllers() ([]definitions.ControllerMetadata, error) {
	controllers, err := p.reduceControllers(p.getControllers())
	if err != nil {
		logger.Error("Failed to reduce controller tree to flat form: %w", err)
		return []definitions.ControllerMetadata{}, err
	}

	slices.SortFunc(controllers, func(a, b definitions.ControllerMetadata) int {
		return strings.Compare(a.Name, b.Name)
	})

	return controllers, nil
}

func (p *GleecePipeline) getImports(controllers []definitions.ControllerMetadata) map[string][]string {
	// Standard Set is actually thread-safe. Just saying.
	imports := make(map[string]MapSet.Set[string])

	for _, controller := range controllers {
		if imports[controller.PkgPath] == nil {
			imports[controller.PkgPath] = MapSet.NewSet[string]()
		}

		imports[controller.PkgPath].Add(controller.Name)
		for _, route := range controller.Routes {
			p.appendRouteImports(imports, route)
		}
	}

	plainImportsMap := make(map[string][]string, len(imports))

	for pkgPath, importSet := range imports {
		plainImportsMap[pkgPath] = importSet.ToSlice()
	}

	return plainImportsMap
}

func (p *GleecePipeline) appendRouteImports(imports map[string]MapSet.Set[string], route definitions.RouteMetadata) {
	for _, param := range route.FuncParams {
		paramPkgPath := param.TypeMeta.PkgPath
		if paramPkgPath == "" {
			continue
		}

		if imports[paramPkgPath] == nil {
			imports[paramPkgPath] = MapSet.NewSet[string]()
		}

		paramImportName := fmt.Sprintf(
			"Param%d%s",
			param.UniqueImportSerial,
			common.UnwrapArrayTypeString(param.Name),
		)
		imports[paramPkgPath].Add(paramImportName)
	}

	for _, retVal := range route.Responses {
		retValPkgPath := retVal.PkgPath
		if retValPkgPath == "" {
			continue
		}

		if imports[retValPkgPath] == nil {
			imports[retValPkgPath] = MapSet.NewSet[string]()
		}

		retValImportName := fmt.Sprintf(
			"Response%d%s",
			retVal.UniqueImportSerial,
			common.UnwrapArrayTypeString(retVal.Name),
		)
		imports[retValPkgPath].Add(retValImportName)
	}
}

// Validate validates the metadata created by the graph generation phase
func (p *GleecePipeline) Validate() ([]diagnostics.EntityDiagnostic, error) {
	allDiags := []diagnostics.EntityDiagnostic{}

	for _, ctrl := range p.getControllers() {
		validator := validators.NewControllerValidator(p.gleeceConfig, p.arbitrationProvider.Pkg(), &ctrl)
		ctrlDiag, err := validator.Validate()
		if err != nil {
			return allDiags, fmt.Errorf("failed to validate controller '%s' due to an error - %w", ctrl.Struct.Name, err)
		}

		if !ctrlDiag.Empty() {
			allDiags = append(allDiags, ctrlDiag)
		}
	}

	return allDiags, nil
}

func (p *GleecePipeline) getControllers() []metadata.ControllerMeta {
	controllerNodes := p.symGraph.FindByKind(common.SymKindController)
	return linq.Map(controllerNodes, func(node *symboldg.SymbolNode) metadata.ControllerMeta {
		return node.Data.(metadata.ControllerMeta)
	})
}

func (p *GleecePipeline) reduceControllers(controllers []metadata.ControllerMeta) ([]definitions.ControllerMetadata, error) {
	var reducedControllers []definitions.ControllerMetadata

	for _, controller := range controllers {
		reduced, err := controller.Reduce(p.getReductionContext())
		if err != nil {
			return []definitions.ControllerMetadata{}, err
		}
		reducedControllers = append(reducedControllers, reduced)
	}

	return reducedControllers, nil
}

func (p *GleecePipeline) getModels() (definitions.Models, error) {
	ctx := p.getReductionContext()

	reducedStructs, err := symboldg.ComposeModels(p.getReductionContext(), p.Graph(), nil)
	if err != nil {
		return definitions.Models{}, err
	}

	reducedEnums := []definitions.EnumMetadata{}
	for _, enumEntity := range p.symGraph.Enums() {
		reduced, err := enumEntity.Reduce(ctx)
		if err != nil {
			return definitions.Models{}, fmt.Errorf("failed during reduction of enum '%s' - %v", enumEntity.Name, err)
		}
		reducedEnums = append(reducedEnums, reduced)
	}

	slices.SortFunc(reducedStructs, func(a, b definitions.StructMetadata) int {
		return strings.Compare(a.Name, b.Name)
	})

	slices.SortFunc(reducedEnums, func(a, b definitions.EnumMetadata) int {
		return strings.Compare(a.Name, b.Name)
	})

	return definitions.Models{
		Structs: reducedStructs,
		Enums:   reducedEnums,
	}, nil
}

func (p *GleecePipeline) getReductionContext() metadata.ReductionContext {
	return metadata.ReductionContext{
		GleeceConfig:   p.gleeceConfig,
		MetaCache:      p.metadataCache,
		SyncedProvider: &p.syncedProvider,
	}
}
