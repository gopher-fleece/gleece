package dot

import "github.com/gopher-fleece/gleece/common"

// RankDir represents Graphviz rankdir attribute options
type RankDir string

const (
	RankDirTB RankDir = "TB" // Top to Bottom (default)
	RankDirBT RankDir = "BT" // Bottom to Top
	RankDirLR RankDir = "LR" // Left to Right
	RankDirRL RankDir = "RL" // Right to Left
)

type DotStyle struct {
	Color     string // node fill or edge stroke
	Shape     string // node shape
	FontColor string // node label color
	EdgeColor string // edge color
	EdgeStyle string // edge line style: solid, dashed, etc.
	ArrowHead string // edge arrow type: vee, dot, none
}

type DotTheme struct {
	NodeStyles     map[common.SymKind]DotStyle
	EdgeStyles     map[string]DotStyle
	EdgeLabels     map[string]string
	ErrorNodeStyle DotStyle
	Direction      RankDir
}

var DefaultDotTheme = DotTheme{
	NodeStyles: map[common.SymKind]DotStyle{
		common.SymKindStruct:     {Color: "lightblue", Shape: "box"},
		common.SymKindField:      {Color: "gold", Shape: "ellipse"},
		common.SymKindEnum:       {Color: "mediumpurple", Shape: "folder"},
		common.SymKindEnumValue:  {Color: "plum", Shape: "note"},
		common.SymKindReceiver:   {Color: "orange", Shape: "hexagon"},
		common.SymKindFunction:   {Color: "darkseagreen", Shape: "oval"},
		common.SymKindParameter:  {Color: "khaki", Shape: "parallelogram"},
		common.SymKindReturnType: {Color: "lightgrey", Shape: "diamond"},
		common.SymKindAlias:      {Color: "palegreen", Shape: "note"},
		common.SymKindConstant:   {Color: "plum", Shape: "egg"},
		common.SymKindBuiltin:    {Color: "gray80", Shape: "box"},
		common.SymKindUnknown:    {Color: "lightcoral", Shape: "triangle"},
		common.SymKindInterface:  {Color: "lightskyblue", Shape: "component"},
		common.SymKindPackage:    {Color: "lightyellow", Shape: "folder"},
		common.SymKindController: {Color: "lightcyan", Shape: "octagon"},
		common.SymKindVariable:   {Color: "lightsteelblue", Shape: "circle"},
	},
	EdgeLabels: map[string]string{
		"ty":    "Type",
		"ret":   "Return Value",
		"param": "Parameter",
		"fld":   "Field",
		"recv":  "Receiver",
		"val":   "Value",
		"init":  "Initialize",
		"ref":   "Reference",
	},
	EdgeStyles: map[string]DotStyle{
		"ty":    {EdgeColor: "black", EdgeStyle: "solid", ArrowHead: "vee"},
		"ret":   {EdgeColor: "gray30", EdgeStyle: "dashed", ArrowHead: "normal"},
		"param": {EdgeColor: "gray50", EdgeStyle: "dotted", ArrowHead: "normal"},
		"fld":   {EdgeColor: "gold4", EdgeStyle: "solid", ArrowHead: "vee"},
		"recv":  {EdgeColor: "orange", EdgeStyle: "dashed", ArrowHead: "vee"},
		"ref":   {EdgeColor: "blue", EdgeStyle: "dotted", ArrowHead: "dot"},
		"init":  {EdgeColor: "green4", EdgeStyle: "solid", ArrowHead: "vee"},
		"val":   {EdgeColor: "plum4", EdgeStyle: "solid", ArrowHead: "vee"},
	},
	ErrorNodeStyle: DotStyle{
		Color:     "red",
		Shape:     "septagon",
		FontColor: "white",
	},
	Direction: RankDirTB,
}
