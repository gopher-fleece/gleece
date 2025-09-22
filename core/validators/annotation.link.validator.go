package validators

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/common/language"
	"github.com/gopher-fleece/gleece/common/linq"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/validators/diagnostics"
)

type classifiedAttributes struct {
	route             annotations.Attribute
	path              []annotations.Attribute
	nonPathAttributes []annotations.Attribute
}

// AnnotationLinkValidator performs route/@Path/param cross-validation.
type AnnotationLinkValidator struct {
	receiver          *metadata.ReceiverMeta
	groupedAttributes classifiedAttributes
	funcParamNames    mapset.Set[string]
	urlParams         []string // Note we're using a slice here since we have to validate there aren't any duplicates
}

// NewAnnotationLinkValidator constructs the validator.
func NewAnnotationLinkValidator(recv *metadata.ReceiverMeta) (AnnotationLinkValidator, error) {
	if recv == nil {
		return AnnotationLinkValidator{}, errors.New("cannot construct an annotation link validator for a nil receiver")
	}

	classifiedAttrs, err := classifyAttributes(recv)
	if err != nil {
		return AnnotationLinkValidator{}, err
	}

	return AnnotationLinkValidator{
		receiver:          recv,
		groupedAttributes: classifiedAttrs,
		funcParamNames:    getReceiverParamsNameSet(recv),
		urlParams:         extractUrlParams(classifiedAttrs.route.Value),
	}, nil
}

// Validate runs the nine checks and returns resolved diagnostics.
func (v AnnotationLinkValidator) Validate() []diagnostics.ResolvedDiagnostic {
	diags := []diagnostics.ResolvedDiagnostic{}
	seenFuncParams := make(map[string]annotations.Attribute)

	// 1. Check @Route parameters for duplications and matching matching @Path attributes
	diags = append(diags, v.validateRoute()...)

	// 2. Validate @Path attributes against function parameters
	diags = append(diags, v.validatePathAnnotations(seenFuncParams)...)

	// 3. Validate other annotations against function parameters
	diags = append(diags, v.validateNonPathAnnotations(seenFuncParams)...)

	// 4. Ensure all function parameters are referenced
	diags = append(diags, v.validateAllReferenced(seenFuncParams)...)

	// 5. Deduplicate diagnostics.
	//
	// This is pretty damn ugly. Prop resolution may yield various errors like invalid casts.
	// Since we may use those funcs in a few separate locations, we may end up with duplicate diagnostics.
	// Need to improve on this.
	finalDiags := []diagnostics.ResolvedDiagnostic{}
	for _, diag := range diags {
		diagExists := slices.ContainsFunc(finalDiags, func(d diagnostics.ResolvedDiagnostic) bool {
			return d.Equal(diag)
		})

		if !diagExists {
			finalDiags = append(finalDiags, diag)
		}
	}

	return finalDiags
}

func (v AnnotationLinkValidator) validateRoute() []diagnostics.ResolvedDiagnostic {
	diags := []diagnostics.ResolvedDiagnostic{}

	pathAliases := []string{}
	for _, pathAttr := range v.groupedAttributes.path {
		alias, pathAliasDiag := v.getPathAliasOrName(pathAttr)
		if pathAliasDiag != nil {
			diags = append(diags, *pathAliasDiag)
		}
		pathAliases = append(pathAliases, alias)
	}

	if len(diags) > 0 {
		// If we've any diagnostics at this point, we can assume receiver is broken,
		// most likely due to invalid JSON syntax in the attribute properties.
		// In such cases, there's no real point in continuing.
		return diags
	}

	referencedParams := mapset.NewSet(pathAliases...)
	witnessedUrlParams := mapset.NewSet[string]()

	for _, urlParam := range v.urlParams {
		// Verify each URL param appears exactly once
		if witnessedUrlParams.Contains(urlParam) {
			diags = append(diags, diagnostics.NewDiagnostic(
				v.receiver.Annotations.FileName(),
				fmt.Sprintf("Duplicate URL parameter '%s'", urlParam),
				diagnostics.DiagLinkerDuplicateUrlParam,
				diagnostics.DiagnosticError,
				getRangeForUrlParam(v.groupedAttributes.route, urlParam),
			))
		} else {
			witnessedUrlParams.Add(urlParam)
		}

		if !referencedParams.Contains(urlParam) {
			diags = append(diags, diagnostics.NewDiagnostic(
				v.receiver.Annotations.FileName(),
				fmt.Sprintf("URL parameter '%s' does not have a corresponding @Path annotation", urlParam),
				diagnostics.DiagLinkerRouteMissingPath,
				diagnostics.DiagnosticError,
				getRangeForUrlParam(v.groupedAttributes.route, urlParam),
			))
		}
	}

	return diags
}

