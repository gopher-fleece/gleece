package extractor

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"runtime"
	"slices"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	MapSet "github.com/deckarep/golang-set/v2"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/external"
	Logger "github.com/gopher-fleece/gleece/infrastructure/logger"
	"golang.org/x/tools/go/packages"
)

type ControllerVisitor struct {
	// Data
	config      *definitions.GleeceConfig
	sourceFiles map[string]*ast.File
	fileSet     *token.FileSet
	packages    []*packages.Package

	// Context
	currentSourceFile *ast.File
	currentGenDecl    *ast.GenDecl
	currentController *definitions.ControllerMetadata
	importIdCounter   uint64

	// Diagnostics
	stackFrozen     bool
	diagnosticStack []string
	lastError       *error

	// Output
	controllers []definitions.ControllerMetadata
}

func (v *ControllerVisitor) GetFormattedDiagnosticStack() string {
	stack := slices.Clone(v.diagnosticStack)
	slices.Reverse(stack)
	return strings.Join(stack, "\n\t")
}

func NewControllerVisitor(config *definitions.GleeceConfig) (*ControllerVisitor, error) {
	visitor := ControllerVisitor{}
	visitor.config = config

	var globs []string
	if len(config.CommonConfig.ControllerGlobs) > 0 {
		globs = config.CommonConfig.ControllerGlobs
	} else {
		globs = []string{"./*.go", "./**/*.go"}
	}

	return &visitor, visitor.loadMappings(globs)
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
			if DoesStructEmbedStruct(
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

func (v ControllerVisitor) getDefaultSecurity() []definitions.RouteSecurity {
	return v.config.OpenAPIGeneratorConfig.DefaultRouteSecurity
}

func (v *ControllerVisitor) enter(message string) {
	if v.stackFrozen {
		return
	}

	var formattedMessage string
	if len(message) > 0 {
		formattedMessage = fmt.Sprintf("- (%s)", message)
	}

	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		v.diagnosticStack = append(v.diagnosticStack, fmt.Sprintf("<unknown>.<unknown> - %s", formattedMessage))
		Logger.Warn("Could not determine caller for diagnostic message %s", formattedMessage)
		return
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		v.diagnosticStack = append(v.diagnosticStack, fmt.Sprintf("%s:%d - %s", file, line, formattedMessage))
		Logger.Warn("Could not determine caller function for diagnostic message %s", formattedMessage)
		return
	}

	v.diagnosticStack = append(v.diagnosticStack, fmt.Sprintf("%s\n\t\t%s:%d%s", fn.Name(), file, line, formattedMessage))
}

func (v *ControllerVisitor) exit() {
	if !v.stackFrozen && len(v.diagnosticStack) > 0 {
		v.diagnosticStack = v.diagnosticStack[:len(v.diagnosticStack)-1]
	}
}

func (v *ControllerVisitor) getNextImportId() uint64 {
	value := v.importIdCounter
	v.importIdCounter++
	return value
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
		if existsInPackage == typeMeta.FullyQualifiedPackage {
			// Same type referenced from a separate location
			return nil
		}

		return v.getFrozenError(
			"type '%s' exists in more that one package (%s and %s). This is not currently supported",
			typeMeta.Name,
			typeMeta.FullyQualifiedPackage,
			existsInPackage,
		)
	}

	(*existingTypesMap)[typeMeta.Name] = typeMeta.FullyQualifiedPackage
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
		if param.TypeMeta.IsUniverseType && param.TypeMeta.Name == "error" && param.TypeMeta.FullyQualifiedPackage == "" {
			// Mark whether we've encountered any 'error' type
			plainErrorEncountered = true
		}
		err := v.addToTypeMap(existingTypesMap, existingModels, param.TypeMeta)
		if err != nil {
			return plainErrorEncountered, v.frozenError(err)
		}
	}

	for _, param := range route.Responses {
		if param.IsUniverseType && param.Name == "error" && param.FullyQualifiedPackage == "" {
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

	typeVisitor := NewTypeVisitor(v.packages)
	for _, model := range models {
		pkg := FilterPackageByFullName(v.packages, model.FullyQualifiedPackage)
		if pkg == nil {
			return nil, hasAnyErrorTypes, v.getFrozenError(
				"could locate packages.Package '%s' whilst looking for type '%s'.\n"+
					"Please note that Gleece currently cannot use any structs from externally imported packages",
				model.FullyQualifiedPackage,
				model.Name,
			)
		}

		structNode, err := FindTypesStructInPackage(pkg, model.Name)
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

func (v *ControllerVisitor) getFrozenError(format string, args ...any) error {
	v.stackFrozen = true
	return fmt.Errorf(format, args...)
}

func (v *ControllerVisitor) frozenError(err error) error {
	// Just a convenient way to freeze the diagnostic stack while returning the same error
	v.stackFrozen = true
	return err
}

func (v *ControllerVisitor) getSecurityFromContext(holder AttributesHolder) ([]definitions.RouteSecurity, error) {
	securities := []definitions.RouteSecurity{}

	normalSec := holder.GetAll(AttributeSecurity)
	if len(normalSec) > 0 {
		for _, secAttrib := range normalSec {
			schemaName := secAttrib.Value
			if len(schemaName) <= 0 {
				return securities, v.getFrozenError("a security schema's name cannot be empty")
			}

			definedScopes, err := GetCastProperty[[]string](secAttrib, PropertySecurityScopes)
			if err != nil {
				return securities, v.frozenError(err)
			}

			scopes := []string{}
			if definedScopes != nil && len(*definedScopes) > 0 {
				scopes = *definedScopes
			}

			securities = append(securities, definitions.RouteSecurity{
				SecurityAnnotation: []definitions.SecurityAnnotationComponent{{
					SchemaName: schemaName,
					Scopes:     scopes,
				}},
			})
		}
	}

	// AdvanceSecurity processing goes here

	return securities, nil
}

func (v *ControllerVisitor) createControllerMetadata(controllerNode *ast.TypeSpec) (definitions.ControllerMetadata, error) {
	fullPackageName, fullNameErr := GetFullPackageName(v.currentSourceFile, v.fileSet)
	packageAlias, aliasErr := GetDefaultPackageAlias(v.currentSourceFile)

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

	if commentSource != nil {
		comments := MapDocListToStrings(commentSource.List)
		holder, err := NewAttributeHolder(comments)
		if err != nil {
			return meta, v.frozenError(err)
		}

		security, err := v.getSecurityFromContext(holder)
		if err != nil {
			return meta, v.frozenError(err)
		}

		if len(security) <= 0 {
			Logger.Debug("Controller %s does not have explicit security; Using user-defined defaults", meta.Name)
			security = v.getDefaultSecurity()
		}

		meta.Tag = holder.GetFirstValueOrEmpty(AttributeTag)
		meta.Description = holder.GetFirstDescriptionOrEmpty(AttributeDescription)
		meta.RestMetadata = definitions.RestMetadata{Path: holder.GetFirstValueOrEmpty(AttributeRoute)}
		meta.Security = security
	}

	return meta, nil
}

func (v *ControllerVisitor) visitController(controllerNode *ast.TypeSpec) (definitions.ControllerMetadata, error) {
	v.enter(fmt.Sprintf("Controller '%s'", controllerNode.Name.Name))
	defer v.exit()

	controller, err := v.createControllerMetadata(controllerNode)
	v.currentController = &controller

	if err != nil {
		return controller, err
	}

	for _, file := range v.sourceFiles {
		for _, declaration := range file.Decls {
			switch funcDeclaration := declaration.(type) {
			case *ast.FuncDecl:
				if IsFuncDeclReceiverForStruct(controller.Name, funcDeclaration) {
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
	v.enter("")
	defer v.exit()

	funcParams := []definitions.FuncParam{}

	paramTypes, err := GetFuncParameterTypeList(v.currentSourceFile, v.fileSet, v.packages, funcDecl)
	if err != nil {
		return funcParams, err
	}

	for _, param := range paramTypes {
		// Record state for diagnostics
		v.enter(fmt.Sprintf("Param %s", param.Name))
		defer v.exit()

		holder, _ := NewAttributeHolder(comments)
		paramAttrib := holder.FindFirstByValue(param.Name)
		if paramAttrib == nil {
			return funcParams, v.getFrozenError("parameter '%s' does not have a matching documentation attribute", param.Name)
		}

		castValidator, err := GetCastProperty[string](paramAttrib, PropertyValidatorString)
		if err != nil {
			return funcParams, v.frozenError(err)
		}

		validatorString := ""
		if castValidator != nil && len(*castValidator) > 0 {
			validatorString = *castValidator
		}

		castName, err := GetCastProperty[string](paramAttrib, PropertyName)
		if err != nil {
			return funcParams, v.frozenError(err)
		}

		nameString := param.Name
		if castName != nil && len(*castName) > 0 {
			nameString = *castName
		}

		finalParamMeta := definitions.FuncParam{
			NameInSchema:       nameString,
			ParamMeta:          param,
			Description:        paramAttrib.Description,
			Validator:          appendParamRequiredValidation(&validatorString),
			UniqueImportSerial: v.getNextImportId(),
		}

		// Currently, only body param can be an object type
		switch strings.ToLower(paramAttrib.Name) {
		case "query":
			finalParamMeta.PassedIn = definitions.PassedInQuery
			if !finalParamMeta.TypeMeta.IsUniverseType {
				return funcParams, v.frozenError(createInvalidParamUsageError(finalParamMeta))
			}
		case "header":
			finalParamMeta.PassedIn = definitions.PassedInHeader
			if !finalParamMeta.TypeMeta.IsUniverseType {
				return funcParams, v.frozenError(createInvalidParamUsageError(finalParamMeta))
			}
		case "path":
			finalParamMeta.PassedIn = definitions.PassedInPath
			if !finalParamMeta.TypeMeta.IsUniverseType {
				return funcParams, v.frozenError(createInvalidParamUsageError(finalParamMeta))
			}
		case "body":
			finalParamMeta.PassedIn = definitions.PassedInBody
		}

		funcParams = append(funcParams, finalParamMeta)
	}

	return funcParams, nil
}

func (v ControllerVisitor) processPossibleErrorType(meta definitions.TypeMetadata) (bool, error) {
	v.enter(fmt.Sprintf("Return type %s (%s)", meta.Name, meta.FullyQualifiedPackage))
	defer v.exit()

	if meta.Name == "error" {
		return true, nil
	}

	pkg := FilterPackageByFullName(v.packages, meta.FullyQualifiedPackage)
	embeds, err := DoesStructEmbedType(pkg, meta.Name, "", "error")
	if err != nil {
		return false, err
	}

	return embeds, nil
}

func (v *ControllerVisitor) getFuncReturnValue(funcDecl *ast.FuncDecl) ([]definitions.FuncReturnValue, error) {
	v.enter("")
	defer v.exit()

	values := []definitions.FuncReturnValue{}
	var errorRetTypeIndex int

	returnTypes, err := GetFuncReturnTypeList(v.currentSourceFile, v.fileSet, v.packages, funcDecl)
	if err != nil {
		return values, err
	}

	// Note that controller methods must return and error or (any, error)

	switch len(returnTypes) {
	case 2:
		// If the method returns a 2-tuple, the error is expected to be the second value in the tuple
		errorRetTypeIndex = 1
	case 1:
		// If the method returns a single value, its expected to be an error
		errorRetTypeIndex = 0
	case 0:
		return values, v.getFrozenError("expected method to return an error or a value and error tuple but found void")
	default:
		typeNames := []string{}
		for _, typeMeta := range returnTypes {
			typeNames = append(typeNames, typeMeta.Name)
		}
		return values, v.getFrozenError(
			"expected method to return an error or a value and error tuple but found (%s)",
			strings.Join(typeNames, ", "),
		)
	}

	// Validate whether the method returns a proper error. This may be the first or second return type in the list
	retType := returnTypes[errorRetTypeIndex]
	isValidError, err := v.processPossibleErrorType(retType)
	if err != nil {
		return values, v.frozenError(err)
	}

	if !isValidError {
		return values, v.getFrozenError("return type '%s' expected to be an error or directly embed it", retType.Name)
	}

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

func (v ControllerVisitor) getMethodHideOpts(attributes *AttributesHolder) definitions.MethodHideOptions {
	attr := attributes.GetFirst(AttributeHidden)
	if attr == nil {
		// No '@Hidden' attribute
		return definitions.MethodHideOptions{Type: definitions.HideMethodNever}
	}

	if attr.Properties == nil || len(attr.Properties) <= 0 {
		return definitions.MethodHideOptions{Type: definitions.HideMethodAlways}
	}

	// Technically a bit redundant since we know by length whether there's a condition defined
	// but nothing stops user from adding text to the comment so this mostly serves as a validation
	if len(attr.Value) <= 0 {
		// Standard '@Hidden' attribute; Always hide.
		return definitions.MethodHideOptions{Type: definitions.HideMethodAlways}
	}

	// A '@Hidden(condition)' attribute
	return definitions.MethodHideOptions{Type: definitions.HideMethodCondition, Condition: attr.Value}
}

func (v ControllerVisitor) getDeprecationOpts(attributes *AttributesHolder) definitions.DeprecationOptions {
	attr := attributes.GetFirst(AttributeDeprecated)
	if attr == nil {
		return definitions.DeprecationOptions{Deprecated: false}
	}

	if len(attr.Description) <= 0 {
		// '@Deprecated' with no description
		return definitions.DeprecationOptions{Deprecated: true}
	}

	// '@Deprecated' with a comment
	return definitions.DeprecationOptions{Deprecated: true, Description: attr.Description}
}

func (v ControllerVisitor) getErrorResponseMetadata(attributes *AttributesHolder) ([]definitions.ErrorResponse, error) {
	responseAttributes := attributes.GetAll(AttributeErrorResponse)

	responses := []definitions.ErrorResponse{}
	encounteredCodes := MapSet.NewSet[external.HttpStatusCode]()

	for _, attr := range responseAttributes {
		code, err := definitions.ConvertToHttpStatus(attr.Value)
		if err != nil {
			return responses, err
		}

		if encounteredCodes.ContainsOne(code) {
			Logger.Warn(
				"Status code '%d' appears multiple time on a controller receiver. Ignoring. Original Comment: %s",
				code,
				attr,
			)
			continue
		}
		responses = append(responses, definitions.ErrorResponse{HttpStatusCode: code, Description: attr.Description})
		encounteredCodes.Add(code)
	}

	return responses, nil
}

func (v *ControllerVisitor) getResponseStatusCodeAndDescription(
	attributes *AttributesHolder,
	hasReturnValue bool,
) (external.HttpStatusCode, string, error) {
	// Set the success attrib code based on whether function returns a value or only error (200 vs 204)
	attrib := attributes.GetFirst(AttributeResponse)
	if attrib == nil {
		if hasReturnValue {
			return external.StatusOK, "", nil
		}

		return external.StatusNoContent, "", nil
	}

	var statusCode external.HttpStatusCode
	if len(attrib.Value) > 0 {
		code, err := definitions.ConvertToHttpStatus(attrib.Value)
		if err != nil {
			return 0, "", v.frozenError(err)
		}
		statusCode = code
	}

	return statusCode, attrib.Description, nil
}

func (v *ControllerVisitor) getRouteSecurityWithInheritance(attributes AttributesHolder) ([]definitions.RouteSecurity, error) {
	explicitSec, err := v.getSecurityFromContext(attributes)
	if err != nil {
		return []definitions.RouteSecurity{}, v.frozenError(err)
	}

	if len(explicitSec) > 0 {
		return explicitSec, nil
	}

	return v.currentController.Security, nil
}

func (v *ControllerVisitor) visitMethod(funcDecl *ast.FuncDecl) (definitions.RouteMetadata, bool, error) {
	v.enter(fmt.Sprintf("Method '%s'", funcDecl.Name.Name))
	defer v.exit()

	// Check whether there are any comments on the method - we expect all API endpoints to contain comments.
	// No comments - not an API endpoint.
	if funcDecl.Doc == nil || funcDecl.Doc.List == nil || len(funcDecl.Doc.List) <= 0 {
		return definitions.RouteMetadata{}, false, nil
	}

	comments := MapDocListToStrings(funcDecl.Doc.List)
	attributes, err := NewAttributeHolder(comments)
	if err != nil {
		return definitions.RouteMetadata{}, false, v.frozenError(err)
	}

	methodAttr := attributes.GetFirst(AttributeMethod)
	if methodAttr == nil {
		Logger.Info("Method '%s' does not have a @Method attribute and will be ignored", funcDecl.Name.Name)
		return definitions.RouteMetadata{}, false, nil
	}

	routePath := attributes.GetFirstPropertyValueOrEmpty(AttributeRoute)
	if len(routePath) <= 0 {
		Logger.Info("Method '%s' does not have an @Route attribute and will be ignored", funcDecl.Name.Name)
		return definitions.RouteMetadata{}, true, nil
	}

	errorResponses, err := v.getErrorResponseMetadata(&attributes)
	if err != nil {
		return definitions.RouteMetadata{}, true, v.frozenError(err)
	}

	security, err := v.getRouteSecurityWithInheritance(attributes)
	if err != nil {
		return definitions.RouteMetadata{}, true, v.frozenError(err)
	}

	if v.config.RoutesConfig.AuthorizationConfig.EnforceSecurityOnAllRoutes && len(security) <= 0 {
		return definitions.RouteMetadata{}, true, v.getFrozenError(
			"'enforceSecurityOnAllRoutes' setting is 'true'' but method '%s' on controller '%s'"+
				"does not have any explicit or implicit (inherited) security attributes",
			funcDecl.Name.Name,
			v.currentController.Name,
		)
	}

	meta := definitions.RouteMetadata{
		OperationId:         funcDecl.Name.Name,
		HttpVerb:            definitions.EnsureValidHttpVerb(methodAttr.Value),
		Description:         attributes.GetDescription(),
		Hiding:              v.getMethodHideOpts(&attributes),
		Deprecation:         v.getDeprecationOpts(&attributes),
		RestMetadata:        definitions.RestMetadata{Path: routePath},
		ErrorResponses:      errorResponses,
		RequestContentType:  definitions.ContentTypeJSON, // Hardcoded for now, should be supported via comments later
		ResponseContentType: definitions.ContentTypeJSON, // Hardcoded for now, should be supported via comments later
		Security:            security,
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
		return meta, true, v.frozenError(err)
	}
	meta.FuncParams = funcParams

	// Set the function's return types
	responses, err := v.getFuncReturnValue(funcDecl)
	if err != nil {
		return meta, true, v.frozenError(err)
	}
	meta.Responses = responses
	meta.HasReturnValue = len(responses) > 1

	successResponseCode, successDescription, err := v.getResponseStatusCodeAndDescription(&attributes, meta.HasReturnValue)
	if err != nil {
		return meta, true, v.frozenError(err)
	}
	meta.ResponseSuccessCode = successResponseCode
	meta.ResponseDescription = successDescription

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

// For now, all params are required, later we will support nil for pointers and slices params
func appendParamRequiredValidation(validation *string) string {
	if validation == nil || *validation == "" {
		return "required"
	}

	// Split the validation string into individual tags
	tags := strings.Split(*validation, ",")

	// Check if "required" is already present
	for _, tag := range tags {
		if tag == "required" {
			return *validation
		}
	}

	// Append "required" to the validation string
	return *validation + ",required"
}
