package dot

import (
	"fmt"
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/graphs"
)

const errorNodeId = "N_ERROR"

var errorSymbolKey = graphs.SymbolKey{
	Name: errorNodeId,
}

type DotBuilder struct {
	sb      strings.Builder
	theme   DotTheme
	idMap   map[graphs.SymbolKey]string
	counter int

	usedNodeTypes mapset.Set[common.SymKind]
}

func NewDotBuilder(theme *DotTheme) *DotBuilder {
	themeToUse := DefaultDotTheme
	if theme != nil {
		themeToUse = *theme
	}
	db := &DotBuilder{
		theme:         themeToUse,
		idMap:         make(map[graphs.SymbolKey]string),
		usedNodeTypes: mapset.NewSet[common.SymKind](),
	}
	db.sb.WriteString("digraph SymbolGraph {\n")
	db.sb.WriteString(fmt.Sprintf("  rankdir=%s;\n", themeToUse.Direction))
	return db
}

func (db *DotBuilder) LegendEnabled() bool {
	return db.theme.LegendEnabled
}

func (db *DotBuilder) AddNode(key graphs.SymbolKey, kind common.SymKind, label string) {
	db.usedNodeTypes.Add(kind)

	id := db.getId(key)
	style, ok := db.theme.NodeStyles[kind]
	if !ok {
		style = DotStyle{Color: "gray90", Shape: "ellipse"}
	}
	if style.Color == "" {
		style.Color = "gray90"
	}
	if style.Shape == "" {
		style.Shape = "ellipse"
	}
	if style.FontColor == "" {
		style.FontColor = "black"
	}
	db.sb.WriteString(fmt.Sprintf(
		"  %s [label=\"%s (%s)\", shape=%s, style=filled, fillcolor=\"%s\", fontcolor=\"%s\"];\n",
		id,
		label,
		kind,
		style.Shape,
		style.Color,
		style.FontColor,
	))
}

func (db *DotBuilder) AddEdge(
	from, to graphs.SymbolKey,
	kind string,
	suffix *string,
) {
	fromId := db.getId(from)

	toId, ok := db.idMap[to]
	if !ok {
		db.addErrorNodeOnce()
		toId = errorNodeId
		kind = "error"
	}

	label := db.theme.EdgeLabels[kind]
	if label == "" {
		label = kind
	}

	if suffix != nil {
		label = label + *suffix
	}

	style, ok := db.theme.EdgeStyles[kind]
	if !ok {
		style = DotStyle{}
	}
	if style.EdgeColor == "" {
		style.EdgeColor = "black"
	}
	if style.EdgeStyle == "" {
		style.EdgeStyle = "solid"
	}
	if style.ArrowHead == "" {
		style.ArrowHead = "vee"
	}

	db.sb.WriteString(fmt.Sprintf(
		"  %s -> %s [label=\"%s\", color=\"%s\", style=\"%s\", arrowhead=\"%s\"];\n",
		fromId,
		toId,
		label,
		style.EdgeColor,
		style.EdgeStyle,
		style.ArrowHead,
	))
}

func (db *DotBuilder) RenderLegend() {
	if db.usedNodeTypes.IsEmpty() {
		// Empty graph - no need to render legend
		return
	}

	db.sb.WriteString("  subgraph cluster_legend {\n")
	db.sb.WriteString("    label = \"Legend\";\n    style = dashed;\n")
	i := 0

	usedNodeTypes := db.usedNodeTypes.ToSlice()
	slices.Sort(usedNodeTypes)

	for _, kind := range usedNodeTypes {
		nodeStyle := db.theme.NodeStyles[kind]

		color := nodeStyle.Color
		if color == "" {
			color = "gray90"
		}
		shape := nodeStyle.Shape
		if shape == "" {
			shape = "ellipse"
		}
		db.sb.WriteString(fmt.Sprintf(
			"    L%d [label=\"%s\", style=filled, shape=%s, fillcolor=\"%s\"];\n",
			i, kind, shape, color))
		i++
	}
	db.sb.WriteString("  }\n")
}

func (db *DotBuilder) Finish() string {
	if db.LegendEnabled() {
		db.RenderLegend()
	}

	db.sb.WriteString("}\n")
	return db.sb.String()
}

func (db *DotBuilder) getId(key graphs.SymbolKey) string {
	id, ok := db.idMap[key]
	if ok {
		return id
	}
	id = fmt.Sprintf("N%d", db.counter)
	db.counter++
	db.idMap[key] = id
	return id
}

func (db *DotBuilder) addErrorNodeOnce() {
	if _, exists := db.idMap[errorSymbolKey]; exists {
		return
	}

	id := errorNodeId
	db.idMap[errorSymbolKey] = id
	style := db.theme.ErrorNodeStyle
	if style.Color == "" {
		style.Color = "red"
	}
	if style.Shape == "" {
		style.Shape = "octagon"
	}
	if style.FontColor == "" {
		style.FontColor = "white"
	}

	db.sb.WriteString(fmt.Sprintf(
		"  %s [label=\"(unresolved)\", shape=%s, style=filled, fillcolor=\"%s\", fontcolor=\"%s\"];\n",
		id, style.Shape, style.Color, style.FontColor))
}
