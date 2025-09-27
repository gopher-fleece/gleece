package metadata

import (
	"fmt"
	"strings"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
)

type TypeUsageMeta struct {
	SymNodeMeta
	TypeParams []TypeUsageMeta // Generic type parameters like TKey/TValue in map[TKey]TValue
	Import     common.ImportType
	Layers     []TypeLayer
}

func (t TypeUsageMeta) GetBaseTypeRefKey() (graphs.SymbolKey, error) {
	if len(t.Layers) == 0 {
		return graphs.SymbolKey{}, fmt.Errorf("TypeUsageMeta has no layers")
	}
	baseRef := t.Layers[len(t.Layers)-1].BaseTypeRef
	if baseRef == nil {
		return graphs.SymbolKey{}, fmt.Errorf("BaseTypeRef is nil on last TypeLayer")
	}
	return *baseRef, nil
}

func (t TypeUsageMeta) GetArrayLayersString() string {
	// Currently we only use arrays for spec generation
	arrayCount := 0
	for _, layer := range t.Layers {
		if layer.Kind == TypeLayerKindArray {
			arrayCount++
		}
	}

	return strings.Repeat("[]", arrayCount)
}

func (t TypeUsageMeta) Resolve(ctx ReductionContext) (definitions.TypeMetadata, error) {
	typeRef, err := t.GetBaseTypeRefKey()
	if err != nil {
		return definitions.TypeMetadata{}, err
	}

	underlyingEnum := ctx.MetaCache.GetEnum(typeRef)

	alias := definitions.AliasMetadata{}
	if underlyingEnum != nil {
		alias.Name = underlyingEnum.Name
		alias.AliasType = string(underlyingEnum.ValueKind)

		values := []string{}
		for _, v := range underlyingEnum.Values {
			values = append(values, fmt.Sprintf("%v", v.Value))
		}
		alias.Values = values
	}

	description := ""
	if t.Annotations != nil {
		description = t.Annotations.GetDescription()
	}

	// Join the actual type name with its "[]" prefixes, as necessary.
	// Ugly, but the spec generator uses that - for now.
	name := t.GetArrayLayersString() + t.Name

	return definitions.TypeMetadata{
		Name:                name,
		PkgPath:             t.PkgPath,
		DefaultPackageAlias: gast.GetDefaultPkgAliasByName(t.PkgPath),
		Description:         description,
		Import:              t.Import,
		IsUniverseType:      t.PkgPath == "" && gast.IsUniverseType(t.Name),
		IsByAddress:         t.IsByAddress(),
		SymbolKind:          t.SymbolKind,
		AliasMetadata:       &alias,
	}, nil
}

func (t TypeUsageMeta) IsContext() bool {
	return t.Name == "Context" && t.PkgPath == "context"
}
