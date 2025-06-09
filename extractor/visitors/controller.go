package visitors

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"slices"
	"strings"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
)

type ControllerWithSymbolNode struct {
	Controller definitions.ControllerMetadata
	SymbolNode *symboldg.SymbolNode
}

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
	currentController *definitions.ControllerMetadata

	currentFVersion gast.FileVersion

	// A list of fully processed controller metadata, ready to be passed to the routes/spec generators
	controllers []ControllerWithSymbolNode
}

// NewControllerVisitor Instantiates a new Gleece Controller visitor.
func NewControllerVisitor(context *VisitContext) (*ControllerVisitor, error) {
	visitor := ControllerVisitor{}
	err := visitor.initialize((context))
	return &visitor, err
}

func (v ControllerVisitor) GetControllers() []definitions.ControllerMetadata {
	var metadata []definitions.ControllerMetadata

	for _, controller := range v.controllers {
		metadata = append(metadata, controller.Controller)
	}

	return metadata
}

func (v ControllerVisitor) DumpContext() (string, error) {
	dump, err := json.MarshalIndent(v.controllers, "", "\t")
	if err != nil {
		return "", err
	}
	return string(dump), err
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
					v.lastError = &err
					return v
				}

				symNode, err := v.insertControllerToGraph(currentNode, controller)
				if err != nil {
					v.lastError = &err
				}

				v.controllers = append(
					v.controllers,
					ControllerWithSymbolNode{
						Controller: controller,
						SymbolNode: symNode,
					},
				)
			}
		}
	}
	return v
}

func (v *ControllerVisitor) insertControllerToGraph(
	node *ast.TypeSpec,
	metadata definitions.ControllerMetadata,
) (*symboldg.SymbolNode, error) {
	v.enter(fmt.Sprintf("Graph insertion - Controller %s", metadata.Name))
	defer v.exit()

	return v.context.GraphBuilder.AddController(
		symboldg.CreateControllerNode{
			Decl: node,
			Data: symboldg.ControllerSymbolicMetadata{
				Name:         metadata.Name,
				Package:      metadata.Package,
				PkgPath:      metadata.PkgPath,
				Tag:          metadata.Tag,
				Description:  metadata.Description,
				RestMetadata: metadata.RestMetadata,
				Security:     metadata.Security,
				FVersion:     v.currentFVersion,
			},
		},
	)
}