// validatePathAnnotations validates @Path annotations against the annotated function's parameters
// Parameter `seenFuncParams` represents function parameters seen, thus far, by the linker and
// is an ***In/Out*** parameter (modified by this method)
func (v AnnotationLinkValidator) validatePathAnnotations(
	seenFuncParams map[string]annotations.Attribute, // Modified by function
) []diagnostics.ResolvedDiagnostic {
	diags := []diagnostics.ResolvedDiagnostic{}

	seenRefValues := mapset.NewSet[string]()
	seenAliases := mapset.NewSet[string]()

	for _, pathAttr := range v.groupedAttributes.path {
		// Note that func params are referenced by the Value field, not the alias.
		expectedFuncParamName := pathAttr.Value

		// Check if the path references a known func parameter
		if !v.funcParamNames.Contains(expectedFuncParamName) {
			suggestion := getContextualAppendedSuggestion(
				expectedFuncParamName,
				v.funcParamNames.ToSlice(),
				common.MapKeys(seenFuncParams),
			)
			diags = append(diags, diagnostics.NewErrorDiagnostic(
				v.receiver.Annotations.FileName(),
				fmt.Sprintf("@Path '%s' is not a parameter of %s%s", expectedFuncParamName, v.receiver.Name, suggestion),
				diagnostics.DiagLinkerPathInvalidRef,
				pathAttr.GetValueRange(),
			))
		} else {
			// Check if the referenced parameter has already been linked by a different annotation
			if _, exists := seenFuncParams[expectedFuncParamName]; exists {
				diags = append(diags, diagnostics.NewErrorDiagnostic(
					v.receiver.Annotations.FileName(),
					fmt.Sprintf("Function parameter '%s' is referenced by multiple @Path attributes", expectedFuncParamName),
					diagnostics.DiagLinkerMultipleParameterRefs,
					pathAttr.GetValueRange(),
				))
			}
			seenFuncParams[expectedFuncParamName] = pathAttr
		}

		// Check if the Path value appears multiple times
		if seenRefValues.Contains(expectedFuncParamName) {
			diags = append(diags, diagnostics.NewErrorDiagnostic(
				v.receiver.Annotations.FileName(),
				fmt.Sprintf("Duplicate @Path parameter reference '%s'", expectedFuncParamName),
				diagnostics.DiagLinkerDuplicatePathParam,
				pathAttr.Comment.Range(),
			))
		} else {
			seenRefValues.Add(expectedFuncParamName)
		}

		pAlias, aliasDiag := v.getPathAliasOrDiag(pathAttr)
		if aliasDiag != nil {
			diags = append(diags, *aliasDiag)
		} else {
			// Check if the Path's alias (i.e. 'name' property) appears multiple times
			if pAlias != nil && *pAlias != "" {
				alias := *pAlias

				if seenAliases.Contains(alias) {
					diags = append(diags, diagnostics.NewErrorDiagnostic(
						v.receiver.Annotations.FileName(),
						fmt.Sprintf("Duplicate @Path parameter alias '%s'", alias),
						diagnostics.DiagLinkerDuplicatePathAliasRef,
						pathAttr.Comment.Range(),
					))
				} else {
					seenAliases.Add(alias)
				}

				// Check if the alias exists as a URL parameter
				if !slices.Contains(v.urlParams, alias) {
					suggestion := getContextualAppendedSuggestion(
						alias,
						v.urlParams,
						common.MapKeys(seenFuncParams),
					)
					diags = append(diags, diagnostics.NewErrorDiagnostic(
						v.receiver.Annotations.FileName(),
						fmt.Sprintf("Unknown @Path parameter alias '%s'%s", alias, suggestion),
						diagnostics.DiagLinkerPathInvalidRef,
						pathAttr.Comment.Range(),
					))
				}
			}
		}
	}

	return diags
}

