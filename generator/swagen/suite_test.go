package swagen

import (
	"testing"

	"github.com/gopher-fleece/gleece/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSwagenModule(t *testing.T) {
	// Disable logging to reduce clutter.
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Swagen Module Suite")
}

var _ = BeforeEach(func() {
	// Clear the schemaRefMap before each test
	schemaRefMap = []SchemaRefMap{}
})
