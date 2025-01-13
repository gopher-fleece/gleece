package extractor

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/bmatcuk/doublestar/v4"
	MapSet "github.com/deckarep/golang-set/v2"
	"github.com/haimkastner/gleece/definitions"
	Logger "github.com/haimkastner/gleece/infrastructure/logger"
	"golang.org/x/tools/go/packages"
)

type ControllerVisitor struct {
	sourceFiles map[string]*ast.File
	fileSet     *token.FileSet
	packages    []*packages.Package

	controllerNodes []*ast.TypeSpec
	controllers     []definitions.ControllerMetadata

	// Context
	currentSourceFile *ast.File
	currentGenDecl    *ast.GenDecl

	lastError *error
}

func (v *ControllerVisitor) Init(sourceFileGlobs []string) error {
	return v.loadMappings(sourceFileGlobs)
}

func (v *ControllerVisitor) GetLastError() *error {
	return v.lastError
}

func (v *ControllerVisitor) GetFiles() []*ast.File {
	return v.getAllSourceFiles()
}

func (v *ControllerVisitor) loadMappings(sourceFileGlobs []string) error {
	v.sourceFiles = make(map[string]*ast.File)
	v.fileSet = token.NewFileSet()

	packages := MapSet.NewSet[string]()

	for _, globExpr := range sourceFileGlobs {
		sourceFiles, err := doublestar.FilepathGlob(globExpr)
		if err != nil {
			v.lastError = &err
			return err
		}

		for _, sourceFile := range sourceFiles {
			file, err := parser.ParseFile(v.fileSet, sourceFile, nil, parser.ParseComments)
			if err != nil {
				Logger.Error("Error parsing file %s - %v", sourceFile, err)
				v.lastError = &err
				return err
			}
			v.sourceFiles[sourceFile] = file

			packageName, err := GetFullPackageName(file, v.fileSet)
			if err != nil {
				return err
			}
			packages.Add(packageName)
		}
	}

	info, err := GetPackagesFromExpressions(packages.ToSlice())
	if err != nil {
		return err
	}
	v.packages = info

	return nil
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
			if DoesStructEmbedStruct(
				v.currentSourceFile,
				"github.com/haimkastner/gleece/controller",
				structType,
				"GleeceController",
			) {
				v.controllerNodes = append(v.controllerNodes, currentNode)
				v.controllers = append(v.controllers, v.visitController(currentNode))
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

	// Comments are usually located on the nearest GenDecl but may also be inlined on the struct itself
	var commentSource *ast.CommentGroup
	if controllerNode.Doc != nil {
		commentSource = controllerNode.Doc
	} else {
		commentSource = v.currentGenDecl.Doc
	}

	if commentSource != nil {
		comments := MapDocListToStrings(commentSource.List)
		meta.Tag = FindAndExtract(comments, "@Tag")
		meta.Description = FindAndExtract(comments, "@Description")
		meta.RestMetadata = BuildRestMetadata(comments)
	}
	return meta
}

func (v *ControllerVisitor) visitController(controllerNode *ast.TypeSpec) definitions.ControllerMetadata {
	controller := v.createControllerMetadata(controllerNode)

	for _, file := range v.sourceFiles {
		for _, declaration := range file.Decls {
			switch funcDeclaration := declaration.(type) {
			case *ast.FuncDecl:
				if IsFuncDeclReceiverForStruct(controller.Name, funcDeclaration) {
					meta, isApiEndpoint, err := v.visitMethod(funcDeclaration)
					if err != nil {
						panic(fmt.Sprintf("Encountered an error visiting function %v", funcDeclaration.Name.Name))
					}
					if !isApiEndpoint {
						continue
					}
					controller.Routes = append(controller.Routes, meta)
				}
			}
		}
	}

	return controller
}

func (v *ControllerVisitor) visitMethod(funcDecl *ast.FuncDecl) (definitions.RouteMetadata, bool, error) {
	if funcDecl.Doc == nil || funcDecl.Doc.List == nil || len(funcDecl.Doc.List) <= 0 {
		return definitions.RouteMetadata{}, false, nil
	}

	comments := MapDocListToStrings(funcDecl.Doc.List)
	responseInterface := GetResponseInterface(v.currentSourceFile, v.fileSet, *funcDecl)

	meta := definitions.RouteMetadata{
		OperationId:         funcDecl.Name.Name,
		HttpVerb:            definitions.EnsureValidHttpVerb(FindAndExtract(comments, "@Method")),
		Description:         FindAndExtract(comments, "@Description"),
		RestMetadata:        BuildRestMetadata(comments),
		ResponseInterface:   responseInterface,
		ResponseSuccessCode: getRouteSuccessResponseCode(comments, responseInterface.InterfaceName != ""),
		ErrorResponses:      getErrorResponseMetadata(comments),
	}

	// Extract function parameters
	if funcDecl.Type.Params != nil {
		meta.FuncParams = getRouteParameters(comments, *funcDecl)
	}

	// Extract function results
	if funcDecl.Type.Results != nil {
	}

	isApiEndpoint := len(meta.HttpVerb) > 0 && len(meta.RestMetadata.Path) > 0

	GetFuncReturnTypeList(v.currentSourceFile, v.fileSet, v.packages, *funcDecl)
	return meta, isApiEndpoint, nil
}

func (v *ControllerVisitor) getAllSourceFiles() []*ast.File {
	result := []*ast.File{}
	for _, file := range v.sourceFiles {
		result = append(result, file)
	}
	return result
}
