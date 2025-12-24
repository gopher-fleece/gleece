package gast_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - Logger", func() {
	BeforeEach(func() {
		logger.SetLogLevel(logger.LogLevelNone)
	})

	It("Sets and get log level", func() {
		logger.SetLogLevel(logger.LogLevelDebug)
		Expect(logger.GetLogLevel()).To(Equal(logger.LogLevelDebug))
		logger.SetLogLevel(logger.LogLevelInfo)
		Expect(logger.GetLogLevel()).To(Equal(logger.LogLevelInfo))
		logger.SetLogLevel(logger.LogLevelWarn)
		Expect(logger.GetLogLevel()).To(Equal(logger.LogLevelWarn))
		logger.SetLogLevel(logger.LogLevelError)
		Expect(logger.GetLogLevel()).To(Equal(logger.LogLevelError))
		logger.SetLogLevel(logger.LogLevelFatal)
		Expect(logger.GetLogLevel()).To(Equal(logger.LogLevelFatal))
		logger.SetLogLevel(logger.LogLevelNone)
		Expect(logger.GetLogLevel()).To(Equal(logger.LogLevelNone))
	})
})

func TestUnitsLogger(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - Logger")
}
