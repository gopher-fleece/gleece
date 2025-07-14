package visitors

import (
	"fmt"
	"go/ast"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/extractor/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"golang.org/x/tools/go/packages"
)

type RouteParentContext struct {
	Controller *ControllerWithStructMeta
}

type RouteVisitor struct {
	BaseVisitor

	// The file currently being worked on
	currentSourceFile *ast.File

	currentFuncDecl *ast.FuncDecl

	currentAnnotationHolder *annotations.AnnotationHolder

	currentFVersion *gast.FileVersion

	currentPackage *packages.Package

	parent RouteParentContext
	// gleeceConfig *definitions.GleeceConfig

	typeVisitor *RecursiveTypeVisitor
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

	typeVisitor, err := NewTypeVisitor(visitor.context)
	if err != nil {
		return &visitor, nil
	}

	visitor.typeVisitor = typeVisitor
	return &visitor, err
}

// visitMethod Visits a controller route given as a FuncDecl and returns its metadata and whether it is an API endpoint
func (v *RouteVisitor) VisitMethod(funcDecl *ast.FuncDecl, sourceFile *ast.File) (*metadata.ReceiverMeta, error) {
	v.enter(fmt.Sprintf("Method '%s'", funcDecl.Name.Name))
	defer v.exit()

	isRoute, err := v.initializeInnerContext(sourceFile, funcDecl)
	if !isRoute || err != nil {
		return nil, err
	}

	cached := v.context.MetadataCache.GetReceiver(graphs.SymbolKeyFor(v.currentFuncDecl, v.currentFVersion))
	if cached != nil {
		return cached, nil
	}

	metadata, err := v.constructRouteMetadata()
	return metadata, err
}

func (v *RouteVisitor) initializeInnerContext(sourceFile *ast.File, funcDecl *ast.FuncDecl) (bool, error) {
	v.enter(fmt.Sprintf("Method '%s' - Initialization", funcDecl.Name.Name))
	defer v.exit()

	// Sets the context for the visit
	v.currentSourceFile = sourceFile
	v.currentFuncDecl = funcDecl

	// Check whether there are any comments on the method - we expect all API endpoints to contain comments.
	// No comments - not an API endpoint.
	if funcDecl.Doc == nil || funcDecl.Doc.List == nil || len(funcDecl.Doc.List) <= 0 {
		return false, nil
	}

	pkgPath, err := gast.GetFullPackageName(
		v.currentSourceFile,
		v.context.ArbitrationProvider.Pkg().FSet(),
	)
	if err != nil {
		return false, v.frozenError(err)
	}

	pkg, err := v.context.ArbitrationProvider.Pkg().GetPackage(pkgPath)
	if err != nil {
		return false, v.frozenError(err)
	}

	if pkg == nil {
		return false, v.getFrozenError("could not obtain package object for path %s", pkgPath)
	}

	v.currentPackage = pkg

	comments := gast.MapDocListToStrings(funcDecl.Doc.List)
	holder, err := annotations.NewAnnotationHolder(comments, annotations.CommentSourceRoute)
	if err != nil {
		// Couldn't read comments. Fail.
		return false, v.frozenError(err)
	}

	// Validations are deferred/delegated to after visitation but these here serve to avoid doing unnecessary work.
	methodAttr := holder.GetFirst(annotations.GleeceAnnotationMethod)
	if methodAttr == nil {
		logger.Info("Method '%s' does not have a @Method attribute and will be ignored", funcDecl.Name.Name)
		return false, nil
	}

	routePath := holder.GetFirstValueOrEmpty(annotations.GleeceAnnotationRoute)
	if len(routePath) <= 0 {
		logger.Info("Method '%s' does not have an @Route attribute and will be ignored", funcDecl.Name.Name)
		return true, nil
	}

	v.currentAnnotationHolder = &holder

	fVersion, err := gast.NewFileVersionFromAstFile(
		v.currentSourceFile,
		v.context.ArbitrationProvider.Pkg().FSet(),
	)
	if err != nil {
		return true, v.frozenError(err)
	}

	v.currentFVersion = &fVersion
	return true, nil
}

func (v *RouteVisitor) constructRouteMetadata() (*metadata.ReceiverMeta, error) {
	params, err := v.getFuncParams()
	if err != nil {
		return nil, v.frozenError(err)
	}

	retVals, err := v.getFuncRetVals()
	if err != nil {
		return nil, v.frozenError(err)
	}

	meta := &metadata.ReceiverMeta{
		SymNodeMeta: metadata.SymNodeMeta{
			Name:        v.currentFuncDecl.Name.Name,
			Node:        v.currentFuncDecl,
			SymbolKind:  common.SymKindReceiver,
			PkgPath:     v.currentPackage.PkgPath,
			Annotations: v.currentAnnotationHolder,
			FVersion:    v.currentFVersion,
		},
		Params:  params,
		RetVals: retVals,
	}

	v.context.MetadataCache.AddReceiver(meta)

	_, err = v.context.GraphBuilder.AddRoute(
		symboldg.CreateRouteNode{
			Data: meta,
			ParentController: symboldg.KeyableNodeMeta{
				Decl:     v.parent.Controller.StructMeta.Node,
				FVersion: *v.parent.Controller.StructMeta.FVersion,
			},
		},
	)

	if err != nil {
		return nil, v.frozenError(err)
	}

	for _, param := range params {
		v.context.GraphBuilder.AddRouteParam(symboldg.CreateParameterNode{
			Data: param,
			ParentRoute: symboldg.KeyableNodeMeta{
				Decl:     meta.Node,
				FVersion: *meta.FVersion,
			},
		})
	}

	for _, retVal := range retVals {
		v.context.GraphBuilder.AddRouteRetVal(symboldg.CreateReturnValueNode{
			Data: retVal,
			ParentRoute: symboldg.KeyableNodeMeta{
				Decl:     meta.Node,
				FVersion: *meta.FVersion,
			},
		})
	}

	return meta, nil
}

func (v *RouteVisitor) getFuncParams() ([]metadata.FuncParam, error) {
	v.enter("")
	defer v.exit()

	paramTypes, err := v.context.ArbitrationProvider.Ast().GetFuncParametersMeta(
		v.typeVisitor,
		v.currentPackage,
		v.currentSourceFile,
		v.currentFuncDecl,
		v.currentAnnotationHolder,
	)

	return paramTypes, err
}

func (v *RouteVisitor) getFuncRetVals() ([]metadata.FuncReturnValue, error) {
	v.enter("")
	defer v.exit()

	retVals, err := v.context.ArbitrationProvider.Ast().GetFuncRetValMeta(
		v.typeVisitor,
		v.currentPackage,
		v.currentSourceFile,
		v.currentFuncDecl,
		v.currentAnnotationHolder,
	)

	return retVals, err
}
