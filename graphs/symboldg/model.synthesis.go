package symboldg

import (
	"fmt"
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

type RawModelsList struct {
	GenericStructs []GenericInstantiationList
	Structs        []metadata.StructMeta
}

func ComposeModels(
	reductionCtx metadata.ReductionContext,
	graph SymbolGraphBuilder,
	modelNameTransformer ModelNameTransformer,
) ([]definitions.StructMetadata, error) {
	models, err := collectModelList(graph)
	if err != nil {
		return nil, fmt.Errorf("failed to collect models from graph - %v", err)
	}

	allModels := make([]definitions.StructMetadata, 0, len(models.Structs)+len(models.GenericStructs))

	// Reduce the standard structs
	for _, stdStruct := range models.Structs {
		reduced, err := stdStruct.Reduce(reductionCtx)
		if err != nil {
			return allModels, fmt.Errorf("failed to reduce struct '%s' - %v", stdStruct.Name, err)
		}
		allModels = append(allModels, reduced)
	}

	// Synthesize ephemeral models
	for _, instComposite := range models.GenericStructs {
		instantiations, err := synthesizeModelsForGenericStruct(
			reductionCtx,
			instComposite,
			modelNameTransformer,
		)

		if err != nil {
			return allModels, fmt.Errorf(
				"failed to synthesize instantiations for struct '%s' - %v",
				instComposite.Struct.Name,
				err,
			)
		}

		allModels = append(allModels, instantiations...)
	}

	return allModels, nil
}

func collectModelList(graph SymbolGraphBuilder) (RawModelsList, error) {
	modelList := RawModelsList{
		GenericStructs: []GenericInstantiationList{},
		Structs:        []metadata.StructMeta{},
	}

	for _, structNode := range graph.FindByKind(common.SymKindStruct) {
		structMeta, isStructMeta := structNode.Data.(metadata.StructMeta)
		if !isStructMeta {
			return RawModelsList{}, fmt.Errorf(
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

			compositeMeta, isCompositeMeta := governingComposite.Data.(*CompositeMeta)
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

func inPlaceInstantiateGenericModel(
	rawStruct *metadata.StructMeta,
	reducedStruct *definitions.StructMetadata,
	typeParamReplacementNodes []*SymbolNode,
	modelNameTransformer func(modelBaseName string, typeParamNames []string) string,
) error {

	rawParamNames := linq.Map(typeParamReplacementNodes, func(tParamNode *SymbolNode) string {
		if tParamNode.Kind.IsBuiltin() {
			return tParamNode.Id.Name
		}
		return tParamNode.Data.(*TypeParamMeta).Name
	})

	if modelNameTransformer != nil {
		reducedStruct.Name = modelNameTransformer(reducedStruct.Name, rawParamNames)
	} else {
		reducedStruct.Name = StandardModelNameTransformer(reducedStruct.Name, rawParamNames)
	}

	for fieldIdx, field := range rawStruct.Fields {
		// Check if this is a generic field
		if field.Type.Root != nil && field.Type.Root.Kind() == metadata.TypeRefKindParam {
			nameParts := strings.SplitN(field.Type.Name, "#", 2)
			if len(nameParts) != 2 {
				return fmt.Errorf(
					"failed to determine replacement to generic placeholder '%s' in field '%s'",
					field.Type.Name,
					field.Name,
				)
			}

			replParamIdx, err := strconv.ParseInt(nameParts[1], 10, 16)
			if err != nil {
				return fmt.Errorf(
					"failed to parse replacement index to generic placeholder '%s' in field '%s'",
					field.Type.Name,
					field.Name,
				)
			}

			// Re-write the type
			reducedStruct.Fields[fieldIdx].Type = rawParamNames[replParamIdx]
		}

	}

	return nil
}

func materializeInstantiationTarget(
	rawStruct *metadata.StructMeta,
	reducedStruct definitions.StructMetadata,
	target InitializationTarget,
	modelNameTransformer func(modelBaseName string, typeParamNames []string) string,
) (definitions.StructMetadata, error) {
	err := inPlaceInstantiateGenericModel(rawStruct, &reducedStruct, target.TypeParams, modelNameTransformer)
	if err != nil {
		return reducedStruct, fmt.Errorf("failed to construct fields for generic struct '%s' - %v", rawStruct.Name, err)
	}

	return reducedStruct, nil
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
