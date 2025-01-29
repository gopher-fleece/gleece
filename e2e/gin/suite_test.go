package e2e

import (
	"testing"

	"github.com/gopher-fleece/gleece/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGleeceGinE2E(t *testing.T) {
	// Disable logging to reduce clutter.
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gleece Gin E2E Suite")
}

var _ = BeforeEach(func() {
})
