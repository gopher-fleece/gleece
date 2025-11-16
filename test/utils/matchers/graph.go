package matchers

import (
	"github.com/gopher-fleece/gleece/common/linq"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func MatchNodeIdNames(expected []string) types.GomegaMatcher {
	return WithTransform(
		func(nodes []*symboldg.SymbolNode) []string {
			return linq.Map(
				nodes,
				func(node *symboldg.SymbolNode) string {
					return node.Id.Name
				},
			)
		},
		ContainElements(expected),
	)
}
