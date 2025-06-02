package imports_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var metadata []definitions.ControllerMetadata
var models []definitions.StructMetadata

var _ = BeforeSuite(func() {
	controllers, flatModels, _ := utils.GetControllersAndModelsOrFail()
	metadata = controllers
	models = flatModels
})

var _ = Describe("Imports Controller", func() {
	It("Dot-imported structs should be properly resolved", func() {
		route := metadata[0].Routes[0]

		Expect(route.FuncParams).To(HaveLen(1))
		Expect(route.FuncParams[0].TypeMeta.Name).To(Equal("ImportedWithDot"))
		Expect(route.FuncParams[0].TypeMeta.FullyQualifiedPackage).To(Equal("github.com/gopher-fleece/gleece/test/types"))
		Expect(route.FuncParams[0].TypeMeta.DefaultPackageAlias).To(Equal("types"))
		Expect(route.FuncParams[0].TypeMeta.Import).To(Equal(definitions.ImportTypeDot))
		Expect(route.FuncParams[0].TypeMeta.IsUniverseType).To(BeFalse())
		Expect(route.FuncParams[0].TypeMeta.IsByAddress).To(BeFalse())
		Expect(route.FuncParams[0].TypeMeta.SymbolKind).To(Equal(definitions.SymKindStruct))

		Expect(route.Responses).To(HaveLen(2))
		Expect(route.Responses[0].TypeMetadata.Name).To(Equal("ImportedWithDot"))
		Expect(route.Responses[0].TypeMetadata.FullyQualifiedPackage).To(Equal("github.com/gopher-fleece/gleece/test/types"))
		Expect(route.Responses[0].TypeMetadata.DefaultPackageAlias).To(Equal("types"))
		Expect(route.Responses[0].TypeMetadata.Import).To(Equal(definitions.ImportTypeDot))
		Expect(route.Responses[0].TypeMetadata.IsUniverseType).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.IsByAddress).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.SymbolKind).To(Equal(definitions.SymKindStruct))
	})

	It("Default-alias-imported structs should be properly resolved", func() {
		route := metadata[0].Routes[1]

		Expect(route.FuncParams).To(HaveLen(1))
		Expect(route.FuncParams[0].TypeMeta.Name).To(Equal("ImportedWithDefaultAlias"))
		Expect(route.FuncParams[0].TypeMeta.FullyQualifiedPackage).To(Equal("github.com/gopher-fleece/gleece/test/types"))
		Expect(route.FuncParams[0].TypeMeta.DefaultPackageAlias).To(Equal("types"))
		Expect(route.FuncParams[0].TypeMeta.Import).To(Equal(definitions.ImportTypeAlias))
		Expect(route.FuncParams[0].TypeMeta.IsUniverseType).To(BeFalse())
		Expect(route.FuncParams[0].TypeMeta.IsByAddress).To(BeFalse())
		Expect(route.FuncParams[0].TypeMeta.SymbolKind).To(Equal(definitions.SymKindStruct))

		Expect(route.Responses).To(HaveLen(2))
		Expect(route.Responses[0].TypeMetadata.Name).To(Equal("ImportedWithDefaultAlias"))
		Expect(route.Responses[0].TypeMetadata.FullyQualifiedPackage).To(Equal("github.com/gopher-fleece/gleece/test/types"))
		Expect(route.Responses[0].TypeMetadata.DefaultPackageAlias).To(Equal("types"))
		Expect(route.Responses[0].TypeMetadata.Import).To(Equal(definitions.ImportTypeAlias))
		Expect(route.Responses[0].TypeMetadata.IsUniverseType).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.IsByAddress).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.SymbolKind).To(Equal(definitions.SymKindStruct))
	})

	It("Custom-alias-imported structs should be properly resolved", func() {
		route := metadata[0].Routes[2]

		Expect(route.FuncParams).To(HaveLen(1))
		Expect(route.FuncParams[0].TypeMeta.Name).To(Equal("ImportedWithCustomAlias"))
		Expect(route.FuncParams[0].TypeMeta.FullyQualifiedPackage).To(Equal("github.com/gopher-fleece/gleece/test/types"))
		Expect(route.FuncParams[0].TypeMeta.DefaultPackageAlias).To(Equal("types"))
		Expect(route.FuncParams[0].TypeMeta.Import).To(Equal(definitions.ImportTypeAlias))
		Expect(route.FuncParams[0].TypeMeta.IsUniverseType).To(BeFalse())
		Expect(route.FuncParams[0].TypeMeta.IsByAddress).To(BeFalse())
		Expect(route.FuncParams[0].TypeMeta.SymbolKind).To(Equal(definitions.SymKindStruct))

		Expect(route.Responses).To(HaveLen(2))
		Expect(route.Responses[0].TypeMetadata.Name).To(Equal("ImportedWithCustomAlias"))
		Expect(route.Responses[0].TypeMetadata.FullyQualifiedPackage).To(Equal("github.com/gopher-fleece/gleece/test/types"))
		Expect(route.Responses[0].TypeMetadata.DefaultPackageAlias).To(Equal("types"))
		Expect(route.Responses[0].TypeMetadata.Import).To(Equal(definitions.ImportTypeAlias))
		Expect(route.Responses[0].TypeMetadata.IsUniverseType).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.IsByAddress).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.SymbolKind).To(Equal(definitions.SymKindStruct))
	})
})

func TestImportsController(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Imports Controller")
}
