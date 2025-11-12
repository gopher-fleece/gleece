package metadata

import (
	"fmt"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/gast"
)

type TypeUsageMeta struct {
	SymNodeMeta
	TypeParams []TypeUsageMeta // Generic type parameters like TKey/TValue in map[TKey]TValue
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
		//symKind, err := getUniverseSymKind(symKey.Name)
		//	if err != nil {
		//		return definitions.TypeMetadata{}, err
		//	}

		// t.SymKind is kind of the usage's kind - not necessarily the underlying type.
		// This is some major spaghetti and needs to be improved later - coherency is lacking.
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

/*
func getUniverseSymKind(typeName string) (common.SymKind, error) {
	if _, ok := symboldg.ToPrimitiveType(typeName); ok {
		return common.SymKindBuiltin, nil
	}

	if _, ok := symboldg.ToSpecialType(typeName); ok {
		return common.SymKindSpecialBuiltin, nil
	}

	return common.SymKindUnknown, fmt.Errorf("typename '%s' is neither a primitive nor a special type", typeName)
}
*/
