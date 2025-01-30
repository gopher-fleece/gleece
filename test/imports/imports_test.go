package imports_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopher-fleece/gleece/cmd"
	"github.com/gopher-fleece/gleece/cmd/arguments"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var metadata []definitions.ControllerMetadata
var models []definitions.ModelMetadata

var _ = BeforeSuite(func() {
	cwd, err := os.Getwd()
	if err != nil {
		Fail(fmt.Sprintf("Could not determine process working directory - %v", err))
	}

	configPath := filepath.Join(cwd, "imports.config.json")
	_, controllers, flatModels, _, err := cmd.GetConfigAndMetadata(arguments.CliArguments{ConfigPath: configPath})
	if err != nil {
		Fail(fmt.Sprintf("Could not generate routes - %v", err))
	}

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
		Expect(route.FuncParams[0].TypeMeta.EntityKind).To(Equal(definitions.AstNodeKindStruct))

		Expect(route.Responses).To(HaveLen(2))
		Expect(route.Responses[0].TypeMetadata.Name).To(Equal("ImportedWithDot"))
		Expect(route.Responses[0].TypeMetadata.FullyQualifiedPackage).To(Equal("github.com/gopher-fleece/gleece/test/types"))
		Expect(route.Responses[0].TypeMetadata.DefaultPackageAlias).To(Equal("types"))
		Expect(route.Responses[0].TypeMetadata.Import).To(Equal(definitions.ImportTypeDot))
		Expect(route.Responses[0].TypeMetadata.IsUniverseType).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.IsByAddress).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.EntityKind).To(Equal(definitions.AstNodeKindStruct))
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
		Expect(route.FuncParams[0].TypeMeta.EntityKind).To(Equal(definitions.AstNodeKindStruct))

		Expect(route.Responses).To(HaveLen(2))
		Expect(route.Responses[0].TypeMetadata.Name).To(Equal("ImportedWithDefaultAlias"))
		Expect(route.Responses[0].TypeMetadata.FullyQualifiedPackage).To(Equal("github.com/gopher-fleece/gleece/test/types"))
		Expect(route.Responses[0].TypeMetadata.DefaultPackageAlias).To(Equal("types"))
		Expect(route.Responses[0].TypeMetadata.Import).To(Equal(definitions.ImportTypeAlias))
		Expect(route.Responses[0].TypeMetadata.IsUniverseType).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.IsByAddress).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.EntityKind).To(Equal(definitions.AstNodeKindStruct))
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
		Expect(route.FuncParams[0].TypeMeta.EntityKind).To(Equal(definitions.AstNodeKindStruct))

		Expect(route.Responses).To(HaveLen(2))
		Expect(route.Responses[0].TypeMetadata.Name).To(Equal("ImportedWithCustomAlias"))
		Expect(route.Responses[0].TypeMetadata.FullyQualifiedPackage).To(Equal("github.com/gopher-fleece/gleece/test/types"))
		Expect(route.Responses[0].TypeMetadata.DefaultPackageAlias).To(Equal("types"))
		Expect(route.Responses[0].TypeMetadata.Import).To(Equal(definitions.ImportTypeAlias))
		Expect(route.Responses[0].TypeMetadata.IsUniverseType).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.IsByAddress).To(BeFalse())
		Expect(route.Responses[0].TypeMetadata.EntityKind).To(Equal(definitions.AstNodeKindStruct))
	})
})

func TestSanityController(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Imports Controller")
}
