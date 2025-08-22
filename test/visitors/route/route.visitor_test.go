package route_test

import (
	"fmt"
	"go/ast"
	"os"
	"testing"
	"time"

	"github.com/gopher-fleece/gleece/core/arbitrators/caching"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/visitors"
	"github.com/gopher-fleece/gleece/core/visitors/providers"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/gleece/test/utils"
	"github.com/titanous/json5"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const controllerFileRelPath = "./resources/micro.valid.controller.go"
const controllerName = "RouteVisitorTestController"
const receiver1Name = "Receiver1"
const receiver2Name = "Receiver2"
const receiver3Name = "Receiver3"

type TestCtx struct {
	controllerAstFile *ast.File
	receiver1Decl     *ast.FuncDecl
	receiver2Decl     *ast.FuncDecl
	receiver3Decl     *ast.FuncDecl

	arbProvider  *providers.ArbitrationProvider
	metaCache    *caching.MetadataCache
	symGraph     symboldg.SymbolGraph
	visitCtx     *visitors.VisitContext
	parentCtx    visitors.RouteParentContext
	routeVisitor *visitors.RouteVisitor
}

var _ = Describe("RouteVisitor", func() {
	var ctx TestCtx
	BeforeEach(func() {
		ctx = createTestCtx([]string{controllerFileRelPath})
	})

	Context("NewRouteVisitor", func() {
		It("Returns an error if initialization with arbitration provider fails", func() {
			_, err := visitors.NewRouteVisitor(nil, ctx.parentCtx)
			Expect(err).To(MatchError(ContainSubstring("nil context was given to contextInitGuard")))
		})

		It("Returns an error if nested TypeVisitor initialization fails", func() {
			ctx.visitCtx.ArbitrationProvider = nil
			ctx.visitCtx.GleeceConfig = &definitions.GleeceConfig{
				CommonConfig: definitions.CommonConfig{
					ControllerGlobs: []string{"./non.go.file.txt"},
				},
			}
			_, err := visitors.NewRouteVisitor(ctx.visitCtx, ctx.parentCtx)
			Expect(err).To(MatchError(ContainSubstring("failed to parse file")))
		})
	})

	Context("VisitMethod", func() {

		It("Returns nil when function has no comments (not an API endpoint)", func() {
			noDocFn := &ast.FuncDecl{
				Name: ast.NewIdent("NoDocFn"),
				Type: &ast.FuncType{
					Params:  &ast.FieldList{},
					Results: &ast.FieldList{},
				},
				Doc: nil,
			}

			meta, err := ctx.routeVisitor.VisitMethod(noDocFn, ctx.controllerAstFile)
			Expect(err).To(BeNil())
			Expect(meta).To(BeNil())
		})

		It("Returns nil when annotations are present but missing @Method (not an API endpoint)", func() {
			doc := &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// @Route(/) Has a route but not a @Method"},
				},
			}
			fn := &ast.FuncDecl{
				Name: ast.NewIdent("HasRouteNoMethod"),
				Doc:  doc,
				Type: &ast.FuncType{
					Params:  &ast.FieldList{},
					Results: &ast.FieldList{},
				},
			}

			meta, err := ctx.routeVisitor.VisitMethod(fn, ctx.controllerAstFile)
			Expect(err).To(BeNil())
			Expect(meta).To(BeNil())
		})

		It("Returns nil when annotations are present but missing @Route (not an API endpoint)", func() {
			doc := &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// @Method(GET) This controller method exists but has no route"},
				},
			}
			fn := &ast.FuncDecl{
				Name: ast.NewIdent("HasMethodNoRoute"),
				Doc:  doc,
				Type: &ast.FuncType{
					Params:  &ast.FieldList{},
					Results: &ast.FieldList{},
				},
			}

			meta, err := ctx.routeVisitor.VisitMethod(fn, ctx.controllerAstFile)
			Expect(err).To(BeNil())
			Expect(meta).To(BeNil())
		})

		It("Returns a SyntaxError when annotations include a malformed JSON5", func() {
			doc := &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// @Method(POST)"},
					{Text: "// @Route(/x, {1invalidJson : })"},
				},
			}
			fn := &ast.FuncDecl{
				Name: ast.NewIdent("HasBadJsonAnnotation"),
				Doc:  doc,
				Type: &ast.FuncType{
					Params:  &ast.FieldList{},
					Results: &ast.FieldList{},
				},
			}

			meta, err := ctx.routeVisitor.VisitMethod(fn, ctx.controllerAstFile)
			Expect(meta).To(BeNil())
			Expect(err).ToNot(BeNil())

			// Assert error type and message details
			Expect(err).To(BeAssignableToTypeOf(&json5.SyntaxError{}))
			Expect(err.Error()).To(ContainSubstring("invalid character '1' looking for beginning of object key"))
		})

		It("Returns an error when package for provided AST file cannot be obtained", func() {
			doc := &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// @Method(GET)"},
					{Text: "// @Route(/unregistered)"},
				},
			}
			fn := &ast.FuncDecl{
				Name: ast.NewIdent("UnregisteredFileFn"),
				Doc:  doc,
				Type: &ast.FuncType{
					Params:  &ast.FieldList{},
					Results: &ast.FieldList{},
				},
			}

			// Use a new AST file (not registered in provider)
			unregisteredFile := &ast.File{Name: ast.NewIdent("not_registered_file")}
			meta, err := ctx.routeVisitor.VisitMethod(fn, unregisteredFile)
			Expect(meta).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("could not obtain package object for file"))
		})

		It("Successfully returns ReceiverMeta for a real controller method parsed by the provider", func() {
			rcvMeta, err := ctx.routeVisitor.VisitMethod(ctx.receiver1Decl, ctx.controllerAstFile)
			Expect(err).To(BeNil())
			Expect(rcvMeta).ToNot(BeNil())

			// Basic sanity checks on metadata
			Expect(rcvMeta.Name).To(Equal(receiver1Name))
			Expect(rcvMeta.Annotations).ToNot(BeNil())
		})

		It("Returns cached ReceiverMeta on subsequent calls (cache hit)", func() {
			// First call populates cache
			firstMeta, firstErr := ctx.routeVisitor.VisitMethod(ctx.receiver1Decl, ctx.controllerAstFile)
			Expect(firstErr).To(BeNil())
			Expect(firstMeta).ToNot(BeNil())

			// Second call should return identical pointer from cache
			secondMeta, secondErr := ctx.routeVisitor.VisitMethod(ctx.receiver1Decl, ctx.controllerAstFile)
			Expect(secondErr).To(BeNil())
			Expect(secondMeta).ToNot(BeNil())
			Expect(secondMeta).To(BeIdenticalTo(firstMeta))
		})

		It("Returns an error when unable to determine package name for source file hosting a given function", func() {
			// Corrupt the AST file struct to trigger a downstream failure in getPkgForSourceFile creation
			ctx.controllerAstFile.Package = -1
			ctx.controllerAstFile.Name.Name = ""
			_, err := ctx.routeVisitor.VisitMethod(ctx.receiver1Decl, ctx.controllerAstFile)
			Expect(err).To(MatchError(ContainSubstring("could not obtain package object for file 'UNNAMED_FILE'")))
		})

		It("Returns an error when hash calculation fails on a given AST file", func() {
			const tempDir = "./temp"

			// First, cleanup any remains from previous tests, if necessary
			utils.DeleteRelativeFolderOrFail(tempDir)

			// Next, copy the original controller file to a temp file - we'll be locking it and we don't want
			// to potentially affect other tests
			originalControllerFile := utils.ReadFileByRelativePathOrFail(controllerFileRelPath)
			tempFile := "./temp/locked.controller.go"
			utils.WriteFileByRelativePathOrFail(tempFile, []byte(originalControllerFile))

			// Defer a cleanup for the copy
			defer utils.DeleteRelativeFolderOrFail(tempDir)

			// Create a new context for this specific test
			thisCtx := createTestCtx([]string{tempFile})

			// Chmod the temp file so FileVersion's hash calculation fails downstream
			err := os.Chmod(tempFile, 0)
			Expect(err).To(BeNil())
			// Restore permissions for cleanup
			defer os.Chmod(tempFile, 0644)

			// Call VisitMethod, should hit os.Stat failure inside FileVersion creation
			_, err = thisCtx.routeVisitor.VisitMethod(thisCtx.receiver1Decl, thisCtx.controllerAstFile)
			Expect(err).ToNot(BeNil())
			Expect(err).To(MatchError(ContainSubstring("failed to compute hash for file")))
		})

		It("Returns a proper error if a receiver parameter has ann invalid type", func() {
			_, err := ctx.routeVisitor.VisitMethod(ctx.receiver2Decl, ctx.controllerAstFile)
			Expect(err).To(MatchError(ContainSubstring(
				"could not create type usage metadata for field paramWithInvalidType - " +
					"failed to build type layers for expression with type name 'string' - " +
					"unsupported type expression: *ast.ChanType",
			)))
		})

		It("Returns a proper error if a receiver return value has ann invalid type", func() {
			_, err := ctx.routeVisitor.VisitMethod(ctx.receiver3Decl, ctx.controllerAstFile)
			Expect(err).To(MatchError(ContainSubstring(
				"could not create type usage metadata for field string - " +
					"failed to build type layers for expression with type name 'string' - " +
					"unsupported type expression: *ast.ChanType",
			)))
		})

		It("Returns a proper error if a receiver return value has ann invalid type", func() {
			_, err := ctx.routeVisitor.VisitMethod(ctx.receiver3Decl, ctx.controllerAstFile)
			Expect(err).To(MatchError(ContainSubstring(
				"could not create type usage metadata for field string - " +
					"failed to build type layers for expression with type name 'string' - " +
					"unsupported type expression: *ast.ChanType",
			)))
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

	// Get the controller's source file so we can use it directly with the visitor
	for _, f := range srcFiles {
		for _, decl := range f.Decls {
			if fd, ok := decl.(*ast.FuncDecl); ok {
				switch fd.Name.Name {
				case receiver1Name:
					ctx.controllerAstFile = f
					ctx.receiver1Decl = fd
				case receiver2Name:
					ctx.receiver2Decl = fd
				case receiver3Name:
					ctx.receiver3Decl = fd
				}
			}
		}
		if ctx.controllerAstFile != nil {
			break
		}
	}
	Expect(ctx.controllerAstFile).ToNot(BeNil(), fmt.Sprintf("Expected to find %s in parsed AST", receiver1Name))
	Expect(ctx.receiver1Decl).ToNot(BeNil(), fmt.Sprintf("Expected to find %s func in parsed AST", receiver1Name))

	// Build the VisitContext and routeVisitor as before using arbProvider
	ctx.metaCache = caching.NewMetadataCache()
	ctx.symGraph = symboldg.NewSymbolGraph()
	ctx.visitCtx = &visitors.VisitContext{
		ArbitrationProvider: arbProvider,
		MetadataCache:       ctx.metaCache,
		GraphBuilder:        &ctx.symGraph,
	}

	ctx.parentCtx = visitors.RouteParentContext{
		Controller: &metadata.ControllerMeta{
			Struct: metadata.StructMeta{
				SymNodeMeta: metadata.SymNodeMeta{
					Node:     &ast.TypeSpec{Name: ast.NewIdent(controllerName)},
					FVersion: &gast.FileVersion{Path: "p", ModTime: time.Now(), Hash: "h"},
				},
			},
		},
	}

	ctx.routeVisitor, err = visitors.NewRouteVisitor(ctx.visitCtx, ctx.parentCtx)
	Expect(err).To(BeNil())
	Expect(ctx.routeVisitor).ToNot(BeNil())

	return ctx
}

func TestRouteVisitor(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "RouteVisitor")
}
