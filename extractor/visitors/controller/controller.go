package controller

import (
	"fmt"
	"go/ast"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor"
	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
)

// visitController traverses a controller node to extract metadata for API routes.
// A controller is a struct that embeds GleeceController.
// The function enumerates receivers to gather route construction details.
func (v *ControllerVisitor) visitController(controllerNode *ast.TypeSpec) (definitions.ControllerMetadata, error) {
	v.enter(fmt.Sprintf("Controller '%s'", controllerNode.Name.Name))
	defer v.exit()

	controller, err := v.createControllerMetadata(controllerNode)
	v.currentController = &controller

	if err != nil {
		return controller, err
	}

	// Go over all enumerated source files and look for receivers for the controller
	for _, file := range v.sourceFiles {
		for _, declaration := range file.Decls {
			switch funcDeclaration := declaration.(type) {
			case *ast.FuncDecl:
				if extractor.IsFuncDeclReceiverForStruct(controller.Name, funcDeclaration) {
					// If the function is a relevant receiver, visit it and extract metadata
					meta, isApiEndpoint, err := v.visitMethod(funcDeclaration)
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

func (v *ControllerVisitor) createControllerMetadata(controllerNode *ast.TypeSpec) (definitions.ControllerMetadata, error) {
	fullPackageName, fullNameErr := extractor.GetFullPackageName(v.currentSourceFile, v.fileSet)
	packageAlias, aliasErr := extractor.GetDefaultPackageAlias(v.currentSourceFile)

	if fullNameErr != nil || aliasErr != nil {
		return definitions.ControllerMetadata{}, v.getFrozenError(
			"could not obtain full/partial package name for source file '%s'", v.currentSourceFile.Name,
		)
	}

	meta := definitions.ControllerMetadata{
		Name:                  controllerNode.Name.Name,
		FullyQualifiedPackage: fullPackageName,
		Package:               packageAlias,
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
		comments := extractor.MapDocListToStrings(commentSource.List)
		holder, err := annotations.NewAnnotationHolder(comments)
		if err != nil {
			return meta, v.frozenError(err)
		}

		security, err := v.getSecurityFromContext(holder)
		if err != nil {
			return meta, v.frozenError(err)
		}

		if len(security) <= 0 {
			logger.Debug("Controller %s does not have explicit security; Using user-defined defaults", meta.Name)
			security = v.getDefaultSecurity()
		}

		meta.Tag = holder.GetFirstValueOrEmpty(annotations.AttributeTag)
		meta.Description = holder.GetFirstDescriptionOrEmpty(annotations.AttributeDescription)
		meta.RestMetadata = definitions.RestMetadata{Path: holder.GetFirstValueOrEmpty(annotations.AttributeRoute)}
		meta.Security = security
	}

	return meta, nil
}
