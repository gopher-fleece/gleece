package routes

import (
	"testing"

	"github.com/aymerick/raymond"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Engine Registration", func() {
	var config *definitions.GleeceConfig

	BeforeEach(func() {
		// Reset the global state before each test
		lastEngine = nil
		raymond.RemoveAllPartials()

		config = &definitions.GleeceConfig{
			RoutesConfig: definitions.RoutesConfig{
				AuthorizationConfig: definitions.AuthorizationConfig{
					AuthFileFullPackageName: "github.com/example/auth",
				},
			},
		}
	})

	Context("Initial engine registration", func() {
		DescribeTable("should register correct engine",
			func(engineType definitions.RoutingEngineType) {
				config.RoutesConfig.Engine = engineType
				err := registerPartials(config)
				Expect(err).NotTo(HaveOccurred())
				Expect(lastEngine).NotTo(BeNil())
				Expect(*lastEngine).To(Equal(engineType))
			},
			Entry("Gin engine", definitions.RoutingEngineGin),
			Entry("Echo engine", definitions.RoutingEngineEcho),
			Entry("Mux engine", definitions.RoutingEngineMux),
			Entry("Fiber engine", definitions.RoutingEngineFiber),
			Entry("Chi engine", definitions.RoutingEngineChi),
		)

		It("should panic for unknown engine type", func() {
			config.RoutesConfig.Engine = "unknown"
			Expect(func() {
				registerPartials(config)
			}).To(Panic())
		})
	})

	It("should panic for unknown engine type", func() {
		unknownEngine := definitions.RoutingEngineType("unknown")
		Expect(func() {
			getDefaultTemplate(unknownEngine)
		}).To(PanicWith("Could not find an embedded template for routing engine unknown"))
	})

	Context("Template Content Verification", func() {
		DescribeTable("should return non-empty template",
			func(engine definitions.RoutingEngineType) {
				template := getDefaultTemplate(engine)
				Expect(template).NotTo(BeEmpty())
			},
			Entry("Gin template", definitions.RoutingEngineGin),
			Entry("Echo template", definitions.RoutingEngineEcho),
			Entry("Mux template", definitions.RoutingEngineMux),
			Entry("Fiber template", definitions.RoutingEngineFiber),
			Entry("Chi template", definitions.RoutingEngineChi),
		)
	})

	Context("Template Uniqueness", func() {
		It("should return different templates for different engines", func() {
			templates := map[string]string{
				string(definitions.RoutingEngineGin):   getDefaultTemplate(definitions.RoutingEngineGin),
				string(definitions.RoutingEngineEcho):  getDefaultTemplate(definitions.RoutingEngineEcho),
				string(definitions.RoutingEngineMux):   getDefaultTemplate(definitions.RoutingEngineMux),
				string(definitions.RoutingEngineFiber): getDefaultTemplate(definitions.RoutingEngineFiber),
				string(definitions.RoutingEngineChi):   getDefaultTemplate(definitions.RoutingEngineChi),
			}

			// Verify all templates are unique
			uniqueTemplates := make(map[string]bool)
			for _, template := range templates {
				uniqueTemplates[template] = true
			}

			// Each engine should have its own unique template
			Expect(len(uniqueTemplates)).To(Equal(len(templates)))
		})
	})
})

func TestRoutes(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Routes Suite")
}
