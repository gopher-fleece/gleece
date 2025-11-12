package visitors

import (
	"fmt"
	"go/ast"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	"golang.org/x/tools/go/packages"
)

type RouteParentContext struct {
	Controller *metadata.ControllerMeta
}

type executionContext struct {
	FuncDecl    *ast.FuncDecl
	SourceFile  *ast.File
	FVersion    *gast.FileVersion
	CurrentPkg  *packages.Package
	Annotations *annotations.AnnotationHolder
}

func (c executionContext) IsApiEndpoint() bool {
	if c.Annotations == nil {
		return false
	}
	// Validations are deferred/delegated to after visitation but these here serve to avoid doing unnecessary work.
	methodAttr := c.Annotations.GetFirst(annotations.GleeceAnnotationMethod)
	if methodAttr == nil {
		return false
	}

	routePath := c.Annotations.GetFirstValueOrEmpty(annotations.GleeceAnnotationRoute)
	return len(routePath) > 0
}

type RouteVisitor struct {
	BaseVisitor

	parent RouteParentContext

	fieldVisitor *FieldVisitor
}

func NewRouteVisitor(
	context *VisitContext,
	parent RouteParentContext,
) (*RouteVisitor, error) {
	visitor := RouteVisitor{parent: parent}

	err := visitor.initializeWithArbitrationProvider(context)
	if err != nil {
		return &visitor, err
	}

	return &visitor, err
}

func (v *RouteVisitor) setFieldVisitor(visitor *FieldVisitor) {
	v.fieldVisitor = visitor
}

// visitMethod Visits a controller route given as a FuncDecl and returns its metadata and whether it is an API endpoint
func (v *RouteVisitor) VisitMethod(funcDecl *ast.FuncDecl, sourceFile *ast.File) (*metadata.ReceiverMeta, error) {
	v.enter(fmt.Sprintf("Method '%s'", funcDecl.Name.Name))
	defer v.exit()

	ctx, err := v.getExecutionContext(sourceFile, funcDecl)
	if err != nil || !ctx.IsApiEndpoint() {
		return nil, err
	}

	cached := v.context.MetadataCache.GetReceiver(graphs.NewSymbolKey(ctx.FuncDecl, ctx.FVersion))
	if cached != nil {
		return cached, nil
	}

	metadata, err := v.constructRouteMetadata(ctx)
	return metadata, err
}

func (v *RouteVisitor) getExecutionContext(sourceFile *ast.File, funcDecl *ast.FuncDecl) (executionContext, error) {
	v.enter(fmt.Sprintf("Method '%s' - Initialization", funcDecl.Name.Name))
	defer v.exit()

	ctx := executionContext{
		SourceFile: sourceFile,
		FuncDecl:   funcDecl,
	}

	// Check whether there are any comments on the method - we expect all API endpoints to contain comments.
	// No comments - not an API endpoint.
	if funcDecl.Doc == nil || funcDecl.Doc.List == nil || len(funcDecl.Doc.List) <= 0 {
		return ctx, nil
	}

	comments := gast.MapDocListToCommentBlock(funcDecl.Doc.List, v.context.ArbitrationProvider.Pkg().FSet())
	holder, err := annotations.NewAnnotationHolder(comments, annotations.CommentSourceRoute)
	if err != nil {
		// Couldn't read comments. Fail.
		return ctx, v.frozenError(err)
	}

	ctx.Annotations = &holder

	pkg, err := v.getPkgForSourceFile(sourceFile)
	if err != nil {
		return ctx, v.frozenError(err)
	}

	ctx.CurrentPkg = pkg

	fVersion, err := gast.NewFileVersionFromAstFile(
		sourceFile,
		v.context.ArbitrationProvider.Pkg().FSet(),
	)

	if err != nil {
		return ctx, v.frozenError(err)
	}

	ctx.FVersion = &fVersion

	return ctx, nil
}

