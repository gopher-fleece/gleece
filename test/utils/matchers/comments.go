package matchers

import (
	"github.com/gopher-fleece/gleece/v2/gast"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func BeAnEmptyCommentBlock() types.GomegaMatcher {
	// The tooling returns an empty slice for sanity's sake, not a nil so comparison's a bit awkward
	compared := gast.CommentBlock{
		Comments: []gast.CommentNode{},
	}

	return Equal(compared)
}
