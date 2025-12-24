package swagen30

import (
	"testing"

	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSwagen30Module(t *testing.T) {
	// Disable logging to reduce clutter.
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Swagen v3.0 Module Suite")
}

var _ = BeforeEach(func() {
	// Clear the schemaRefMap before each test
	schemaRefMap = []SchemaRefMap{}
})
