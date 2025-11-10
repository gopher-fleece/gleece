package metadata

import (
	"fmt"

	"github.com/gopher-fleece/gleece/definitions"
)

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

	symKey, err := v.Type.Root.CacheLookupKey(v.FVersion)
	if err != nil {
		return definitions.FuncReturnValue{}, fmt.Errorf("failed to derive a symbol key for field '%s' - %v", v.Name, err)
	}

	return definitions.FuncReturnValue{
		Ordinal:            v.Ordinal,
		UniqueImportSerial: ctx.SyncedProvider.GetIdForKey(symKey),
		TypeMetadata:       typeMeta,
	}, nil
}
