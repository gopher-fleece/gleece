package controller

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"slices"
	"strings"

	"github.com/gopher-fleece/gleece/common"
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
				structType,
				"github.com/gopher-fleece/runtime",
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
		for _, route := range controller.Routes {
			encounteredErrorType, err := v.insertRouteTypeList(&existingTypesMap, &models, &route)
			if err != nil {
				return nil, false, v.frozenError(err)
			}
			if encounteredErrorType {
				hasAnyErrorTypes = true
			}
		}
	}

	typeVisitor := visitors.NewTypeVisitor(&v.packagesFacade, &v.astArbitrator)
	for _, model := range models {
		pkg, err := v.packagesFacade.GetPackage(model.FullyQualifiedPackage)
		if err != nil {
			return nil, hasAnyErrorTypes, v.frozenError(err)
		}

		if pkg == nil {
			return nil, hasAnyErrorTypes, v.getFrozenError(
				"could locate packages.Package '%s' whilst looking for type '%s'.\n"+
					"Please note that Gleece currently cannot use any structs from externally imported packages",
				model.FullyQualifiedPackage,
				model.Name,
			)
		}

		// Currently, Name includes a "[]" prefix if the type is an array.
		// Need to remove it so lookup can actually succeed.
		// Might move to an "IsArray" field in the near future.
		cleanedName := common.UnwrapArrayTypeString(model.Name)

		// Enums are handled separately
		if model.EntityKind == definitions.AstNodeKindAlias {
			err := typeVisitor.VisitEnum(cleanedName, model)
			if err != nil {
				return nil, hasAnyErrorTypes, v.frozenError(err)
			}
			continue
		}

		structNode, err := extractor.FindTypesStructInPackage(pkg, cleanedName)
		if err != nil {
			return nil, hasAnyErrorTypes, v.frozenError(err)
		}

		if structNode == nil {
			return nil,
				hasAnyErrorTypes,
				v.getFrozenError(
					"could not find struct '%s' in package '%s'",
					cleanedName,
					model.FullyQualifiedPackage,
				)
		}

		err = typeVisitor.VisitStruct(model.FullyQualifiedPackage, cleanedName, structNode)
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