func (v AnnotationLinkValidator) validateNonPathAnnotations(
	seenFuncParams map[string]annotations.Attribute, // Modified by function
) []diagnostics.ResolvedDiagnostic {
	diags := []diagnostics.ResolvedDiagnostic{}

	// Sort everything so diags are emitted in a deterministic order
	nonPathAttributes := v.groupedAttributes.nonPathAttributes
	slices.SortFunc(nonPathAttributes, func(a, b annotations.Attribute) int {
		return cmp.Compare(a.Name, b.Name)
	})

	for _, attr := range nonPathAttributes {
		if attr.Value == "" {
			// Potentially partial annotation. May be linted by a context free analyzer (i.e., not here)
			continue
		}

		if !v.funcParamNames.Contains(attr.Value) {
			suggestion := getContextualAppendedSuggestion(
				attr.Value,
				v.funcParamNames.ToSlice(),
				common.MapKeys(seenFuncParams),
			)

			diags = append(diags, diagnostics.NewErrorDiagnostic(
				v.receiver.Annotations.FileName(),
				fmt.Sprintf(
					"@%s '%s' does not match any parameter of %s%s",
					attr.Name,
					attr.Value,
					v.receiver.Name,
					suggestion,
				),
				diagnostics.DiagLinkerPathInvalidRef,
				attr.GetValueRange(),
			))
		} else {
			seenFuncParams[attr.Value] = attr
		}
	}

	return diags
}

func (v AnnotationLinkValidator) validateAllReferenced(
	seenFuncParams map[string]annotations.Attribute, // Modified by function
) []diagnostics.ResolvedDiagnostic {
	diags := []diagnostics.ResolvedDiagnostic{}

	for _, paramName := range v.funcParamNames.ToSlice() {
		if _, wasSeen := seenFuncParams[paramName]; !wasSeen {
			matchingParam := linq.First(v.receiver.Params, func(p metadata.FuncParam) bool {
				return p.Name == paramName
			})
			if matchingParam == nil {
				// This should *never* happen. If it does, well, ignore- it'll break linting but really, should never happen.
				continue
			}

			if matchingParam.Type.IsContext() {
				// Ignore context params - they're unique in that they don't require any referencing
				continue
			}

			diags = append(diags, diagnostics.NewErrorDiagnostic(
				v.receiver.Annotations.FileName(),
				fmt.Sprintf(
					"Function parameter '%s' is not referenced by a parameter annotation",
					paramName,
				),
				diagnostics.DiagLinkerUnreferencedParameter,
				matchingParam.Range,
			))
		}
	}

	return diags
}

