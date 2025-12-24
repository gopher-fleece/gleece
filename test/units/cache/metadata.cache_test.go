package cache_test

import (
	"go/ast"
	"testing"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/arbitrators/caching"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/visitors"
	"github.com/gopher-fleece/gleece/core/visitors/providers"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const controllerFileRelPath = "./resources/micro.valid.controller.go"

type TestCtx struct {
	arbProvider       *providers.ArbitrationProvider
	metaCache         *caching.MetadataCache
	symGraph          symboldg.SymbolGraph
	visitCtx          *visitors.VisitContext
	controllerVisitor *visitors.ControllerVisitor
}

var _ = Describe("MetadataCache", func() {
	var ctx TestCtx
	BeforeEach(func() {
		ctx = createTestCtx([]string{controllerFileRelPath})
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
			for _, file := range ctx.controllerVisitor.GetAllSourceFiles() {
				ast.Walk(ctx.controllerVisitor, file)
			}

			controllers := ctx.controllerVisitor.GetControllers()
			Expect(controllers).To(HaveLen(1))

			hasController := ctx.metaCache.HasController(&controllers[0])
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
			for _, file := range ctx.controllerVisitor.GetAllSourceFiles() {
				ast.Walk(ctx.controllerVisitor, file)
			}

			structType := ctx.visitCtx.GraphBuilder.FindByKind(common.SymKindStruct)
			Expect(structType).To(HaveLen(1))

			hasRoute := ctx.metaCache.HasStruct(structType[0].Id)
			Expect(hasRoute).To(BeTrue())
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
			for _, file := range ctx.controllerVisitor.GetAllSourceFiles() {
				ast.Walk(ctx.controllerVisitor, file)
			}

			receivers := ctx.visitCtx.GraphBuilder.FindByKind(common.SymKindStruct)
			Expect(receivers).To(HaveLen(1))

			hasRoute := ctx.metaCache.HasStruct(receivers[0].Id)
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
			for _, file := range ctx.controllerVisitor.GetAllSourceFiles() {
				ast.Walk(ctx.controllerVisitor, file)
			}

			enums := ctx.visitCtx.GraphBuilder.FindByKind(common.SymKindEnum)
			Expect(enums).To(HaveLen(1))

			hasEnum := ctx.metaCache.HasEnum(enums[0].Id)
			Expect(hasEnum).To(BeTrue())
		})
	})

	Context("GetFileVersion", func() {
		It("Returns an error when unable to construct a FileVersion", func() {
			_, err := ctx.metaCache.GetFileVersion(&ast.File{}, nil)
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

			err := ctx.metaCache.AddEnum(&enumMeta)
			Expect(err).To(BeNil())

			err = ctx.metaCache.AddEnum(&enumMeta)
			Expect(err).To(MatchError(ContainSubstring("already exists in cache")))
		})
	})
})

func createTestCtx(fileGlobs []string) TestCtx {
	ctx := TestCtx{}

	// Pass the real controller file so the providers actually load it
	arbProvider, err := providers.NewArbitrationProvider(fileGlobs)
	Expect(err).To(BeNil())
	ctx.arbProvider = arbProvider

	// Verify files were properly loaded
	srcFiles := arbProvider.Pkg().GetAllSourceFiles()
	Expect(srcFiles).ToNot(BeEmpty(), "Arbitration provider parsed zero files; check glob and file contents")

	// Build the VisitContext and routeVisitor as before using arbProvider
	ctx.metaCache = caching.NewMetadataCache()
	ctx.symGraph = symboldg.NewSymbolGraph()
	ctx.visitCtx = &visitors.VisitContext{
		ArbitrationProvider: arbProvider,
		MetadataCache:       ctx.metaCache,
		GraphBuilder:        &ctx.symGraph,
	}

	ctx.controllerVisitor, err = visitors.NewControllerVisitor(ctx.visitCtx)
	Expect(err).To(BeNil(), "Failed to construct a new controller visitor")
	return ctx
}

func TestMetadataCacheVisitor(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "MetadataCache")
}
