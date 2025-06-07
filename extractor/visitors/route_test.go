package visitors

import (
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor/arbitrators"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Route Tests", func() {
	var routeVisitor RouteVisitor

	BeforeEach(func() {
		routeVisitor = RouteVisitor{}
	})

	Context("when validating parameter combinations", func() {
		It("should allow first body parameter", func() {
			// Arrange
			params := []arbitrators.FuncParamWithAst{}

			// Act
			err := routeVisitor.validateParamsCombinations(params, definitions.PassedInBody)

			// Assert
			Expect(err).To(BeNil())
		})

		It("should reject second body parameter", func() {
			// Arrange
			params := []arbitrators.FuncParamWithAst{
				{
					PassedIn: definitions.PassedInBody,
				},
			}

			// Act
			err := routeVisitor.validateParamsCombinations(params, definitions.PassedInBody)

			// Assert
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("only one body per route is allowed"))
		})

		It("should allow first form parameter", func() {
			// Arrange
			params := []arbitrators.FuncParamWithAst{}

			// Act
			err := routeVisitor.validateParamsCombinations(params, definitions.PassedInForm)

			// Assert
			Expect(err).To(BeNil())
		})

		It("should reject body parameter when form parameter exists", func() {
			// Arrange
			params := []arbitrators.FuncParamWithAst{
				{
					PassedIn: definitions.PassedInForm,
				},
			}

			// Act
			err := routeVisitor.validateParamsCombinations(params, definitions.PassedInBody)

			// Assert
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("body parameter is invalid, using body is not allowed when a form is in use"))
		})

		It("should reject form parameter when body parameter exists", func() {
			// Arrange
			params := []arbitrators.FuncParamWithAst{
				{
					PassedIn: definitions.PassedInBody,
				},
			}

			// Act
			err := routeVisitor.validateParamsCombinations(params, definitions.PassedInForm)

			// Assert
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("form parameter is invalid, using form is not allowed when a body is in use"))
		})

		It("should allow multiple form parameters", func() {
			// Arrange
			params := []arbitrators.FuncParamWithAst{
				{
					PassedIn: definitions.PassedInForm,
				},
			}

			// Act
			err := routeVisitor.validateParamsCombinations(params, definitions.PassedInForm)

			// Assert
			Expect(err).To(BeNil())
		})

		It("should allow other parameter types with body", func() {
			// Arrange
			params := []arbitrators.FuncParamWithAst{
				{
					PassedIn: definitions.PassedInBody,
				},
			}

			// Act
			err := routeVisitor.validateParamsCombinations(params, definitions.PassedInQuery)

			// Assert
			Expect(err).To(BeNil())
		})

		It("should allow other parameter types with form", func() {
			// Arrange
			params := []arbitrators.FuncParamWithAst{
				{
					PassedIn: definitions.PassedInForm,
				},
			}

			// Act
			err := routeVisitor.validateParamsCombinations(params, definitions.PassedInHeader)

			// Assert
			Expect(err).To(BeNil())
		})

		It("should allow multiple non-body and non-form parameters", func() {
			// Arrange
			params := []arbitrators.FuncParamWithAst{
				{
					PassedIn: definitions.PassedInQuery,
				},
				{
					PassedIn: definitions.PassedInHeader,
				},
				{
					PassedIn: definitions.PassedInPath,
				},
			}

			// Act
			err := routeVisitor.validateParamsCombinations(params, definitions.PassedInQuery)

			// Assert
			Expect(err).To(BeNil())
		})
	})
})