func (v *RouteVisitor) getPkgForSourceFile(sourceFile *ast.File) (*packages.Package, error) {
	fileName := gast.GetAstFileNameOrFallback(sourceFile, nil)
	v.enterFmt("Obtaining package for source file %s", fileName)
	defer v.exit()

	pkg, err := v.context.ArbitrationProvider.Pkg().GetPackageForFile(sourceFile)
	if err != nil {
		return nil, v.getFrozenError(
			"could not obtain package object for file '%s' due to error - %v",
			fileName,
			err,
		)
	}

	if pkg == nil {
		return nil, v.getFrozenError("could not find a package object for file '%s'", sourceFile.Name.Name)
	}

	return pkg, nil
}

func (v *RouteVisitor) constructRouteMetadata(ctx executionContext) (*metadata.ReceiverMeta, error) {
	v.enterFmt("Creating route metadata for function %s", ctx.FuncDecl.Name)
	defer v.exit()

	params, err := v.getFuncParams(ctx)
	if err != nil {
		return nil, v.frozenError(err)
	}

	retVals, err := v.getFuncRetVals(ctx)
	if err != nil {
		return nil, v.frozenError(err)
	}

	meta := &metadata.ReceiverMeta{
		SymNodeMeta: metadata.SymNodeMeta{
			Name:        ctx.FuncDecl.Name.Name,
			Node:        ctx.FuncDecl,
			SymbolKind:  common.SymKindReceiver,
			PkgPath:     ctx.CurrentPkg.PkgPath,
			Annotations: ctx.Annotations,
			FVersion:    ctx.FVersion,
			// Range here encapsulates the entire function, from "func" to closing brace
			Range: common.ResolveNodeRange(ctx.CurrentPkg.Fset, ctx.FuncDecl),
		},
		Params:  params,
		RetVals: retVals,
	}

	v.context.MetadataCache.AddReceiver(meta)

	_, err = v.context.Graph.AddRoute(
		symboldg.CreateRouteNode{
			Data: meta,
			ParentController: symboldg.KeyableNodeMeta{
				Decl:     v.parent.Controller.Struct.Node,
				FVersion: *v.parent.Controller.Struct.FVersion,
			},
		},
	)

	if err != nil {
		return nil, v.frozenError(err)
	}

	for _, param := range params {
		v.context.Graph.AddRouteParam(symboldg.CreateParameterNode{
			Data: param,
			ParentRoute: symboldg.KeyableNodeMeta{
				Decl:     meta.Node,
				FVersion: *meta.FVersion,
			},
		})
	}

	for _, retVal := range retVals {
		v.context.Graph.AddRouteRetVal(symboldg.CreateReturnValueNode{
			Data: retVal,
			ParentRoute: symboldg.KeyableNodeMeta{
				Decl:     meta.Node,
				FVersion: *meta.FVersion,
			},
		})
	}

	return meta, nil
}

func (v *RouteVisitor) getFuncParams(ctx executionContext) ([]metadata.FuncParam, error) {
	v.enterFmt("Retrieving params for function %s", ctx.FuncDecl.Name)
	defer v.exit()

	paramTypes, err := v.context.ArbitrationProvider.Ast().GetFuncParametersMeta(
		v.fieldVisitor,
		ctx.CurrentPkg,
		ctx.SourceFile,
		ctx.FuncDecl,
		ctx.Annotations,
	)

	return paramTypes, err
}

func (v *RouteVisitor) getFuncRetVals(ctx executionContext) ([]metadata.FuncReturnValue, error) {
	v.enterFmt("Retrieving return values for function %s", ctx.FuncDecl.Name)
	defer v.exit()

	retVals, err := v.context.ArbitrationProvider.Ast().GetFuncRetValMeta(
		v.fieldVisitor,
		ctx.CurrentPkg,
		ctx.SourceFile,
		ctx.FuncDecl,
		ctx.Annotations,
	)

	return retVals, err
}