func (v *ControllerVisitor) GetModelsFlat() (*definitions.Models, bool, error) {
	v.enter(fmt.Sprintf("%d controllers", len(v.controllers)))
	defer v.exit()

	if len(v.controllers) <= 0 {
		return nil, false, nil
	}

	existingTypesMap := make(map[string]string)
	models := []definitions.TypeMetadata{}

	hasAnyErrorTypes := false
	for _, controller := range v.controllers {
		for _, route := range controller.Controller.Routes {
			encounteredErrorType, err := v.insertRouteTypeList(&existingTypesMap, &models, &route)
			if err != nil {
				return nil, false, v.frozenError(err)
			}
			if encounteredErrorType {
				hasAnyErrorTypes = true
			}
		}
	}

	typeVisitor, err := NewTypeVisitor(v.context)
	if err != nil {
		logger.Error("Could not create a new TypeVisitor by Arbitration Provider - %v", err)
		return nil, false, v.frozenError(err)
	}

	for _, model := range models {

		// Ignore Context parameters - they're injected at the template levels and
		// do not reach the OpenAPI schema
		if model.Name == "Context" && model.PkgPath == "context" {
			continue
		}

		pkg, err := v.context.ArbitrationProvider.Pkg().GetPackage(model.PkgPath)
		if err != nil {
			return nil, hasAnyErrorTypes, v.frozenError(err)
		}

		if pkg == nil {
			return nil, hasAnyErrorTypes, v.getFrozenError(
				"could locate packages.Package '%s' whilst looking for type '%s'.\n"+
					"Please note that Gleece currently cannot use any structs from externally imported packages",
				model.PkgPath,
				model.Name,
			)
		}

		// Currently, Name includes a "[]" prefix if the type is an array.
		// Need to remove it so lookup can actually succeed.
		// Might move to an "IsArray" field in the near future.
		cleanedName := common.UnwrapArrayTypeString(model.Name)

		// Enums are handled separately
		if model.SymbolKind == common.SymKindAlias {
			err := typeVisitor.VisitEnum(cleanedName, model)
			if err != nil {
				return nil, hasAnyErrorTypes, v.frozenError(err)
			}
			continue
		}

		structNode, err := gast.FindTypesStructInPackage(pkg, cleanedName)
		if err != nil {
			return nil, hasAnyErrorTypes, v.frozenError(err)
		}

		if structNode == nil {
			return nil,
				hasAnyErrorTypes,
				v.getFrozenError(
					"could not find struct '%s' in package '%s'",
					cleanedName,
					model.PkgPath,
				)
		}

		err = typeVisitor.VisitStruct(model.PkgPath, cleanedName, structNode)
		if err != nil {
			return nil, hasAnyErrorTypes, v.frozenError(err)
		}
	}

	structs := typeVisitor.GetStructs()
	enums := typeVisitor.GetEnums()

	slices.SortFunc(structs, func(a, b definitions.StructMetadata) int {
		return strings.Compare(a.Name, b.Name)
	})

	slices.SortFunc(enums, func(a, b definitions.EnumMetadata) int {
		return strings.Compare(a.Name, b.Name)
	})

	flatModels := &definitions.Models{
		Structs: structs,
		Enums:   enums,
	}

	return flatModels, hasAnyErrorTypes, nil
}

// visitController traverses a controller node to extract metadata for API routes.
// A controller is a struct that embeds GleeceController.
// The function enumerates receivers to gather route construction details.
func (v *ControllerVisitor) visitController(controllerNode *ast.TypeSpec) (definitions.ControllerMetadata, error) {
	v.enter(fmt.Sprintf("Controller '%s'", controllerNode.Name.Name))
	defer v.exit()

	fVersion, err := gast.NewFileVersionFromAstFile(v.currentSourceFile, v.context.ArbitrationProvider.FileSet())
	if err != nil {
		return definitions.ControllerMetadata{}, v.frozenError(err)
	}

	v.currentFVersion = fVersion

	controller, err := v.createControllerMetadata(controllerNode)
	v.currentController = &controller

	if err != nil {
		return controller, err
	}

	routeVisitor, err := NewRouteVisitor(
		v.context, RouteParentContext{
			Decl:     controllerNode,
			FVersion: fVersion,
			Metadata: &controller,
		},
	)

	if err != nil {
		logger.Error("Could not initialize a new route visitor - %v", err)
		return *v.currentController, v.frozenError(err)
	}

	// Go over all enumerated source files and look for receivers for the controller
	for _, file := range v.context.ArbitrationProvider.GetAllSourceFiles() {
		for _, declaration := range file.Decls {
			switch funcDeclaration := declaration.(type) {
			case *ast.FuncDecl:
				if gast.IsFuncDeclReceiverForStruct(controller.Name, funcDeclaration) {
					// If the function is a relevant receiver, visit it and extract metadata
					meta, isApiEndpoint, err := routeVisitor.VisitMethod(funcDeclaration, file)
					if err != nil {
						return controller, v.getFrozenError(
							"encountered an error visiting controller %s method %v - %v",
							controller.Name,
							funcDeclaration.Name.Name,
							err,
						)
					}

					if !isApiEndpoint {
						// If the receiver is deemed to not be an API endpoint, ignore it
						continue
					}

					controller.Routes = append(controller.Routes, meta)
				}
			}
		}
	}

	return controller, nil
}

