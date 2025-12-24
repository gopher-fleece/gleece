package visitors

import (
	"fmt"
	"go/ast"

	"github.com/gopher-fleece/gleece/v2/common"
	"github.com/gopher-fleece/gleece/v2/core/annotations"
	"github.com/gopher-fleece/gleece/v2/core/metadata"
	"github.com/gopher-fleece/gleece/v2/gast"
	"github.com/gopher-fleece/gleece/v2/graphs/symboldg"
	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
)

// ControllerVisitor locates and walks over Gleece controllers to extract metadata and route information
// This construct serves as the main business logic for Gleece's context generation
type ControllerVisitor struct {
	BaseVisitor

	// The file currently being worked on
	currentSourceFile *ast.File

	// The last-encountered GenDecl.
	//
	// Documentation may be placed on TypeDecl or their parent GenDecl so we track these,
	// in case we need to fetch the docs from the TypeDecl's parent.
	currentGenDecl *ast.GenDecl

	// The current Gleece Controller being processed
	currentController *metadata.ControllerMeta

	currentFVersion gast.FileVersion

	// A list of fully processed controller metadata, ready to be passed to the routes/spec generators
	controllers []metadata.ControllerMeta

	fieldVisitor *FieldVisitor
}

// NewControllerVisitor Instantiates a new Gleece Controller visitor.
func NewControllerVisitor(context *VisitContext) (*ControllerVisitor, error) {
	visitor := ControllerVisitor{}
	err := visitor.initialize((context))
	return &visitor, err
}

func (v *ControllerVisitor) setFieldVisitor(visitor *FieldVisitor) {
	v.fieldVisitor = visitor
}

// GetControllers returns all controllers known by this visitor.
// Note that the returned values are mutable.
// When used, care must be taken to not corrupt the internal state
func (v ControllerVisitor) GetControllers() []metadata.ControllerMeta {
	return v.controllers
}

func (v *ControllerVisitor) Visit(node ast.Node) ast.Visitor {
	switch currentNode := node.(type) {
	case *ast.File:
		// Update the current file when visiting an *ast.File node
		v.currentSourceFile = currentNode
	case *ast.GenDecl:
		v.currentGenDecl = currentNode
	case *ast.TypeSpec:
		// Check if it's a struct and if it embeds GleeceController
		if structType, isOk := currentNode.Type.(*ast.StructType); isOk {
			if gast.DoesStructEmbedStruct(
				v.currentSourceFile,
				structType,
				"github.com/gopher-fleece/runtime",
				"GleeceController",
			) {
				controller, err := v.visitController(currentNode)
				if err != nil {
					v.setLastError(err)
					return v
				}

				err = v.addSelfToGraph(controller)
				if err != nil {
					v.setLastError(err)
					return v
				}

				v.controllers = append(v.controllers, controller)
			}
		}
	}
	return v
}

func (v *ControllerVisitor) addSelfToGraph(meta metadata.ControllerMeta) error {
	v.enter(fmt.Sprintf("Graph insertion - Controller %s", meta.Struct.Name))
	defer v.exit()

	_, err := v.context.Graph.AddController(
		symboldg.CreateControllerNode{
			Data:        meta,
			Annotations: meta.Struct.Annotations,
		},
	)

	v.context.MetadataCache.AddController(&meta)

	return err
}

// visitController traverses a controller node to extract metadata for API routes.
// A controller is a struct that embeds GleeceController.
// The function enumerates receivers to gather route construction details.
func (v *ControllerVisitor) visitController(controllerNode *ast.TypeSpec) (metadata.ControllerMeta, error) {
	v.enter(fmt.Sprintf("Controller '%s'", controllerNode.Name.Name))
	defer v.exit()

	fVersion, err := gast.NewFileVersionFromAstFile(
		v.currentSourceFile,
		v.context.ArbitrationProvider.Pkg().FSet(),
	)
	if err != nil {
		return metadata.ControllerMeta{}, v.frozenError(err)
	}

	v.currentFVersion = fVersion

	controllerMeta, err := v.createControllerMetadata(controllerNode)
	v.currentController = &controllerMeta

	if err != nil {
		return controllerMeta, err
	}

	routeVisitor, err := NewRouteVisitor(
		v.context,
		RouteParentContext{Controller: &controllerMeta},
	)
	routeVisitor.setFieldVisitor(v.fieldVisitor)

	if err != nil {
		logger.Error("Could not initialize a new route visitor - %v", err)
		return *v.currentController, v.frozenError(err)
	}

	// Go over all enumerated source files and look for receivers for the controller
	for _, file := range v.context.ArbitrationProvider.GetAllSourceFiles() {
		for _, declaration := range file.Decls {
			switch funcDeclaration := declaration.(type) {
			case *ast.FuncDecl:
				if gast.IsFuncDeclReceiverForStruct(controllerMeta.Struct.Name, funcDeclaration) {
					// If the function is a relevant receiver, visit it and extract metadata
					receiverMeta, err := routeVisitor.VisitMethod(funcDeclaration, file)
					if err != nil {
						return controllerMeta, v.getFrozenError(
							"encountered an error visiting controller %s method %v - %v",
							controllerMeta.Struct.Name,
							funcDeclaration.Name.Name,
							err,
						)
					}

					if receiverMeta == nil {
						// Nil meta implies the receiver is not an API endpoint and can be ignored
						continue
					}

					controllerMeta.Receivers = append(controllerMeta.Receivers, *receiverMeta)
				}
			}
		}
	}

	return controllerMeta, nil
}

// createControllerMetadata Creates a standard ControllerMetadata struct for the given node
func (v *ControllerVisitor) createControllerMetadata(controllerNode *ast.TypeSpec) (metadata.ControllerMeta, error) {
	v.enterFmt("Creating metadata for controller %s", controllerNode.Name.Name)
	defer v.exit()

	pkg, err := v.context.ArbitrationProvider.Pkg().GetPackageForFile(v.currentSourceFile)
	if err != nil || pkg == nil {
		return metadata.ControllerMeta{}, v.getFrozenError(
			"could not obtain full/partial package name for source file '%s'", v.currentSourceFile.Name,
		)
	}

	// Start off by filling the name and package
	meta := metadata.ControllerMeta{
		Struct: metadata.StructMeta{
			SymNodeMeta: metadata.SymNodeMeta{
				Name:    controllerNode.Name.Name,
				PkgPath: pkg.PkgPath,
				Range: common.ResolveNodeRange(
					v.context.ArbitrationProvider.Pkg().FSet(),
					controllerNode,
				),
			},
		},
	}

	comments := gast.GetCommentsFromTypeSpec(
		controllerNode,
		v.currentGenDecl,
		v.context.ArbitrationProvider.Pkg().FSet(),
	)

	annotationHolder, err := annotations.NewAnnotationHolder(comments, annotations.CommentSourceController)
	if err != nil {
		return metadata.ControllerMeta{}, v.frozenError(err)
	}

	result := metadata.ControllerMeta{
		//Controller: meta,
		Struct: metadata.StructMeta{
			SymNodeMeta: metadata.SymNodeMeta{
				Name:        meta.Struct.Name,
				Node:        controllerNode,
				SymbolKind:  common.SymKindStruct,
				PkgPath:     meta.Struct.PkgPath,
				Annotations: &annotationHolder,
				FVersion:    &v.currentFVersion,
			},
		},
	}

	return result, nil
}
