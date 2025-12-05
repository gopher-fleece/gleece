package symboldg

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/common/linq"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/graphs"
)

type ModelNameTransformer = func(modelBaseName string, typeParamNames []string) string

type InitializationTarget struct {
	TypeParams []*SymbolNode
}

type GenericInstantiationList struct {
	Struct  metadata.StructMeta
	Targets []InitializationTarget
}

type MapComposite struct {
	Composite  *SymbolNode
	TypeParams [2]*SymbolNode
}

type RawStructModelsList struct {
	GenericStructs []GenericInstantiationList
	Structs        []metadata.StructMeta
}

func ComposeModels(
	reductionCtx metadata.ReductionContext,
	graph SymbolGraphBuilder,
	modelNameTransformer ModelNameTransformer,
) ([]definitions.StructMetadata, error) {
	// Collect struct and struct-like models (Structs, Aliases, generic struct instantiations)
	modelsList, err := collectStructModelList(graph)
	if err != nil {
		return nil, fmt.Errorf("failed to collect models from graph - %v", err)
	}

	// Reduce said struct/struct-like HIR items to the flattened, emitter-ready structs (definitions)
	allModels, err := reduceStructLists(reductionCtx, modelsList, modelNameTransformer)
	if err != nil {
		return nil, fmt.Errorf("failed to reduce struct models list - %v", err)
	}

	// Collect/Reduce 'Alias' type models (e.g. `type Something = string`)
	aliasModels, err := collectAliasModels(reductionCtx, graph)
	if err != nil {
		return allModels, fmt.Errorf("failed to collect alias models - %v", err)
	}

	// Concat all struct-like entities
	allModels = slices.Concat(allModels, aliasModels)

	return allModels, nil
}

func collectStructModelList(graph SymbolGraphBuilder) (RawStructModelsList, error) {
	modelList := RawStructModelsList{
		GenericStructs: []GenericInstantiationList{},
		Structs:        []metadata.StructMeta{},
	}

	for _, structNode := range graph.FindByKind(common.SymKindStruct) {
		structMeta, isStructMeta := structNode.Data.(metadata.StructMeta)
		if !isStructMeta {
			return RawStructModelsList{}, fmt.Errorf(
				"expected node with ID '%s' to be a StructMeta node but got '%v'",
				structNode.Id.Name,
				structNode.Kind,
			)
		}

		instEdges := graph.GetEdges(
			structNode.Id,
			[]SymbolEdgeKind{EdgeKindInstantiates},
		)

		if len(instEdges) <= 0 {
			// Normal struct - append as-is
			modelList.Structs = append(modelList.Structs, structMeta)
			continue
		}

		// Generic struct - need to return create an instantiation list
		instantiations := GenericInstantiationList{
			Struct:  structMeta,
			Targets: []InitializationTarget{},
		}

		for _, instEdge := range instEdges {
			governingComposite := graph.Get(instEdge.Edge.From)
			if governingComposite == nil {
				return modelList, fmt.Errorf(
					"failed to determine governing composite for generic struct '%s'",
					structMeta.Name,
				)
			}

			compositeMeta, isCompositeMeta := governingComposite.Data.(*metadata.CompositeMeta)
			if !isCompositeMeta {
				return modelList, fmt.Errorf(
					"expected node '%s' to be a composite meta but got '%v'",
					governingComposite.Id.Name,
					governingComposite.Kind,
				)
			}

			typeParams := linq.Map(compositeMeta.Operands, func(key graphs.SymbolKey) *SymbolNode {
				return graph.Get(key)
			})

			instantiations.Targets = append(
				instantiations.Targets,
				InitializationTarget{TypeParams: typeParams},
			)
		}

		modelList.GenericStructs = append(modelList.GenericStructs, instantiations)
	}

	return modelList, nil
}

func reduceStructLists(
	ctx metadata.ReductionContext,
	structList RawStructModelsList,
	modelNameTransformer ModelNameTransformer,
) ([]definitions.StructMetadata, error) {
	// Allocate enough for both standard structs and whatever materialized ephemeral/generic we need
	reducedList := make(
		[]definitions.StructMetadata,
		0,
		len(structList.Structs)+len(structList.GenericStructs),
	)

	// Reduce the normal structs
	for _, stdStruct := range structList.Structs {
		reduced, err := stdStruct.Reduce(ctx)
		if err != nil {
			return reducedList, fmt.Errorf("failed to reduce struct '%s' - %v", stdStruct.Name, err)
		}
		reducedList = append(reducedList, reduced)
	}

	// Synthesize ephemeral models (i.e.,  instantiations like SomeStruct[string, int])
	for _, instComposite := range structList.GenericStructs {
		instantiations, err := synthesizeModelsForGenericStruct(
			ctx,
			instComposite,
			modelNameTransformer,
		)

		if err != nil {
			return reducedList, fmt.Errorf(
				"failed to synthesize instantiations for struct '%s' - %v",
				instComposite.Struct.Name,
				err,
			)
		}

		reducedList = append(reducedList, instantiations...)
	}

	return reducedList, nil
}

