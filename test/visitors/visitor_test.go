package visitors_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Visitor Tests", func() {

	var _ = Context("Route Visitor", func() {

		It("Ignores receivers with no annotations", func() {
			controllers, _, _ := utils.GetMetadataByRelativeConfigOrFail("./configs/receiver.with.no.comments.json")
			Expect((controllers[0].Routes)).ToNot(ContainElement(Satisfy(func(route definitions.RouteMetadata) bool {
				return route.OperationId == "NotAnApiMethod"
			})))
		})

		It("Returns a correct error when a receiver annotation has invalid JSON", func() {
			_, _, _, err := utils.GetMetadataByRelativeConfig("./configs/receiver.with.invalid.json.json")
			Expect(err).To(MatchError(ContainSubstring("method HasInvalidJson - unexpected end of JSON input")))
		})

		It("Ignores receivers with no @Method annotation", func() {
			controllers, _, _ := utils.GetMetadataByRelativeConfigOrFail("./configs/receiver.with.no.method.annotation.json")
			Expect((controllers[0].Routes)).ToNot(ContainElement(Satisfy(func(route definitions.RouteMetadata) bool {
				return route.OperationId == "NoMethodAnnotation"
			})))
		})

		It("Ignores receivers with no @Route annotation", func() {
			controllers, _, _ := utils.GetMetadataByRelativeConfigOrFail("./configs/receiver.with.no.route.annotation.json")
			Expect((controllers[0].Routes)).ToNot(ContainElement(Satisfy(func(route definitions.RouteMetadata) bool {
				return route.OperationId == "NoRouteAnnotation"
			})))
		})

		Context("Return Type Checks", func() {
			It("Returns a correct error when a receiver returns void", func() {
				_, _, _, err := utils.GetMetadataByRelativeConfig("./configs/receiver.with.void.return.json")
				errStr := err.Error()
				Expect(errStr).To(ContainSubstring("Controller ReceiverWithVoidReturn"))
				Expect(errStr).To(ContainSubstring("Route VoidReturn"))
				Expect(errStr).To(ContainSubstring("expected method to return an error or a value and error tuple but found void"))
			})

			It("Returns a correct error when a receiver has one return type that is not a error", func() {
				_, _, _, err := utils.GetMetadataByRelativeConfig("./configs/receiver.with.non.error.return.json")
				errStr := err.Error()
				Expect(errStr).To(ContainSubstring("Controller ReceiverWithNonErrorReturn"))
				Expect(errStr).To(ContainSubstring("Route NonErrorReturn"))
				Expect(errStr).To(ContainSubstring("return type 'bool' expected to be an error or directly embed it"))

			})
		})
	})

})

func TestVisitors(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Visitor Tests")
}
