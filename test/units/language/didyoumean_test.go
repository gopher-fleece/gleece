package language_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/v2/common/language"
	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Quick tests for the untested parts of the resident copy of 'didyoumean' package
// (https://github.com/sc0Vu/didyoumean).
//
// Could import the lib but preferred not to, as its a very small/low traffic one.
// Copy of license is attached on the implementation elsewhere in this project.
var _ = Describe("Unit Tests - DidYouMean", func() {
	BeforeEach(func() {
		language.CaseInsensitive = false
		language.ThresholdRate = 0
	})

	It("Returns an empty string when given an empty key", func() {
		Expect(language.DidYouMean("", []string{"abcdef", "ghijk"})).To(Equal(""))
	})

	It("Considers threshold", func() {
		// Use a high threshold with case sensitivity enabled.
		// This should fail the lookup.
		language.ThresholdRate = 0.7
		Expect(language.DidYouMean("BCDEF", []string{"abcdef", "ghijk"})).To(Equal(""))
	})

	It("Considers threshold", func() {
		// Now we also add case insensitivity - search should now return a recommendation
		language.ThresholdRate = 0.7
		language.CaseInsensitive = true
		Expect(language.DidYouMean("BCDEF", []string{"abcdef", "ghijk"})).To(Equal("abcdef"))
	})

	It("Returns first option when give non-UTF8 strings regardless of match", func() {
		// This is kind of a weird behavior, present in the original code.
		// Just adding a test to keep coverage at 100%.
		a := string([]byte{0x80})
		b := string([]byte{0xFF})

		Expect(language.DidYouMean(a, []string{b})).To(Equal(b))
	})

})

func TestUnitDidYouMean(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - DidYouMean")
}
