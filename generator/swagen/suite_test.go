package swagen

import (
	"testing"

	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSwagenManagerModule(t *testing.T) {
	// Disable logging to reduce clutter.
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Swagen Manager Suite")
}

var _ = BeforeEach(func() {
})
