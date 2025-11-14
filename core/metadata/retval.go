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

	// This cannot actually error, for now - v.Type.Resolve(ctx) uses the same CacheLookupKey
	// so if the above call succeeds, this one will too
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
