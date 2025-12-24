package metadata

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/definitions"
)

type ReceiverMeta struct {
	SymNodeMeta
	Params  []FuncParam
	RetVals []FuncReturnValue
}

func (m ReceiverMeta) Reduce(
	ctx ReductionContext,
	parentSecurity []definitions.RouteSecurity,
) (definitions.RouteMetadata, error) {

	verbAnnotation := m.Annotations.GetFirst(annotations.GleeceAnnotationMethod)
	if verbAnnotation == nil || verbAnnotation.Value == "" {
		// Not ideal- we'd like to separate visitation, reduction and validation but typing currently doesn't
		// cleanly allow it so we have to embed a bit of validation in a few other places as well
		return definitions.RouteMetadata{}, fmt.Errorf("receiver %s has not @Method annotation", m.Name)
	}

	security, err := GetRouteSecurityWithInheritance(m.Annotations, parentSecurity)
	if err != nil {
		return definitions.RouteMetadata{}, err
	}

	templateCtx, err := GetTemplateContextMetadata(m.Annotations)
	if err != nil {
		return definitions.RouteMetadata{}, err
	}

	hasReturnValue := len(m.RetVals) > 1

	responses := []definitions.FuncReturnValue{}
	for _, fRetVal := range m.RetVals {
		response, err := fRetVal.Reduce(ctx)
		if err != nil {
			return definitions.RouteMetadata{}, err
		}
		responses = append(responses, response)
	}

	reducedParams := []definitions.FuncParam{}
	for _, param := range m.Params {
		reducedParam, err := param.Reduce(ctx)
		if err != nil {
			return definitions.RouteMetadata{}, err
		}
		reducedParams = append(reducedParams, reducedParam)
	}

	successResponseCode, successResponseDescription, err := GetResponseStatusCodeAndDescription(m.Annotations, hasReturnValue)
	if err != nil {
		return definitions.RouteMetadata{}, err
	}

	errorResponses, err := GetErrorResponses(m.Annotations)
	if err != nil {
		return definitions.RouteMetadata{}, err
	}

	return definitions.RouteMetadata{
		OperationId: m.Name,
		HttpVerb:    definitions.HttpVerb(verbAnnotation.Value),
		Hiding:      GetMethodHideOpts(m.Annotations),
		Deprecation: GetDeprecationOpts(m.Annotations),
		Description: m.Annotations.GetDescription(),
		RestMetadata: definitions.RestMetadata{
			Path: m.Annotations.GetFirstValueOrEmpty(annotations.GleeceAnnotationRoute),
		},
		HasReturnValue:      hasReturnValue,
		RequestContentType:  definitions.ContentTypeJSON, // Hardcoded for now, should be supported via annotations later on
		ResponseContentType: definitions.ContentTypeJSON, // Hardcoded for now, should be supported via annotations later on
		Security:            security,
		TemplateContext:     templateCtx,
		ResponseSuccessCode: successResponseCode,
		ResponseDescription: successResponseDescription,
		FuncParams:          reducedParams,
		Responses:           responses,
		ErrorResponses:      errorResponses,
	}, nil
}

func (v ReceiverMeta) RetValsRange() common.ResolvedRange {
	switch len(v.RetVals) {
	case 0:
		return common.ResolvedRange{}
	case 1:
		return common.ResolvedRange{
			StartLine: v.RetVals[0].Range.StartLine,
			EndLine:   v.RetVals[0].Range.EndLine,
			StartCol:  v.RetVals[0].Range.StartCol,
			EndCol:    v.RetVals[0].Range.EndCol,
		}
	default:
		// Copy so as not to affect original order
		retVals := append([]FuncReturnValue{}, v.RetVals...)
		slices.SortFunc(retVals, func(valA, valB FuncReturnValue) int {
			return cmp.Compare(valA.Ordinal, valB.Ordinal)
		})

		lastIndex := len(retVals) - 1
		return common.ResolvedRange{
			StartLine: retVals[0].Range.StartLine,
			EndLine:   retVals[lastIndex].Range.EndLine,
			StartCol:  retVals[0].Range.StartCol,
			EndCol:    retVals[lastIndex].Range.EndCol,
		}
	}
}
