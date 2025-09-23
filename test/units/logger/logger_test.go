package gast_test

import (
	"bytes"
	"log"
	"testing"

	"github.com/gopher-fleece/gleece/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - Logger", func() {
	var buf *bytes.Buffer

	BeforeEach(func() {
		buf = new(bytes.Buffer)
		log.SetOutput(buf) // Redirect log output to capture it
	})

	AfterEach(func() {
		log.SetOutput(GinkgoWriter) // Restore default output
	})

	Context("Logging behavior", func() {
		BeforeEach(func() {
			logger.SetLogLevel(logger.LogLevelInfo)
			buf.Reset()
		})

		It("Should always print SYSTEM messages", func() {
			logger.SetLogLevel(logger.LogLevelNone)
			logger.System("System test")
			Expect(buf.String()).To(ContainSubstring("[SYSTEM] System test"))
		})

		It("Should log messages at or above the set level", func() {
			logger.Info("This is an info message")
			Expect(buf.String()).To(ContainSubstring("[INFO]   This is an info message"))
			buf.Reset()

			logger.Warn("This is a warning")
			Expect(buf.String()).To(ContainSubstring("[WARN]   This is a warning"))
			buf.Reset()

			logger.Error("This is an error")
			Expect(buf.String()).To(ContainSubstring("[ERROR]  This is an error"))
		})

		It("Should not log messages below the set level", func() {
			logger.Debug("This debug message should not appear")
			Expect(buf.String()).NotTo(ContainSubstring("[DEBUG]"))
		})

		It("Should log fatal messages even at high log levels", func() {
			logger.SetLogLevel(logger.LogLevelFatal)
			buf.Reset()
			logger.Fatal("This is a fatal error")
			Expect(buf.String()).To(ContainSubstring("[FATAL]  This is a fatal error"))
		})
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
