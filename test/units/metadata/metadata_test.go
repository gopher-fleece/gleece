package metadata_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/core/visitors"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	ctx visitors.VisitContext
)

var _ = BeforeSuite(func() {
	ctx = utils.GetVisitContextByRelativeConfigOrFail("gleece.test.config.json")
})

func TestUnitMetadata(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - Metadata")
}
