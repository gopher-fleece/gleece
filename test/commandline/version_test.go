package commandline_test

import (
	"github.com/gopher-fleece/gleece/v2/cmd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version Command", func() {
	It("Prints expected version information", func() {
		// Note that Version, Build Date, Commit, etc. should be populated during build so they're expected to be empty
		result := cmd.ExecuteWithArgs([]string{"version", "--no-banner"})
		Expect(result.StdOut).To(Equal("Gleece\nVersion: \nBuild Date: \nCommit: \nTarget architecture: \nTarget platform: \n"))

	})
})
