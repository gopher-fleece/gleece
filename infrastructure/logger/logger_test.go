package logger_test

import (
	"bytes"
	"log"
	"testing"

	"github.com/gopher-fleece/gleece/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logger Tests", func() {
	var buf *bytes.Buffer

	BeforeEach(func() {
		buf = new(bytes.Buffer)
		log.SetOutput(buf) // Redirect log output to capture it
	})

	AfterEach(func() {
		log.SetOutput(GinkgoWriter) // Restore default output
	})

	Context("Setting log levels", func() {
		It("should set the log level correctly", func() {
			logger.SetLogLevel(logger.LogLevelDebug)
			Expect(buf.String()).To(ContainSubstring("[SYSTEM] Verbosity level set to DEBUG"))

			buf.Reset()
			logger.SetLogLevel(logger.LogLevelInfo)
			Expect(buf.String()).To(ContainSubstring("[SYSTEM] Verbosity level set to INFO"))
		})
	})

	Context("Logging behavior", func() {
		BeforeEach(func() {
			logger.SetLogLevel(logger.LogLevelInfo)
			buf.Reset()
		})

		It("should log messages at or above the set level", func() {
			logger.Info("This is an info message")
			Expect(buf.String()).To(ContainSubstring("[INFO]   This is an info message"))
			buf.Reset()

			logger.Warn("This is a warning")
			Expect(buf.String()).To(ContainSubstring("[WARN]   This is a warning"))
			buf.Reset()

			logger.Error("This is an error")
			Expect(buf.String()).To(ContainSubstring("[ERROR]  This is an error"))
		})

		It("should not log messages below the set level", func() {
			logger.Debug("This debug message should not appear")
			Expect(buf.String()).NotTo(ContainSubstring("[DEBUG]"))
		})

		It("should log fatal messages even at high log levels", func() {
			logger.SetLogLevel(logger.LogLevelFatal)
			buf.Reset()
			logger.Fatal("This is a fatal error")
			Expect(buf.String()).To(ContainSubstring("[FATAL]  This is a fatal error"))
		})
	})
})

func TestSanityController(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logger Tests")
}
