package controller

import (
	"encoding/json"
	"go/ast"
	"slices"
	"strings"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor"
	"github.com/gopher-fleece/gleece/extractor/visitors"
)

func (v *ControllerVisitor) GetFormattedDiagnosticStack() string {
	stack := slices.Clone(v.diagnosticStack)
	slices.Reverse(stack)
	return strings.Join(stack, "\n\t")
}

func (v *ControllerVisitor) GetLastError() *error {
	return v.lastError
}

func (v *ControllerVisitor) GetFiles() []*ast.File {
	return v.getAllSourceFiles()
}

func (v ControllerVisitor) GetControllers() []definitions.ControllerMetadata {
	return v.controllers
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
			if extractor.DoesStructEmbedStruct(
				v.currentSourceFile,
				"github.com/gopher-fleece/gleece/external",
				structType,
				"GleeceController",
			) {
				controller, err := v.visitController(currentNode)
				if err != nil {
					v.lastError = &err
					return v
				}
				v.controllers = append(v.controllers, controller)
			}
		}
	}
	return v
}

func (v *ControllerVisitor) GetModelsFlat() ([]definitions.ModelMetadata, bool, error) {
	if len(v.controllers) <= 0 {
		return []definitions.ModelMetadata{}, false, nil
	}

	existingTypesMap := make(map[string]string)
	models := []definitions.TypeMetadata{}

	hasAnyErrorTypes := false
	for _, controller := range v.controllers {
		for _, route := range controller.Routes {
			encounteredErrorType, err := v.insertRouteTypeList(&existingTypesMap, &models, &route)
			if err != nil {
				return []definitions.ModelMetadata{}, false, v.frozenError(err)
			}
			if encounteredErrorType {
				hasAnyErrorTypes = true
			}
		}
	}

	typeVisitor := visitors.NewTypeVisitor(v.packages)
	for _, model := range models {
		pkg := extractor.FilterPackageByFullName(v.packages, model.FullyQualifiedPackage)
		if pkg == nil {
			return nil, hasAnyErrorTypes, v.getFrozenError(
				"could locate packages.Package '%s' whilst looking for type '%s'.\n"+
					"Please note that Gleece currently cannot use any structs from externally imported packages",
				model.FullyQualifiedPackage,
				model.Name,
			)
		}

		structNode, err := extractor.FindTypesStructInPackage(pkg, model.Name)
		if err != nil {
			return nil, hasAnyErrorTypes, v.frozenError(err)
		}

		if structNode == nil {
			return nil, hasAnyErrorTypes, v.getFrozenError("could not find struct '%s' in package '%s'", model.Name, model.FullyQualifiedPackage)
		}

		err = typeVisitor.VisitStruct(model.FullyQualifiedPackage, model.Name, structNode)
		if err != nil {
			return nil, hasAnyErrorTypes, v.frozenError(err)
		}
	}

	return typeVisitor.GetStructs(), hasAnyErrorTypes, nil
}