// createControllerMetadata Creates a standard ControllerMetadata struct for the given node
func (v *ControllerVisitor) createControllerMetadata(controllerNode *ast.TypeSpec) (definitions.ControllerMetadata, error) {
	fullPackageName, fullNameErr := gast.GetFullPackageName(v.currentSourceFile, v.context.ArbitrationProvider.FileSet())
	packageAlias, aliasErr := gast.GetDefaultPackageAlias(v.currentSourceFile)

	if fullNameErr != nil || aliasErr != nil {
		return definitions.ControllerMetadata{}, v.getFrozenError(
			"could not obtain full/partial package name for source file '%s'", v.currentSourceFile.Name,
		)
	}

	// Start off by filling the name and package
	meta := definitions.ControllerMetadata{
		Name:    controllerNode.Name.Name,
		PkgPath: fullPackageName,
		Package: packageAlias,
	}

	// Comments are usually located on the nearest GenDecl but may also be inlined on the struct itself
	var commentSource *ast.CommentGroup
	if controllerNode.Doc != nil {
		commentSource = controllerNode.Doc
	} else {
		commentSource = v.currentGenDecl.Doc
	}

	// Do we want to fail if there are no attributes on the controller?
	if commentSource != nil {
		comments := gast.MapDocListToStrings(commentSource.List)
		holder, err := annotations.NewAnnotationHolder(comments, annotations.CommentSourceController)
		if err != nil {
			return meta, v.frozenError(err)
		}

		// Parse any explicit Security annotations
		security, err := getSecurityFromContext(holder)
		if err != nil {
			return meta, v.frozenError(err)
		}

		// If there are no explicitly defined securities, check for inherited ones
		if len(security) <= 0 {
			logger.Debug("Controller %s does not have explicit security; Using user-defined defaults", meta.Name)
			security = getDefaultSecurity(v.context.GleeceConfig)
		}

		meta.Tag = holder.GetFirstValueOrEmpty(annotations.AttributeTag)
		meta.Description = holder.GetFirstDescriptionOrEmpty(annotations.AttributeDescription)
		meta.RestMetadata = definitions.RestMetadata{Path: holder.GetFirstValueOrEmpty(annotations.AttributeRoute)}
		meta.Security = security
	}

	return meta, nil
}

func (v *ControllerVisitor) addToTypeMap(
	existingTypesMap *map[string]string,
	existingModels *[]definitions.TypeMetadata,
	typeMeta definitions.TypeMetadata,
) error {
	if typeMeta.IsUniverseType {
		return nil
	}

	existsInPackage, exists := (*existingTypesMap)[typeMeta.Name]
	if exists {
		if existsInPackage == typeMeta.PkgPath {
			// Same type referenced from a separate location
			return nil
		}

		return v.getFrozenError(
			"type '%s' exists in more that one package (%s and %s). This is not currently supported",
			typeMeta.Name,
			typeMeta.PkgPath,
			existsInPackage,
		)
	}

	(*existingTypesMap)[typeMeta.Name] = typeMeta.PkgPath
	(*existingModels) = append((*existingModels), typeMeta)
	return nil
}

func (v *ControllerVisitor) insertRouteTypeList(
	existingTypesMap *map[string]string,
	existingModels *[]definitions.TypeMetadata,
	route *definitions.RouteMetadata,
) (bool, error) {

	plainErrorEncountered := false
	for _, param := range route.FuncParams {
		if param.TypeMeta.IsUniverseType && param.TypeMeta.Name == "error" && param.TypeMeta.PkgPath == "" {
			// Mark whether we've encountered any 'error' type
			plainErrorEncountered = true
		}
		err := v.addToTypeMap(existingTypesMap, existingModels, param.TypeMeta)
		if err != nil {
			return plainErrorEncountered, v.frozenError(err)
		}
	}

	for _, param := range route.Responses {
		if param.IsUniverseType && param.Name == "error" && param.PkgPath == "" {
			// Mark whether we've encountered any 'error' type
			plainErrorEncountered = true
		}
		err := v.addToTypeMap(existingTypesMap, existingModels, param.TypeMetadata)
		if err != nil {
			return plainErrorEncountered, v.frozenError(err)
		}
	}

	return plainErrorEncountered, nil
}
