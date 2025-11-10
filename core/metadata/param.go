package metadata

import (
	"fmt"

	"github.com/gopher-fleece/gleece/definitions"
)

type FuncParam struct {
	SymNodeMeta
	Ordinal int
	Type    TypeUsageMeta
}

func (v FuncParam) Reduce(ctx ReductionContext) (definitions.FuncParam, error) {
	typeMeta, err := v.Type.Resolve(ctx)
	if err != nil {
		return definitions.FuncParam{}, err
	}

	var nameInSchema string
	var passedIn definitions.ParamPassedIn
	var validator string

	isContext := v.Type.IsContext()

	if !isContext {
		nameInSchema, err = GetParameterSchemaName(v.Name, v.Annotations)
		if err != nil {
			return definitions.FuncParam{}, err
		}

		passedIn, err = GetParamPassedIn(v.Name, v.Annotations)
		if err != nil {
			return definitions.FuncParam{}, err
		}

		validator, err = GetParamValidator(v.Name, v.Annotations, passedIn, v.Type.IsByAddress())
		if err != nil {
			return definitions.FuncParam{}, err
		}
	}

	// Find the parameter's attribute in the receiver's annotations
	var paramDescription string
	paramAttrib := v.Annotations.FindFirstByValue(v.Name)
	if paramAttrib != nil {
		// Note that nil here is not valid and should be rejected at the validation stage
		paramDescription = paramAttrib.Description
	}

	symKey, err := v.Type.Root.CacheLookupKey(v.FVersion)
	if err != nil {
		return definitions.FuncParam{}, fmt.Errorf(
			"failed to derive symbol key for function parameter '%s' - %v",
			v.Name,
			err,
		)
	}

	return definitions.FuncParam{
		ParamMeta: definitions.ParamMeta{
			Name:      v.Name,
			Ordinal:   v.Ordinal,
			TypeMeta:  typeMeta,
			IsContext: isContext,
		},
		PassedIn:           passedIn,
		NameInSchema:       nameInSchema,
		Description:        paramDescription,
		UniqueImportSerial: ctx.SyncedProvider.GetIdForKey(symKey),
		Validator:          validator,
		Deprecation:        GetDeprecationOpts(v.Annotations),
	}, nil
}
