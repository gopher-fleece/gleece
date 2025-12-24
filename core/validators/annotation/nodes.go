package annotation

import (
	"regexp"

	"github.com/gopher-fleece/gleece/v2/common"
	"github.com/gopher-fleece/gleece/v2/core/annotations"
)

var routeParamRe = regexp.MustCompile(`\{([\w\d-_]+)\}`)

// small structs
type routeParam struct {
	Name  string
	Range common.ResolvedRange
}

type routeNode struct {
	Raw    string
	Range  common.ResolvedRange
	Params []routeParam
}

type pathNode struct {
	Attr  *annotations.Attribute
	Value string // function param name
	Alias string // route alias (name property)
	Range common.ResolvedRange
	File  string
}
