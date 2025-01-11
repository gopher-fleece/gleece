package extractor

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/haimkastner/gleece/definitions"
	Logger "github.com/haimkastner/gleece/infrastructure/logger"
)

type ControllerVisitor struct {
	nodes             map[string]*ast.File
	fileSet           *token.FileSet
	controllers       []*ast.TypeSpec
	currentSourceFile *ast.File
}

func (v *ControllerVisitor) Setup(nodes map[string]*ast.File, fileSet *token.FileSet) {
	v.nodes = nodes
	v.fileSet = fileSet
}

func (v *ControllerVisitor) Visit(node ast.Node) ast.Visitor {
	switch currentNode := node.(type) {
	case *ast.File:
		// Update the current file when visiting an *ast.File node
		v.currentSourceFile = currentNode
	case *ast.TypeSpec:
		// Check if it's a struct and if it embeds GleeceController
		if structType, isOk := currentNode.Type.(*ast.StructType); isOk {
			if DoesStructEmbedStruct(
				v.currentSourceFile,
				"github.com/haimkastner/gleece/controller",
				structType,
				"GleeceController",
			) {
				v.visitController(currentNode)
			}
		}
	}
	return v
}

func (v ControllerVisitor) createControllerMetadata(controllerNode *ast.TypeSpec) definitions.ControllerMetadata {
	fullPackageName, fullNameErr := GetFullPackageName(v.currentSourceFile, v.fileSet)
	packageAlias, aliasErr := GetDefaultPackageAlias(v.currentSourceFile)

	if fullNameErr != nil || aliasErr != nil {
		panic(fmt.Sprintf("Could not obtain full/partial package name for source file '%s'", v.currentSourceFile.Name))
	}

	meta := definitions.ControllerMetadata{
		Name:                  controllerNode.Name.Name,
		FullyQualifiedPackage: fullPackageName,
		Package:               packageAlias,
	}

	if controllerNode.Doc != nil {
		comments := MapDocListToStrings(controllerNode.Doc.List)
		meta.Tag = FindAndExtract(comments, "@Tag")
		meta.Description = FindAndExtract(comments, "@Description")
		meta.RestMetadata = BuildRestMetadata(comments)
	}

	return meta
}

func (v *ControllerVisitor) visitController(controllerNode *ast.TypeSpec) {
	controller := v.createControllerMetadata(controllerNode)

	for _, file := range v.nodes {
		for _, declaration := range file.Decls {
			switch funcDeclaration := declaration.(type) {
			case *ast.FuncDecl:
				if IsFuncDeclReceiverForStruct(controller.Name, funcDeclaration) {
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
