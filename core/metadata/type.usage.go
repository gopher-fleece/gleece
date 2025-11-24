package metadata

import (
	"fmt"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/gast"
)

type TypeUsageMeta struct {
	SymNodeMeta
	Import     common.ImportType
	// Describes the actual type reference.
	// Recursively encodes information such as 'is this a pointer type?' or 'what generic parameters does this usage have?'
	Root TypeRef
}

func (t TypeUsageMeta) Resolve(ctx ReductionContext) (definitions.TypeMetadata, error) {
	symKey, err := t.Root.CacheLookupKey(t.FVersion)
	if err != nil {
		return definitions.TypeMetadata{}, fmt.Errorf(
			"failed to derive cache lookup symbol key for type usage '%s' - %v",
			t.Name,
			err,
		)
	}

	if symKey.IsUniverse {
		return definitions.TypeMetadata{
			Name:           t.Root.SimpleTypeString(),
			Import:         common.ImportTypeNone,
			IsUniverseType: true,
			IsByAddress:    t.IsByAddress(),
			SymbolKind:     t.SymbolKind,
			AliasMetadata:  nil,
		}, nil
	}

	// Check if this usage is for an enum, if so, flatten it to an AliasMetadata
	underlyingEnum := ctx.MetaCache.GetEnum(symKey)

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

	name := t.Root.SimpleTypeString()
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
