package dot

import (
	"fmt"
	"strings"

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
}

func NewDotBuilder(theme *DotTheme) *DotBuilder {
	themeToUse := DefaultDotTheme
	if theme != nil {
		themeToUse = *theme
	}
	db := &DotBuilder{
		theme: themeToUse,
		idMap: make(map[graphs.SymbolKey]string),
	}
	db.sb.WriteString("digraph SymbolGraph {\n")
	db.sb.WriteString(fmt.Sprintf("  rankdir=%s;\n", themeToUse.Direction))
	return db
}
func (db *DotBuilder) AddNode(key graphs.SymbolKey, kind common.SymKind, label string) {
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
		id, label, kind, style.Shape, style.Color, style.FontColor))
}

func (db *DotBuilder) AddEdge(from, to graphs.SymbolKey, kind string) {
	fromID := db.getId(from)

	toID, ok := db.idMap[to]
	if !ok {
		db.addErrorNodeOnce()
		toID = errorNodeId
		kind = "error"
	}

	label := db.theme.EdgeLabels[kind]
	if label == "" {
		label = kind
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
		fromID, toID, label, style.EdgeColor, style.EdgeStyle, style.ArrowHead))
}

func (db *DotBuilder) RenderLegend() {
	db.sb.WriteString("  subgraph cluster_legend {\n")
	db.sb.WriteString("    label = \"Legend\";\n    style = dashed;\n")
	i := 0

	for _, nodeStyle := range db.theme.NodeStylesOrdered() {
		color := nodeStyle.Style.Color
		if color == "" {
			color = "gray90"
		}
		shape := nodeStyle.Style.Shape
		if shape == "" {
			shape = "ellipse"
		}
		db.sb.WriteString(fmt.Sprintf(
			"    L%d [label=\"%s\", style=filled, shape=%s, fillcolor=\"%s\"];\n",
			i, nodeStyle.Kind, shape, color))
		i++
	}
	db.sb.WriteString("  }\n")
}

func (db *DotBuilder) Finish() string {
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
