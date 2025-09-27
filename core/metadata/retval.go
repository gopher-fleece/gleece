package metadata

import "github.com/gopher-fleece/gleece/definitions"

type FuncReturnValue struct {
	SymNodeMeta
	Ordinal int
	Type    TypeUsageMeta
}

func (v FuncReturnValue) Reduce(ctx ReductionContext) (definitions.FuncReturnValue, error) {
	typeMeta, err := v.Type.Resolve(ctx)
	if err != nil {
		return definitions.FuncReturnValue{}, err
	}

	typeRef, err := v.Type.GetBaseTypeRefKey()
	if err != nil {
		return definitions.FuncReturnValue{}, err
	}

	return definitions.FuncReturnValue{
		Ordinal:            v.Ordinal,
		UniqueImportSerial: ctx.SyncedProvider.GetIdForKey(typeRef),
		TypeMetadata:       typeMeta,
	}, nil
}
