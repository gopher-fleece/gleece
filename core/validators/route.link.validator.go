package validators

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/common/language"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/validators/diagnostics"
)

// AnnotationLinkValidator performs route/@Path/param cross-validation.
type AnnotationLinkValidator struct {
	receiver *metadata.ReceiverMeta
}

// NewAnnotationLinkValidator constructs the validator.
func NewAnnotationLinkValidator(recv *metadata.ReceiverMeta) AnnotationLinkValidator {
	return AnnotationLinkValidator{
		receiver: recv,
	}
}

// Validate runs the nine checks and returns resolved diagnostics.
func (v AnnotationLinkValidator) Validate() []diagnostics.ResolvedDiagnostic {
	var out []diagnostics.ResolvedDiagnostic

	// build func param set and a seen map (which attribute referenced which param)
	funcParamSet := map[string]struct{}{}
	seenFuncParams := map[string]*annotations.Attribute{}
	for _, p := range v.receiver.Params {
		funcParamSet[p.Name] = struct{}{}
	}

	routeAttr, pathAttrs, nonPathAttrs := v.classifyAttributes()

	// 1 & 2: route duplicates + missing @Path
	if routeAttr != nil {
		out = append(out, v.validateRoute(pathAttrs, routeAttr)...)
	}

	// 3,4,5,6,7: path annotation checks
	out = append(out, v.validatePathAnnotations(routeAttr, pathAttrs, funcParamSet, seenFuncParams)...)

	// 8: non-path annotations (Query/Header/Body/FormField)
	out = append(out, v.validateNonPathAnnotations(nonPathAttrs, funcParamSet, seenFuncParams)...)

	// 9: unreferenced params (skip context.Context)
	for _, p := range v.receiver.Params {
		if p.Type.IsContext() {
			continue
		}
		if _, seen := seenFuncParams[p.Name]; !seen {
			out = append(out, diagnostics.NewDiagnostic(
				v.receiver.Annotations.FileName(),
				fmt.Sprintf("Function parameter '%s' is not referenced by a parameter annotation", p.Name),
				diagnostics.DiagLinkerUnreferencedParameter,
				diagnostics.DiagnosticError,
				p.Range,
			))
		}
	}

	return out
}

// classifyAttributes groups attributes like the TS version.
func (v AnnotationLinkValidator) classifyAttributes() (
	routeAttribute *annotations.Attribute,
	pathAttributes []annotations.Attribute,
	nonPathAttributes map[string]annotations.Attribute,
) {
	pathAttributes = []annotations.Attribute{}
	nonPathAttributes = map[string]annotations.Attribute{}

	for _, attr := range v.receiver.Annotations.Attributes() {
		switch attr.Name {
		case annotations.GleeceAnnotationRoute:
			if routeAttribute == nil {
				routeAttribute = &attr
			}
		case annotations.GleeceAnnotationPath:
			pathAttributes = append(pathAttributes, attr)
		case annotations.GleeceAnnotationQuery,
			annotations.GleeceAnnotationHeader,
			annotations.GleeceAnnotationBody,
			annotations.GleeceAnnotationFormField:
			if strings.TrimSpace(attr.Value) != "" {
				nonPathAttributes[attr.Value] = attr
			}
		}
	}
	return
}