func instantiateGenericModel(
	rawStruct *metadata.StructMeta,
	reducedStruct definitions.StructMetadata,
	typeParamReplacementNodes []*SymbolNode,
	modelNameTransformer func(modelBaseName string, typeParamNames []string) string,
) (definitions.StructMetadata, error) {

	// Clone the reduced struct - Fields is a slice and therefore the struct as a whole
	// is not safe to modify.
	clonedStruct := reducedStruct.Clone()

	rawParamNames := linq.Map(typeParamReplacementNodes, func(tParamNode *SymbolNode) string {
		if tParamNode.Kind.IsBuiltin() {
			return tParamNode.Id.Name
		}
		return tParamNode.Data.(*metadata.TypeParamMeta).Name
	})

	if modelNameTransformer != nil {
		clonedStruct.Name = modelNameTransformer(clonedStruct.Name, rawParamNames)
	} else {
		clonedStruct.Name = StandardModelNameTransformer(clonedStruct.Name, rawParamNames)
	}

	for fieldIdx, field := range rawStruct.Fields {
		// Check if this is a generic field
		if field.Type.Root != nil && field.Type.Root.Kind() == metadata.TypeRefKindParam {
			nameParts := strings.SplitN(field.Type.Name, "#", 2)
			if len(nameParts) != 2 {
				return clonedStruct, fmt.Errorf(
					"failed to determine replacement to generic placeholder '%s' in field '%s'",
					field.Type.Name,
					field.Name,
				)
			}

			replParamIdx, err := strconv.ParseInt(nameParts[1], 10, 16)
			if err != nil {
				return clonedStruct, fmt.Errorf(
					"failed to parse replacement index to generic placeholder '%s' in field '%s'",
					field.Type.Name,
					field.Name,
				)
			}

			// Re-write the type
			clonedStruct.Fields[fieldIdx].Type = rawParamNames[replParamIdx]
		}

	}

	return clonedStruct, nil
}

func materializeInstantiationTarget(
	rawStruct *metadata.StructMeta,
	reducedStruct definitions.StructMetadata,
	target InitializationTarget,
	modelNameTransformer func(modelBaseName string, typeParamNames []string) string,
) (definitions.StructMetadata, error) {
	instance, err := instantiateGenericModel(rawStruct, reducedStruct, target.TypeParams, modelNameTransformer)
	if err != nil {
		return instance, fmt.Errorf("failed to construct fields for generic struct '%s' - %v", rawStruct.Name, err)
	}

	return instance, nil
}

func synthesizeModelsForGenericStruct(
	reductionCtx metadata.ReductionContext,
	genInstantiation GenericInstantiationList,
	modelNameTransformer func(modelBaseName string, typeParamNames []string) string,
) ([]definitions.StructMetadata, error) {
	// Reduce once and re-use
	reduced, err := genInstantiation.Struct.Reduce(reductionCtx)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to reduce struct '%s' for materialization - %v",
			genInstantiation.Struct.Name,
			err,
		)
	}

	materialized := []definitions.StructMetadata{}
	for _, target := range genInstantiation.Targets {
		instantiated, err := materializeInstantiationTarget(
			&genInstantiation.Struct,
			reduced,
			target,
			modelNameTransformer,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to materialize an instantiation of struct '%s' - %v",
				genInstantiation.Struct.Name,
				err,
			)
		}
		materialized = append(materialized, instantiated)
	}

	return materialized, nil
}

func collectAliasModels(
	ctx metadata.ReductionContext,
	graph SymbolGraphBuilder,
) ([]definitions.StructMetadata, error) {
	reduced := []definitions.StructMetadata{}

	for _, aliasNode := range graph.FindByKind(common.SymKindAlias) {
		aliasMeta, isAliasMeta := aliasNode.Data.(metadata.AliasMeta)
		if !isAliasMeta {
			return reduced, fmt.Errorf(
				"expected node with ID '%s' to be an AliasMeta node but got '%v'",
				aliasNode.Id.Name,
				aliasNode.Kind,
			)
		}

		alias, err := aliasMeta.Reduce(ctx)
		if err != nil {
			return reduced, fmt.Errorf("failed to reduce AliasMeta '%s' - %v", aliasMeta.Name, err)
		}

		reduced = append(reduced, alias)
	}

	return reduced, nil
}

// StandardModelNameTransformer turns a model base name and type parameters
// into a PascalCase model name. Example:
//
//	StandardModelNameTransformer("user", []string{"id", "meta"}) == "UserIdMeta"
func StandardModelNameTransformer(modelBaseName string, typeParamNames []string) string {
	toPascal := func(s string) string {
		if s == "" {
			return ""
		}
		runes := []rune(s)
		runes[0] = unicode.ToUpper(runes[0])
		return string(runes)
	}

	var b strings.Builder
	b.WriteString(toPascal(modelBaseName))
	for _, tp := range typeParamNames {
		b.WriteString(toPascal(tp))
	}

	return b.String()
}
