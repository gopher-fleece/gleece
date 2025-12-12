package metadata

import (
	"fmt"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
)

type TypeUsageMeta struct {
	SymNodeMeta
	Import common.ImportType
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

	return definitions.TypeMetadata{
		Name:                t.Root.SimpleTypeString(),
		PkgPath:             t.PkgPath,
		DefaultPackageAlias: gast.GetDefaultPkgAliasByName(t.PkgPath),
		Description:         annotations.GetDescription(t.Annotations),
		Import:              t.Import,
		IsUniverseType:      t.PkgPath == "" && gast.IsUniverseType(t.Name),
		IsByAddress:         t.IsByAddress(),
		SymbolKind:          t.SymbolKind,
		AliasMetadata:       common.Ptr(getAliasMeta(ctx, symKey)),
	}, nil
}

func (t TypeUsageMeta) IsContext() bool {
	return t.Name == "Context" && t.PkgPath == "context"
}

// getAliasMeta attempts to retrieve Alias metadata for the given type.
// Alias metadata is relevant for enums and true aliases. For other kinds, it'll remain empty.
func getAliasMeta(
	ctx ReductionContext,
	typeSymKey graphs.SymbolKey,
) definitions.AliasMetadata {
	// Take the actual data from the cache. Quite hacky but good enough for now.
	alias := definitions.AliasMetadata{}

	// Check if this usage is for an enum, if so, flatten it to an AliasMetadata
	underlyingEnum := ctx.MetaCache.GetEnum(typeSymKey)

	if underlyingEnum != nil {
		// This is an enum
		alias.Name = underlyingEnum.Name
		alias.AliasType = string(underlyingEnum.ValueKind)

		values := []string{}
		for _, v := range underlyingEnum.Values {
			values = append(values, fmt.Sprintf("%v", v.Value))
		}
		alias.Values = values
	} else {
		// If the type is not an enum, check if it's an alias
		underlyingAlias := ctx.MetaCache.GetAlias(typeSymKey)
		if underlyingAlias != nil {
			// This is an alias
			alias.Name = underlyingAlias.Name
			alias.AliasType = underlyingAlias.Type.Root.SimpleTypeString()
		}
	}

	return alias
}
