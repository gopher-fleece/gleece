package metadata

import (
	"fmt"

	"github.com/gopher-fleece/gleece/v2/common"
	"github.com/gopher-fleece/gleece/v2/core/annotations"
	"github.com/gopher-fleece/gleece/v2/definitions"
	"github.com/gopher-fleece/gleece/v2/gast"
	"github.com/gopher-fleece/gleece/v2/graphs"
)

// Represents a type usage site's metadata.
//
// Example:
// The type usage for:
//
//	m := map[string]int
//
// is
//
//	map[string]int
//
// Metadata includes the usage's location in code, how it's imported and any modifiers like pointers/slices.
//
// This structure is part of the HIR in that it's linked to the source AST but also also serves as a high level abstraction that can be reduced
// to a flatter, simpler IR that's used by the downstream generators to emit routing and schema code.
type TypeUsageMeta struct {
	SymNodeMeta
	Import common.ImportType
	// Describes the actual type reference.
	// Recursively encodes information such as 'is this a pointer type?' or 'what generic parameters does this usage have?'
	Root TypeRef
}

// Reduce returns the IR for the type usage.
// This IR is a simpler and flatter TypeMetadata that can be used by downstream generators to emit the final code.
func (t TypeUsageMeta) Reduce(ctx ReductionContext) (definitions.TypeMetadata, error) {
	// First, we need to get the cache key for this usage. This allows us to better understand
	// what we're looking at and pull off related info from the graph.
	// Not an ideal situation but good enough for now.
	symKey, err := t.Root.CacheLookupKey(t.FVersion)
	if err != nil {
		return definitions.TypeMetadata{}, fmt.Errorf(
			"failed to derive cache lookup symbol key for type usage '%s' - %v",
			t.Name,
			err,
		)
	}

	// The Symbol Key is the easiest way to tell whether something is a 'universe' or 'builtin' type.
	// Universe are a special case as they never have PkgPath, imports, annotations etc.
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

// IsContext returns a boolean indicating whether this usage is of Go's context.Context.
// Context is a special object and has specific treatment at both visitation and validation layers.
func (t TypeUsageMeta) IsContext() bool {
	return t.Name == "Context" && t.PkgPath == "context"
}

// IsIterable returns a boolean indicating whether this usage is of a slice or an array
func (t TypeUsageMeta) IsIterable() bool {
	return t.Root.Kind() == TypeRefKindSlice || t.Root.Kind() == TypeRefKindArray
}

// getAliasMeta attempts to retrieve alias metadata for the given type.
// Returns a populated AliasMetadata, if the type is an enum or alias, otherwise returns an empty AliasMetadata
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