// validateRoute: duplicate url params (warning) + missing @Path alias (error)
func (v AnnotationLinkValidator) validateRoute(
	pathAttributes []annotations.Attribute,
	routeAttr *annotations.Attribute,
) []diagnostics.ResolvedDiagnostic {
	var out []diagnostics.ResolvedDiagnostic
	if routeAttr == nil {
		return out
	}
	urlParams := extractUrlParams(routeAttr.Value)

	// BROKEN - NEED TO REIMPORT VALIDATE
	// build set of aliases referenced by @Path
	refAliases := map[string]struct{}{}
	for _, pathAttrib := range pathAttributes {
		if alias := getPathAnnotationAlias(&pathAttrib); alias != "" {
			refAliases[alias] = struct{}{}
		}
	}

	seen := map[string]bool{}
	for _, param := range urlParams {
		if seen[param] {
			out = append(out, diagnostics.NewDiagnostic(
				v.receiver.Annotations.FileName(),
				fmt.Sprintf("Duplicate route parameter '%s'", param),
				diagnostics.DiagLinkerDuplicatePathParam,
				diagnostics.DiagnosticWarning,
				getRangeForUrlParam(routeAttr, param),
			))
		}
		seen[param] = true

		if _, ok := refAliases[param]; !ok {
			out = append(out, diagnostics.NewDiagnostic(
				fmt.Sprintf("%s (ln. %d)", v.receiver.Annotations.FileName(), routeAttr.GetValueRange().StartLine),
				fmt.Sprintf("Route parameter '%s' does not have a corresponding @Path annotation", param),
				diagnostics.DiagLinkerRouteMissingPath,
				diagnostics.DiagnosticError,
				getRangeForUrlParam(routeAttr, param),
			))
		}
	}
	return out
}

// validatePathAnnotations implements the TS logic for @Path checks (3,4,5,6,7)
func (v AnnotationLinkValidator) validatePathAnnotations(
	routeAttr *annotations.Attribute,
	pathAttributes []annotations.Attribute,
	funcParamSet map[string]struct{},
	seenFuncParams map[string]*annotations.Attribute, // OUT param
) []diagnostics.ResolvedDiagnostic {
	var out []diagnostics.ResolvedDiagnostic
	seenRefValues := map[string]struct{}{}
	seenAliases := map[string]struct{}{}

	// build url param set for alias checks
	urlParamSet := map[string]struct{}{}
	if routeAttr != nil {
		for _, n := range extractUrlParams(routeAttr.Value) {
			urlParamSet[n] = struct{}{}
		}
	}

	for i := range pathAttributes {
		attr := &pathAttributes[i]
		expectedFuncParamName := strings.TrimSpace(attr.Value)
		attrRange := attr.Comment.Range()

		// incomplete attribute (missing value)
		if expectedFuncParamName == "" {
			out = append(out, diagnostics.NewDiagnostic(
				v.receiver.Annotations.FileName(),
				"@Path missing value",
				diagnostics.DiagLinkerIncompleteAttribute,
				diagnostics.DiagnosticError,
				attrRange,
			))
			continue
		}

		// 3: value must be a function parameter
		if _, ok := funcParamSet[expectedFuncParamName]; !ok {
			suggestion := getContextualSuggestion(expectedFuncParamName, funcParamSet, mapKeysToSlice(seenFuncParams))
			msg := fmt.Sprintf("@Path '%s' is not a parameter of %s", expectedFuncParamName, v.receiver.Name)
			if suggestion != "" {
				msg += ". Did you mean '" + suggestion + "'?"
			}
			out = append(out, diagnostics.NewDiagnostic(
				v.receiver.Annotations.FileName(),
				msg,
				diagnostics.DiagLinkerPathInvalidRef,
				diagnostics.DiagnosticError,
				attrRange,
			))
		} else {
			// 4: parameter referenced by multiple param annotations?
			if prev := seenFuncParams[expectedFuncParamName]; prev != nil {
				out = append(out, diagnostics.NewDiagnostic(
					v.receiver.Annotations.FileName(),
					fmt.Sprintf("Function parameter '%s' is referenced by multiple @Path attributes", expectedFuncParamName),
					diagnostics.DiagLinkerMultipleParameterRefs,
					diagnostics.DiagnosticError,
					attrRange,
				))
			}
			seenFuncParams[expectedFuncParamName] = attr
		}

		// 5: duplicate @Path value (warning)
		if _, dup := seenRefValues[expectedFuncParamName]; dup {
			out = append(out, diagnostics.NewDiagnostic(
				v.receiver.Annotations.FileName(),
				fmt.Sprintf("Duplicate @Path parameter reference '%s'", expectedFuncParamName),
				diagnostics.DiagLinkerDuplicatePathParamRef,
				diagnostics.DiagnosticWarning,
				attrRange,
			))
		} else {
			seenRefValues[expectedFuncParamName] = struct{}{}
		}

		// 6 & 7: alias checks
		alias := getPathAnnotationAlias(attr)
		if alias != "" {
			if _, dup := seenAliases[alias]; dup {
				out = append(out, diagnostics.NewDiagnostic(
					v.receiver.Annotations.FileName(),
					fmt.Sprintf("Duplicate @Path parameter alias '%s'", alias),
					diagnostics.DiagLinkerDuplicatePathAliasRef,
					diagnostics.DiagnosticWarning,
					attrRange,
				))
			}
			seenAliases[alias] = struct{}{}

			if _, ok := urlParamSet[alias]; !ok {
				suggestion := getAppendedSuggestion(alias, mapKeys(urlParamSet))
				msg := fmt.Sprintf("Unknown @Path parameter alias '%s'", alias)
				if suggestion != "" {
					msg = msg + suggestion
				}
				out = append(out, diagnostics.NewDiagnostic(
					v.receiver.Annotations.FileName(),
					msg,
					diagnostics.DiagLinkerPathInvalidRef,
					diagnostics.DiagnosticError,
					attrRange,
				))
			}
		}
	}

	return out
}

