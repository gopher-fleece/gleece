package extractor

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

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

	importIdCounter uint64

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

func (v *ControllerVisitor) getNextImportId() uint64 {
	value := v.importIdCounter
	v.importIdCounter++
	return value
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

func (v *ControllerVisitor) createControllerMetadata(controllerNode *ast.TypeSpec) definitions.ControllerMetadata {
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

func (v *ControllerVisitor) visitController(controllerNode *ast.TypeSpec) (definitions.ControllerMetadata, error) {
	controller := v.createControllerMetadata(controllerNode)

	for _, file := range v.sourceFiles {
		for _, declaration := range file.Decls {
			switch funcDeclaration := declaration.(type) {
			case *ast.FuncDecl:
				if IsFuncDeclReceiverForStruct(controller.Name, funcDeclaration) {
					meta, isApiEndpoint, err := v.visitMethod(funcDeclaration)
					if err != nil {
						return controller, fmt.Errorf(
							"encountered an error visiting controller %s function %v - %v",
							controller.Name,
							funcDeclaration.Name.Name,
							err,
						)
					}
					if !isApiEndpoint {
						continue
					}
					controller.Routes = append(controller.Routes, meta)
				}
			}
		}
	}

	return controller, nil
}

func (v *ControllerVisitor) getFuncParams(funcDecl *ast.FuncDecl, comments []string) ([]definitions.FuncParam, error) {
	funcParams := []definitions.FuncParam{}

	paramTypes, err := GetFuncParameterTypeList(v.currentSourceFile, v.fileSet, v.packages, funcDecl)
	if err != nil {
		return funcParams, err
	}

	for _, param := range paramTypes {
		line := SearchForParamTerm(comments, param.Name)
		if line == "" {
			return funcParams, fmt.Errorf("no comment metadata found for %v", param.Name)
		}

		finalParamMeta := definitions.FuncParam{
			ParamMeta:          param,
			Description:        strings.TrimSpace(GetTextAfterParenthesis(line, " "+param.Name+" ")),
			UniqueImportSerial: v.getNextImportId(),
		}

		if nameInSchema := ExtractParenthesesContent(line); nameInSchema != "" {
			finalParamMeta.NameInSchema = nameInSchema
		} else {
			finalParamMeta.NameInSchema = param.Name
		}

		// Currently, only body param can be an object type
		passedIn := strings.ToLower(ExtractParamTerm(line))
		switch passedIn {
		case "query":
			finalParamMeta.PassedIn = definitions.PassedInQuery
			if !finalParamMeta.TypeMeta.IsUniverseType {
				return funcParams, createInvalidParamUsageError(finalParamMeta)
			}
		case "header":
			finalParamMeta.PassedIn = definitions.PassedInHeader
			if !finalParamMeta.TypeMeta.IsUniverseType {
				return funcParams, createInvalidParamUsageError(finalParamMeta)
			}
		case "path":
			finalParamMeta.PassedIn = definitions.PassedInPath
			if !finalParamMeta.TypeMeta.IsUniverseType {
				return funcParams, createInvalidParamUsageError(finalParamMeta)
			}
		case "body":
			finalParamMeta.PassedIn = definitions.PassedInBody
		}

		funcParams = append(funcParams, finalParamMeta)
	}

	return funcParams, nil
}

func (v *ControllerVisitor) getFuncReturnValue(funcDecl *ast.FuncDecl) ([]definitions.FuncReturnValue, error) {
	returnTypes, err := GetFuncReturnTypeList(v.currentSourceFile, v.fileSet, v.packages, funcDecl)
	if err != nil {
		return []definitions.FuncReturnValue{}, err
	}

	values := []definitions.FuncReturnValue{}

	for _, value := range returnTypes {
		values = append(
			values,
			definitions.FuncReturnValue{
				TypeMetadata:       value,
				UniqueImportSerial: v.getNextImportId(),
			},
		)
	}
	return values, nil
}

func (v *ControllerVisitor) visitMethod(funcDecl *ast.FuncDecl) (definitions.RouteMetadata, bool, error) {
	// Check whether there are any comments on the method - we expect all API endpoints to contain comments.
	// No comments - not an API endpoint.
	if funcDecl.Doc == nil || funcDecl.Doc.List == nil || len(funcDecl.Doc.List) <= 0 {
		return definitions.RouteMetadata{}, false, nil
	}

	comments := MapDocListToStrings(funcDecl.Doc.List)

	meta := definitions.RouteMetadata{
		OperationId:         funcDecl.Name.Name,
		HttpVerb:            definitions.EnsureValidHttpVerb(FindAndExtract(comments, "@Method")),
		Description:         FindAndExtract(comments, "@Description"),
		RestMetadata:        BuildRestMetadata(comments),
		ErrorResponses:      getErrorResponseMetadata(comments),
		RequestContentType:  definitions.ContentTypeJSON, // Hardcoded for now, should be supported via comments later
		ResponseContentType: definitions.ContentTypeJSON, // Hardcoded for now, should be supported via comments later
	}

	// Check whether the method is an API endpoint, i.e., has all the relevant metadata.
	// Methods without expected metadata are ignored (may switch to raising an error instead)
	isApiEndpoint := len(meta.HttpVerb) > 0 && len(meta.RestMetadata.Path) > 0
	if !isApiEndpoint {
		return meta, false, nil
	}

	// Retrieve parameter information
	funcParams, err := v.getFuncParams(funcDecl, comments)
	if err != nil {
		return meta, true, err
	}
	meta.FuncParams = funcParams

	// Set the function's return types
	responses, err := v.getFuncReturnValue(funcDecl)
	if err != nil {
		return meta, true, err
	}
	meta.Responses = responses
	meta.HasReturnValue = len(responses) > 1

	// Set the success response code based on whether function returns a value or only error (200 vs 204)
	meta.ResponseSuccessCode = getRouteSuccessResponseCode(comments, meta.HasReturnValue)

	return meta, isApiEndpoint, nil
}

func (v *ControllerVisitor) getAllSourceFiles() []*ast.File {
	result := []*ast.File{}
	for _, file := range v.sourceFiles {
		result = append(result, file)
	}
	return result
}

func createInvalidParamUsageError(param definitions.FuncParam) error {
	return fmt.Errorf(
		"parameter %s (type %s) is passed in '%s' but is not a 'universe' type (i.e., a primitive). This is not currently supported",
		param.Name,
		param.TypeMeta.Name,
		param.PassedIn,
	)
}
