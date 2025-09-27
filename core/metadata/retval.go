package metadata

import "github.com/gopher-fleece/gleece/definitions"

type FuncReturnValue struct {
	SymNodeMeta
	Ordinal int
	Type    TypeUsageMeta
}

func (v FuncReturnValue) Reduce(metaCache MetaCache, syncedProvider IdProvider) (definitions.FuncReturnValue, error) {
	typeMeta, err := v.Type.Resolve(metaCache)
	if err != nil {
		return definitions.FuncReturnValue{}, err
	}

	typeRef, err := v.Type.GetBaseTypeRefKey()
	if err != nil {
		return definitions.FuncReturnValue{}, err
	}

	return definitions.FuncReturnValue{
		Ordinal:            v.Ordinal,
		UniqueImportSerial: syncedProvider.GetIdForKey(typeRef),
		TypeMetadata:       typeMeta,
	}, nil
}