// validateNonPathAnnotations implements TS validateNonPathAnnotations (8)
func (v AnnotationLinkValidator) validateNonPathAnnotations(
	nonPathAttributes map[string]annotations.Attribute,
	funcParamSet map[string]struct{},
	seenFuncParams map[string]*annotations.Attribute, // OUT param
) []diagnostics.ResolvedDiagnostic {
	var out []diagnostics.ResolvedDiagnostic

	for param, attr := range nonPathAttributes {
		if param == "" {
			continue
		}
		attrRange := attr.Comment.Range()

		if _, ok := funcParamSet[param]; !ok {
			suggestion := getContextualSuggestion(param, funcParamSet, mapKeysToSlice(seenFuncParams))
			msg := fmt.Sprintf("@%s '%s' does not match any parameter of %s", attr.Name, param, v.receiver.Name)
			if suggestion != "" {
				msg += ". Did you mean '" + suggestion + "'?"
			}
			out = append(out, diagnostics.NewDiagnostic(
				v.receiver.Annotations.FileName(),
				msg,
				diagnostics.DiagLinkerPathInvalidRef,
				diagnostics.DiagnosticError,
				attrRange,
			))
			continue
		}
		seenFuncParams[param] = &attr
	}

	return out
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
func getRangeForUrlParam(attr *annotations.Attribute, param string) common.ResolvedRange {
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

func getPathAnnotationAlias(attr *annotations.Attribute) string {
	if attr == nil || attr.Properties == nil {
		return ""
	}
	if v, ok := attr.Properties[annotations.PropertyName]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	if v, ok := attr.Properties["name"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func mapKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func mapKeysToSlice(attrMap map[string]*annotations.Attribute) []string {
	out := make([]string, 0, len(attrMap))
	for k := range attrMap {
		out = append(out, k)
	}
	return out
}

func getAppendedSuggestion(input string, options []string) string {
	if len(options) == 0 {
		return ""
	}
	best := language.DidYouMean(input, options)
	if best == "" {
		return ""
	}
	return fmt.Sprintf(". Did you mean '%s'?", best)
}

func getContextualSuggestion(input string, funcParamSet map[string]struct{}, alreadyUsed []string) string {
	all := make([]string, 0, len(funcParamSet))
	used := map[string]struct{}{}
	for _, u := range alreadyUsed {
		used[u] = struct{}{}
	}
	for k := range funcParamSet {
		if _, skip := used[k]; !skip {
			all = append(all, k)
		}
	}
	return language.DidYouMean(input, all)
}
