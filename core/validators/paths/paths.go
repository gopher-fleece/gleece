package paths

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gopher-fleece/gleece/core/metadata"
)

type trieNode struct {
	literalChildren map[string]*trieNode
	paramChild      *trieNode
	// endpoints keyed by HTTP method (e.g. "GET", "POST")
	endpoint map[string]*RouteEntry
}

func newTrieNode() *trieNode {
	return &trieNode{
		literalChildren: map[string]*trieNode{},
		endpoint:        map[string]*RouteEntry{},
	}
}

type RouteEntryMeta struct {
	Controller *metadata.ControllerMeta
	Receiver   *metadata.ReceiverMeta
}

// RouteEntry represents a discovered route to validate.
type RouteEntry struct {
	Path   string
	Method string // GET, POST etc.
	Meta   RouteEntryMeta
}

type Conflict struct {
	A      RouteEntry
	B      RouteEntry
	Reason string
}

// FindConflicts returns deterministic list of route conflicts.
// Methods are respected: only endpoints with the same method can conflict.
func FindConflicts(entries []RouteEntry) []Conflict {
	root := newTrieNode()
	var conflicts []Conflict
	seen := map[string]bool{}

	for i := range entries {
		entry := entries[i]
		normPath := normalizePath(entry.Path)
		newSegments := splitSegments(normPath)

		curr := root
		for idx, seg := range newSegments {
			if isParamSegment(seg) {
				reportParamVsLiterals(&conflicts, seen, curr, entry, newSegments, seg)
				reportParamVsParam(&conflicts, seen, curr, entry, newSegments, idx, seg)

				if curr.paramChild == nil {
					curr.paramChild = newTrieNode()
				}
				curr = curr.paramChild
				continue
			}

			// literal segment
			reportLiteralVsParam(&conflicts, seen, curr, entry, newSegments, idx, seg)
			next := curr.literalChildren[seg]
			if next == nil {
				next = newTrieNode()
				curr.literalChildren[seg] = next
			}
			curr = next
		}

		// endpoint (method-aware)
		existing := curr.endpoint[entry.Method]
		if existing != nil {
			addConflict(&conflicts, seen, entry, *existing, "duplicate method/path combination")
		} else {
			// register endpoint for this method
			curr.endpoint[entry.Method] = &entries[i]
		}
	}

	inPlaceSortConflicts(conflicts)
	return conflicts
}

func normalizePath(p string) string {
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	for strings.Contains(p, "//") {
		p = strings.ReplaceAll(p, "//", "/")
	}
	if len(p) > 1 && strings.HasSuffix(p, "/") {
		p = strings.TrimRight(p, "/")
	}
	return p
}

func splitSegments(p string) []string {
	if p == "/" {
		return []string{}
	}
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		return []string{}
	}
	return strings.Split(p, "/")
}

func isParamSegment(seg string) bool {
	return strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}")
}

func trimParamName(seg string) string {
	return strings.Trim(seg, "{}")
}

// patternsConflict returns true iff two templates can match the same concrete path.
// Requires same number of segments; at each position either equal literals or at
// least one parameter segment.
func patternsConflict(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		aSeg, bSeg := a[i], b[i]
		if aSeg == bSeg {
			continue
		}
		if isParamSegment(aSeg) || isParamSegment(bSeg) {
			continue
		}
		// different literals -> cannot match same concrete path
		return false
	}
	return true
}

// collectEndpointsByMethod traverses the subtree rooted at n and returns
// registered endpoints that have the same HTTP method as requested.
func collectEndpointsByMethod(n *trieNode, method string) []*RouteEntry {
	var out []*RouteEntry
	var dfs func(*trieNode)
	dfs = func(cur *trieNode) {
		if cur == nil {
			return
		}
		if cur.endpoint != nil {
			if ep := cur.endpoint[method]; ep != nil {
				out = append(out, ep)
			}
		}
		for _, child := range cur.literalChildren {
			dfs(child)
		}
		if cur.paramChild != nil {
			dfs(cur.paramChild)
		}
	}
	dfs(n)
	return out
}

func reportParamVsLiterals(
	conflicts *[]Conflict,
	seen map[string]bool,
	curr *trieNode,
	entry RouteEntry,
	newSegments []string,
	seg string,
) {

	if curr == nil {
		return
	}

	for lit, litNode := range curr.literalChildren {
		for _, ep := range collectEndpointsByMethod(litNode, entry.Method) {
			epSegments := splitSegments(normalizePath(ep.Path))
			if !patternsConflict(newSegments, epSegments) {
				continue
			}
			reason := fmt.Sprintf(
				"parameter %q in path '%s %s' conflicts with literal %q in path '%s %s'",
				seg,
				entry.Method,
				entry.Path,
				lit,
				ep.Method,
				ep.Path,
			)
			addConflict(conflicts, seen, entry, *ep, reason)
		}
	}
}

func reportParamVsParam(
	conflicts *[]Conflict,
	seen map[string]bool,
	curr *trieNode,
	entry RouteEntry,
	newSegments []string,
	idx int,
	seg string,
) {

	if curr == nil || curr.paramChild == nil {
		return
	}
	for _, ep := range collectEndpointsByMethod(curr.paramChild, entry.Method) {
		epSegments := splitSegments(normalizePath(ep.Path))
		if !patternsConflict(newSegments, epSegments) {
			continue
		}
		reason := fmt.Sprintf(
			"parameter %q in path '%s %s' conflicts with parameter %q in path '%s %s'",
			seg,
			entry.Method,
			entry.Path,
			epSegments[idx],
			ep.Method,
			ep.Path,
		)
		addConflict(conflicts, seen, entry, *ep, reason)
	}
}

func reportLiteralVsParam(
	conflicts *[]Conflict,
	seen map[string]bool,
	curr *trieNode,
	entry RouteEntry,
	newSegments []string,
	idx int,
	seg string,
) {

	if curr == nil || curr.paramChild == nil {
		return
	}
	for _, ep := range collectEndpointsByMethod(curr.paramChild, entry.Method) {
		epSegments := splitSegments(normalizePath(ep.Path))
		if !patternsConflict(newSegments, epSegments) {
			continue
		}
		reason := fmt.Sprintf(
			"literal %q in path \"%s %s\" conflicts with parameter %q of in path '%s %s'",
			seg,
			entry.Method,
			entry.Path,
			epSegments[idx],
			ep.Method,
			ep.Path,
		)
		addConflict(conflicts, seen, entry, *ep, reason)
	}
}

func addConflict(out *[]Conflict, seen map[string]bool, a RouteEntry, b RouteEntry, reason string) {
	aPath, bPath := a.Path, b.Path
	// canonical order
	if aPath > bPath {
		aPath, bPath = bPath, aPath
		a, b = b, a
	}
	key := aPath + "||" + bPath + "||" + reason
	if seen[key] {
		return
	}
	seen[key] = true
	*out = append(*out, Conflict{A: a, B: b, Reason: reason})
}

func inPlaceSortConflicts(conflicts []Conflict) []Conflict {
	slices.SortStableFunc(conflicts, func(a, b Conflict) int {
		if a.A.Path == b.A.Path {
			if a.B.Path == b.B.Path {
				if a.Reason == b.Reason {
					return 0
				}
				if a.Reason < b.Reason {
					return -1
				}
				return 1
			}
			if a.B.Path < b.B.Path {
				return -1
			}
			return 1
		}

		if a.A.Path < b.A.Path {
			return -1
		}

		return 1
	})

	return conflicts
}
