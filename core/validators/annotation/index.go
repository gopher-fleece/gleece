package annotation

import (
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/annotations"
)

// RouteAnnotationIndex: focused graph for route/path checks.
type RouteAnnotationIndex struct {
	holder      *annotations.AnnotationHolder
	route       *routeNode
	pathNodes   []*pathNode          // source order
	pathByValue map[string]*pathNode // value -> node
	pathByAlias map[string]*pathNode // alias -> node
}

// NewRouteAnnotationIndex builds the graph in one pass.
func NewRouteAnnotationIndex(holder *annotations.AnnotationHolder) *RouteAnnotationIndex {
	g := &RouteAnnotationIndex{
		holder:      holder,
		pathByValue: map[string]*pathNode{},
		pathByAlias: map[string]*pathNode{},
	}
	g.build()
	return g
}

func (g *RouteAnnotationIndex) build() {
	attrs := g.holder.Attributes()
	for i := range attrs {
		switch attrs[i].Name {
		case annotations.GleeceAnnotationPath:
			g.processPathAttrib(&attrs[i])
		case annotations.GleeceAnnotationRoute:
			g.processRouteAttrib(&attrs[i])
		}
	}
}

func (g *RouteAnnotationIndex) processPathAttrib(attr *annotations.Attribute) {
	nPath := &pathNode{
		Attr:  attr,
		Value: attr.Value,
		Alias: "",
		Range: attr.Comment.Range(),
		File:  g.holder.FileName(),
	}

	if attr.Properties != nil {
		if propValue, ok := attr.Properties[annotations.PropertyName]; ok {
			if s, ok := propValue.(string); ok {
				nPath.Alias = s
			}
		} else if propValue, ok := attr.Properties["name"]; ok {
			if s, ok := propValue.(string); ok {
				nPath.Alias = s
			}
		}
	}

	g.pathNodes = append(g.pathNodes, nPath)
	if nPath.Value != "" {
		g.pathByValue[nPath.Value] = nPath
	}
	if nPath.Alias != "" {
		g.pathByAlias[nPath.Alias] = nPath
	}
}

func (g *RouteAnnotationIndex) processRouteAttrib(attr *annotations.Attribute) {
	r := &routeNode{Raw: attr.Value, Range: attr.Comment.Range(), Params: nil}
	
	for _, matchIdx := range routeParamRe.FindAllStringSubmatchIndex(attr.Value, -1) {
		// mi: [fullStart, fullEnd, groupStart, groupEnd]
		startMatch := matchIdx[0]
		endMatch := matchIdx[1]
		startGroup := matchIdx[2]
		endGroup := matchIdx[3]
		paramName := attr.Value[startGroup:endGroup]
		paramStart := attr.Comment.Position.StartCol + startMatch
		paramEnd := attr.Comment.Position.StartCol + endMatch
		r.Params = append(r.Params, routeParam{
			Name: paramName,
			Range: common.ResolvedRange{
				StartLine: attr.Comment.Position.StartLine,
				StartCol:  paramStart,
				EndLine:   attr.Comment.Position.EndLine,
				EndCol:    paramEnd,
			},
		})
	}
	g.route = r
}

// RouteParamNames returns ordered route parameter names (nil iff no route).
func (g *RouteAnnotationIndex) RouteParamNames() []string {
	if g.route == nil {
		return nil
	}
	out := make([]string, 0, len(g.route.Params))
	for _, p := range g.route.Params {
		out = append(out, p.Name)
	}
	return out
}

// HasRouteParam reports membership quickly.
func (g *RouteAnnotationIndex) HasRouteParam(name string) bool {
	if g.route == nil {
		return false
	}
	for _, p := range g.route.Params {
		if p.Name == name {
			return true
		}
	}
	return false
}

// PathByAlias returns the @Path node with that alias, or nil.
func (g *RouteAnnotationIndex) PathByAlias(alias string) *pathNode {
	if alias == "" {
		return nil
	}
	return g.pathByAlias[alias]
}

// PathByValue returns the @Path node referencing that value (function param name), or nil.
func (g *RouteAnnotationIndex) PathByValue(value string) *pathNode {
	if value == "" {
		return nil
	}
	return g.pathByValue[value]
}

// DuplicateRouteParams returns any route params that repeat (each duplicate once).
func (g *RouteAnnotationIndex) DuplicateRouteParams() []routeParam {
	out := []routeParam{}
	if g.route == nil {
		return out
	}
	seen := map[string]int{}
	for _, p := range g.route.Params {
		seen[p.Name]++
	}
	for _, p := range g.route.Params {
		if seen[p.Name] > 1 {
			out = append(out, p)
			seen[p.Name] = 0 // only report once per group
		}
	}
	return out
}

// DuplicatePathValues returns path nodes that share the same Value (function param).
func (g *RouteAnnotationIndex) DuplicatePathValues() []*pathNode {
	count := map[string]int{}
	for _, pn := range g.pathNodes {
		if pn.Value != "" {
			count[pn.Value]++
		}
	}
	out := []*pathNode{}
	for _, pn := range g.pathNodes {
		if pn.Value != "" && count[pn.Value] > 1 {
			out = append(out, pn)
		}
	}
	return out
}

// DuplicatePathAliases returns path nodes that reuse the same alias.
func (g *RouteAnnotationIndex) DuplicatePathAliases() []*pathNode {
	count := map[string]int{}
	for _, pn := range g.pathNodes {
		if pn.Alias != "" {
			count[pn.Alias]++
		}
	}
	out := []*pathNode{}
	for _, pn := range g.pathNodes {
		if pn.Alias != "" && count[pn.Alias] > 1 {
			out = append(out, pn)
		}
	}
	return out
}

// RouteAliasesMissingPath returns route params that lack a matching @Path alias.
func (g *RouteAnnotationIndex) RouteAliasesMissingPath() []routeParam {
	out := []routeParam{}
	if g.route == nil {
		return out
	}
	// build set of aliases
	aliasSet := map[string]struct{}{}
	for a := range g.pathByAlias {
		aliasSet[a] = struct{}{}
	}
	for _, rp := range g.route.Params {
		if _, ok := aliasSet[rp.Name]; !ok {
			out = append(out, rp)
		}
	}
	return out
}

// UnreferencedFuncParams returns func params not referenced by any @Path (value).
// caller can combine with non-path checks to refine.
func (g *RouteAnnotationIndex) UnreferencedFuncParams(funcParams []string) []string {
	out := []string{}
	seen := map[string]struct{}{}
	for _, pn := range g.pathNodes {
		if pn.Value != "" {
			seen[pn.Value] = struct{}{}
		}
	}
	for _, p := range funcParams {
		if _, ok := seen[p]; !ok {
			out = append(out, p)
		}
	}
	return out
}
