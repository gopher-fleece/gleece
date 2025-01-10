package extractor

import (
	"fmt"
	"go/ast"
	"go/token"

	Logger "github.com/haimkastner/gleece/infrastructure/logger"
)

type ControllerVisitor struct {
	nodes       map[string]*ast.File
	fileSet     *token.FileSet
	controllers []*ast.TypeSpec
}

func (v *ControllerVisitor) Setup(nodes map[string]*ast.File, fileSet *token.FileSet) {
	v.nodes = nodes
	v.fileSet = fileSet
}

func (v *ControllerVisitor) Visit(node ast.Node) ast.Visitor {
	switch currentNode := node.(type) {
	case *ast.TypeSpec:
		// Check if it's a struct and if it embeds GleeceController
		if structType, isOk := currentNode.Type.(*ast.StructType); isOk {
			if DoesStructEmbedStruct(structType, "GleeceController") {
				v.controllers = append(v.controllers, currentNode)
				v.visitController(currentNode)
			}
		}
	}
	return v
}

func (v *ControllerVisitor) visitController(controllerNode *ast.TypeSpec) {
	controllerName := controllerNode.Name.Name

	for _, file := range v.nodes {
		for _, declaration := range file.Decls {
			switch funcDeclaration := declaration.(type) {
			case *ast.FuncDecl:
				if IsFuncDeclReceiverForStruct(controllerName, funcDeclaration) {
					v.extractMethodDetails(funcDeclaration)
				}
			}
		}
	}
}

func (v *ControllerVisitor) extractMethodDetails(fn *ast.FuncDecl) {
	// Method name
	methodName := fn.Name.Name
	// Return types
	returnTypes := []string{}
	if fn.Type.Results != nil {
		for _, result := range fn.Type.Results.List {
			if ident, ok := result.Type.(*ast.Ident); ok {
				returnTypes = append(returnTypes, ident.Name)
			}
		}
	}
	// Parameters and their types
	paramTypes := []string{}
	for _, param := range fn.Type.Params.List {
		for _, paramName := range param.Names {
			if ident, ok := param.Type.(*ast.Ident); ok {
				paramTypes = append(paramTypes, fmt.Sprintf("%s %s", paramName.Name, ident.Name))
			}
		}
	}
	// Method comments
	methodComments := ""
	if fn.Doc != nil {
		methodComments = fn.Doc.Text()
	}

	// Print method details
	Logger.Debug("Method: %s\n", methodName)
	Logger.Debug("  Return Types: %v\n", returnTypes)
	Logger.Debug("  Parameters: %v\n", paramTypes)
	Logger.Debug("  Comments: %v\n", methodComments)
}
