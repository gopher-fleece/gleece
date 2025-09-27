package matchers

import (
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
)

type ComparableNode struct {
	Name       string
	IsUniverse bool
	IsBuiltIn  bool
	Kind       common.SymKind
}

func BeAnEquivalentSymbolTo(node ComparableNode) types.GomegaMatcher {
	return WithTransform(
		transformSymNodeToComparable,
		SatisfyAll(
			gstruct.MatchAllFields(gstruct.Fields{
				"Name":       HavePrefix(node.Name),
				"IsUniverse": Equal(node.IsUniverse),
				"IsBuiltIn":  Equal(node.IsBuiltIn),
				"Kind":       BeEquivalentTo(node.Kind),
			}),
		),
	)
}

func transformSymNodeToComparable(s *symboldg.SymbolNode) ComparableNode {
	return ComparableNode{
		Name:       s.Id.Name,
		IsUniverse: s.Id.IsUniverse,
		IsBuiltIn:  s.Id.IsBuiltIn,
		Kind:       s.Kind,
	}
}