func (v AnnotationLinkValidator) getPathAliasOrDiag(attr annotations.Attribute) (*string, *diagnostics.ResolvedDiagnostic) {
	value, err := annotations.GetCastProperty[string](&attr, annotations.PropertyName)
	if err == nil {
		return value, nil
	}

	rawPropValue := attr.GetProperty(annotations.PropertyName)
	propValueStr := "Unknown value"
	if rawPropValue != nil {
		propValueStr = fmt.Sprintf("%v", *rawPropValue)
	}

	diag := diagnostics.NewErrorDiagnostic(
		v.receiver.Annotations.FileName(),
		fmt.Sprintf(
			"Invalid value for property 'name' in attribute %s ('%v')",
			attr.Name,
			propValueStr,
		),
		diagnostics.DiagAnnotationPropertiesInvalidValueForKey,
		attr.PropertiesRange,
	)
	return nil, &diag
}

func (v AnnotationLinkValidator) getPathAliasOrName(attr annotations.Attribute) (string, *diagnostics.ResolvedDiagnostic) {
	value, diag := v.getPathAliasOrDiag(attr)
	if diag != nil {
		return "", diag
	}

	if value != nil {
		return *value, nil
	}

	return attr.Value, nil
}

func getContextualAppendedSuggestion(input string, allOpts []string, alreadyUsedOpts []string) string {
	opts := mapset.NewSet(allOpts...)
	for _, usedOpt := range alreadyUsedOpts {
		opts.Remove(usedOpt)
	}

	suggestion := language.DidYouMean(input, opts.ToSlice())
	if suggestion == "" {
		return ""
	}

	return fmt.Sprintf(". Did you mean '%s'?", suggestion)
}

func classifyAttributes(receiver *metadata.ReceiverMeta) (classifiedAttributes, error) {
	classified := classifiedAttributes{
		path:              []annotations.Attribute{},
		nonPathAttributes: []annotations.Attribute{},
	}

	routeAttrSeen := false
	for _, attr := range receiver.Annotations.Attributes() {
		switch attr.Name {
		case annotations.GleeceAnnotationRoute:
			classified.route = attr
			routeAttrSeen = true
		case annotations.GleeceAnnotationPath:
			classified.path = append(classified.path, attr)
		case annotations.GleeceAnnotationQuery,
			annotations.GleeceAnnotationHeader,
			annotations.GleeceAnnotationBody,
			annotations.GleeceAnnotationFormField:
			if strings.TrimSpace(attr.Value) != "" {
				classified.nonPathAttributes = append(classified.nonPathAttributes, attr)
			}
		}
	}

	if routeAttrSeen {
		return classified, nil
	}

	// A bit of a violation of single responsibility here.
	// We just don't want to proceed without a @Route - it implies something has gone horribly wrong
	// since the visitor/gathering stage should not have picked up any method that does not have a @Route
	return classified, fmt.Errorf("receiver '%s' does not have route annotation", receiver.Name)
}

func getReceiverParamsNameSet(receiver *metadata.ReceiverMeta) mapset.Set[string] {
	names := linq.Map(receiver.Params, func(param metadata.FuncParam) string {
		return param.Name
	})

	return mapset.NewSet(names...)
}

// extractUrlParams returns ordered {param} names
func extractUrlParams(route string) []string {
	if route == "" {
		return nil
	}
	out := []string{}
	start := -1
	for i, ch := range route {
		if ch == '{' {
			start = i
			continue
		}
		if ch == '}' && start >= 0 {
			out = append(out, route[start+1:i])
			start = -1
		}
	}
	return out
}

// getRangeForUrlParam returns a ResolvedRange for the first {param} occurrence
func getRangeForUrlParam(attr annotations.Attribute, param string) common.ResolvedRange {
	paramIdx := strings.Index(attr.Value, "{"+param+"}")
	if paramIdx < 0 {
		// Fallback: just return the full value range
		return attr.GetValueRange()
	}

	start := attr.GetValueRange()
	return common.ResolvedRange{
		StartLine: start.StartLine,
		StartCol:  start.StartCol + utf8.RuneCountInString(attr.Value[:paramIdx]),
		EndLine:   start.EndLine,
		EndCol:    start.StartCol + utf8.RuneCountInString(attr.Value[:paramIdx+len(param)+2]),
	}
}
