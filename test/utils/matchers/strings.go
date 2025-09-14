package matchers

import (
	"strings"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

// ContainSubstringFromRight is the same as ContainSubstring but optimized for rFind instead of lFind
func ContainSubstringFromRight(str string) types.GomegaMatcher {
	return WithTransform(
		func(s string) int {
			return strings.LastIndex(s, str)
		},
		BeNumerically(">=", 0),
	)
}
