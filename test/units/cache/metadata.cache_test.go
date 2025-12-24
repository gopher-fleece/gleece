package cache_test

import (
	"go/ast"
	"testing"

	"github.com/gopher-fleece/gleece/v2/common"
	"github.com/gopher-fleece/gleece/v2/core/arbitrators/caching"
	"github.com/gopher-fleece/gleece/v2/core/metadata"
	"github.com/gopher-fleece/gleece/v2/graphs"
	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
	"github.com/gopher-fleece/gleece/v2/test/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MetadataCache", func() {
	var ctx utils.StdTestCtx
	BeforeEach(func() {
		ctx = utils.CreateStdTestCtx("gleece.test.config.json")
	})

	Context("NewMetadataCache", func() {
		It("Successfully creates metadata cache", func() {
			Expect(func() {
				caching.NewMetadataCache()
			}).ToNot(Panic())
		})
	})

	Context("HasController", func() {
		It("Returns nil when no controller is present", func() {
			hasController := caching.NewMetadataCache().HasController(&metadata.ControllerMeta{
				Struct: metadata.StructMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Node:     utils.MakeIdent("F1"),
						FVersion: utils.MakeFileVersion("a", ""),
					},
				},
			})

			Expect(hasController).To(BeFalse())
		})

		It("Returns controller when it exists in the cache", func() {
			for _, file := range ctx.Orc.GetAllSourceFiles() {
				ast.Walk(ctx.Orc, file)
			}

			controllers := ctx.VisitCtx.Graph.FindByKind(common.SymKindController)
			Expect(controllers).To(HaveLen(1))

			meta := controllers[0].Data.(metadata.ControllerMeta)

			hasController := ctx.VisitCtx.MetadataCache.HasController(&meta)
			Expect(hasController).To(BeTrue())
		})
	})

	Context("HasReceiver", func() {
		It("Returns nil when no receiver is present", func() {
			hasReceiver := caching.NewMetadataCache().HasReceiver(
				graphs.NewSymbolKey(utils.MakeIdent("F1"),
					utils.MakeFileVersion("a", ""),
				))

			Expect(hasReceiver).To(BeFalse())
		})

		It("Returns receiver when it exists in the cache", func() {
			for _, file := range ctx.Orc.GetAllSourceFiles() {
				ast.Walk(ctx.Orc, file)
			}

			structType := ctx.VisitCtx.Graph.FindByKind(common.SymKindStruct)
			Expect(structType).To(HaveLen(1))

			hasStruct := ctx.VisitCtx.MetadataCache.HasStruct(structType[0].Id)
			Expect(hasStruct).To(BeTrue())
		})
	})

	Context("HasStruct", func() {
		It("Returns nil when no struct is present", func() {
			hasStruct := caching.NewMetadataCache().HasStruct(
				graphs.NewSymbolKey(utils.MakeIdent("F1"),
					utils.MakeFileVersion("a", ""),
				))

			Expect(hasStruct).To(BeFalse())
		})

		It("Returns struct when it exists in the cache", func() {
			for _, file := range ctx.Orc.GetAllSourceFiles() {
				ast.Walk(ctx.Orc, file)
			}

			receivers := ctx.VisitCtx.Graph.FindByKind(common.SymKindStruct)
			Expect(receivers).To(HaveLen(1))

			hasRoute := ctx.VisitCtx.MetadataCache.HasStruct(receivers[0].Id)
			Expect(hasRoute).To(BeTrue())
		})
	})

	Context("HasEnum", func() {
		It("Returns nil when no struct is present", func() {
			hasStruct := caching.NewMetadataCache().HasStruct(
				graphs.NewSymbolKey(utils.MakeIdent("F1"),
					utils.MakeFileVersion("a", ""),
				))

			Expect(hasStruct).To(BeFalse())
		})

		It("Returns enum when it exists in the cache", func() {
			for _, file := range ctx.Orc.GetAllSourceFiles() {
				ast.Walk(ctx.Orc, file)
			}

			enums := ctx.VisitCtx.Graph.FindByKind(common.SymKindEnum)
			Expect(enums).To(HaveLen(1))

			hasEnum := ctx.VisitCtx.MetadataCache.HasEnum(enums[0].Id)
			Expect(hasEnum).To(BeTrue())
		})
	})

	Context("GetFileVersion", func() {
		It("Returns an error when unable to construct a FileVersion", func() {
			_, err := ctx.VisitCtx.MetadataCache.GetFileVersion(&ast.File{}, nil)
			Expect(err).To(MatchError(ContainSubstring("GetFileFullPath was provided nil file or fileSet")))
		})
	})

	Context("AddEnum", func() {
		It("Returns an error when attempting to add the same entity multiple times", func() {
			fVersion := utils.MakeFileVersion("a", "")

			enumMeta := metadata.EnumMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Name:     "MyEnum",
					Node:     utils.MakeIdent("MyEnum"),
					FVersion: fVersion,
				},
				ValueKind: metadata.EnumValueKindString,
				Values: []metadata.EnumValueDefinition{
					{
						SymNodeMeta: metadata.SymNodeMeta{
							Name:     "ValueOne",
							Node:     utils.MakeIdent("ValueOne"),
							FVersion: fVersion,
						},
						Value: "First",
					},
					{
						SymNodeMeta: metadata.SymNodeMeta{
							Name:     "ValueTwo",
							Node:     utils.MakeIdent("ValueTwo"),
							FVersion: fVersion,
						},
						Value: "Second",
					},
				},
			}

			err := ctx.VisitCtx.MetadataCache.AddEnum(&enumMeta)
			Expect(err).To(BeNil())

			err = ctx.VisitCtx.MetadataCache.AddEnum(&enumMeta)
			Expect(err).To(MatchError(ContainSubstring("already exists in cache")))
		})
	})
})

func TestMetadataCacheVisitor(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "MetadataCache")
}
