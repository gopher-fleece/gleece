package extractor

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	MapSet "github.com/deckarep/golang-set/v2"

	"github.com/haimkastner/gleece/definitions"
	"github.com/haimkastner/gleece/external"
	Logger "github.com/haimkastner/gleece/infrastructure/logger"
)

func ExtractClassMetadata(d ast.GenDecl, baseStruct string) (*definitions.ControllerMetadata, error) {
	for _, spec := range d.Specs {
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				if EmbedsBaseStruct(structType, baseStruct) {
					// Initialize metadata for this controller
					ctrlMetadata := definitions.ControllerMetadata{
						Name: typeSpec.Name.Name,
					}

					// Extract comments for tags (e.g., @Tag)
					if d.Doc != nil {
						comments := MapDocListToStrings(d.Doc.List)
						ctrlMetadata.Tag = FindAndExtract(comments, "@Tag")
						ctrlMetadata.Description = FindAndExtract(comments, "@Description")
						ctrlMetadata.RestMetadata = BuildRestMetadata(comments)
					}

					return &ctrlMetadata, nil
				}
			}
		}
	}
	return nil, nil
}

func parseErrorResponseComment(comment string) definitions.ErrorResponse {
	httpCodeMatch := regexp.MustCompile(`^\d{3,3}`)
	if !httpCodeMatch.Match([]byte(comment)) {
		panic(fmt.Sprintf("ErrorResponse annotations must start with an HTTP status code. Received value: '%s'", comment))
	}

	statusCodeUint, err := strconv.ParseUint(comment[:3], 10, 32)
	if err != nil {
		panic(fmt.Sprintf("Could not parse ErrorResponse HTTP Code value '%s' - %v", comment[:3], err))
	}

	statusCode := definitions.EnsureHttpStatusCode(uint(statusCodeUint))
	return definitions.ErrorResponse{
		HttpStatusCode: statusCode,
		Description:    comment[3:],
	}
}

func getErrorResponseMetadata(comments []string) []definitions.ErrorResponse {
	errorResponses := FindAndExtractOccurrences(comments, "@ErrorResponse", 0)
	responses := []definitions.ErrorResponse{}
	encounteredCodes := MapSet.NewSet[external.HttpStatusCode]()

	for _, errorResponseComment := range errorResponses {
		response := parseErrorResponseComment(errorResponseComment)
		if encounteredCodes.ContainsOne(response.HttpStatusCode) {
			Logger.Warn(
				"Status code '%d' appears multiple time on a controller receiver. Ignoring. Original Comment: %s",
				response.HttpStatusCode,
				errorResponseComment,
			)
			continue
		}
		responses = append(responses, response)
		encounteredCodes.Add(response.HttpStatusCode)
	}
	return responses
}

func getExpressionName(expression ast.Expr) string {
	switch e := expression.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", e.X, e.Sel.Name)
	default:
		panic(fmt.Sprintf("Unsupported expression type '%v'", e))
	}
}

func getRouteSuccessResponseCode(comments []string, routeHasResultValue bool) external.HttpStatusCode {
	responseCode := FindAndExtract(comments, "@ResponseCode")

	if responseCode != "" {
		return definitions.EnsureHttpStatusCodeString(responseCode)
	}
	if routeHasResultValue {
		return external.StatusOK
	}

	return external.StatusNoContent
}

func getRouteParameters(comments []string, routeFuncDecl ast.FuncDecl) []definitions.FuncParamLegacy {
	params := []definitions.FuncParamLegacy{}

	for _, field := range routeFuncDecl.Type.Params.List {
		for _, name := range field.Names {
			param := definitions.FuncParamLegacy{
				Name: name.Name,
			}
			line := SearchForParamTerm(comments, name.Name)
			if line == "" {
				Logger.Warn("No line found for ", name.Name)
				continue
			}

			// Check whenever the actual name in the HTTP request is different from the function parameter name
			if httpName := ExtractParenthesesContent(line); httpName != "" {
				param.Name = httpName
			}

			if pType := strings.ToLower(ExtractParamTerm(line)); pType != "" {
				switch pType {
				case "query":
					param.ParamType = definitions.PassedInQuery
				case "header":
					param.ParamType = definitions.PassedInHeader
				case "path":
					param.ParamType = definitions.PassedInPath
				case "body":
					param.ParamType = definitions.PassedInBody
				}
			}

			// Extract the rest of the line as the description
			param.Description = strings.TrimSpace(GetTextAfterParenthesis(line, " "+name.Name+" "))

			// NOTE:
			// This takes the qualified name - it WILL cause problems if the import is renamed in the controller
			// i.e., the inferred package name may be different than the actual one
			param.ParamExpressionName = getExpressionName(field.Type)
			params = append(params, param)
		}
	}
	return params
}

func ExtractClassRoutesMetaData(routeFuncDecl ast.FuncDecl) (*definitions.RouteMetadata, error) {
	routeMetadata := definitions.RouteMetadata{}
	comments := MapDocListToStrings(routeFuncDecl.Doc.List)

	routeMetadata.OperationId = routeFuncDecl.Name.Name
	routeMetadata.HttpVerb = definitions.EnsureValidHttpVerb(FindAndExtract(comments, "@Method"))
	routeMetadata.Description = FindAndExtract(comments, "@Description")
	routeMetadata.RestMetadata = BuildRestMetadata(comments)
	// routeMetadata.Response = getResponseInterface(routeFuncDecl)
	// routeMetadata.ResponseSuccessCode = getRouteSuccessResponseCode(comments, routeMetadata.Response.InterfaceName != "")
	routeMetadata.ErrorResponses = getErrorResponseMetadata(comments)

	// Extract function parameters
	if routeFuncDecl.Type.Params != nil {
		//	routeMetadata.FuncParams = getRouteParameters(comments, routeFuncDecl)
	}

	// Extract function results
	if routeFuncDecl.Type.Results != nil {
	}

	return &routeMetadata, nil
}

func GetMetadata(codeFileGlobs ...string) ([]definitions.ControllerMetadata, error) {
	var globs []string
	if len(codeFileGlobs) > 0 {
		globs = codeFileGlobs
	} else {
		globs = []string{"./*.go", "./**/*.go"}
	}

	visitor := &ControllerVisitor{}
	visitor.Init(globs)
	for _, file := range visitor.GetFiles() {
		ast.Walk(visitor, file)
	}

	lastErr := visitor.GetLastError()
	if lastErr != nil {
		Logger.Error("Visitor encountered at-least one error. Last error - %v", *lastErr)
		return []definitions.ControllerMetadata{}, *lastErr
	}
	controllers, _ := visitor.DumpContext()
	Logger.Info("%v", controllers)
	return visitor.GetControllers(), nil
}

func ExtractMetadata() ([]definitions.ControllerMetadata, error) {
	GetMetadata()

	// Define the directory containing the Go files
	dir := "./ctrl"

	// Define the name of the base struct we are looking for
	baseStruct := "GleeceController"

	// Array to hold metadata
	var metadata []definitions.ControllerMetadata

	// Iterate over all Go files in the directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		// Parse each Go file
		fileSet := token.NewFileSet()
		node, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Printf("Error parsing file %s: %v\n", path, err)
			return err
		}

		// Use FilterDecls to get all GenDecls
		allGenDecls := FilterDecls(node.Decls, func(d ast.Decl) bool {
			_, ok := d.(*ast.GenDecl)
			return ok
		})

		// Filter only the GenDecls that inherits from the base struct name "GleeceController"
		genDecls := FilterDecls(allGenDecls, func(d ast.Decl) bool {
			genDecl := d.(*ast.GenDecl)
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structType := typeSpec.Type.(*ast.StructType)
				if EmbedsBaseStruct(structType, baseStruct) {
					return true
				}
			}
			return false
		})

		if genDecls == nil {
			return nil
		}

		allFuncDecl := FilterDecls(node.Decls, func(d ast.Decl) bool {
			_, ok := d.(*ast.FuncDecl)
			return ok
		})

		// Print the GenDecls
		for _, decl := range genDecls {

			// Get all funcDecl of current struct/class only
			funcDecl := FilterDecls(allFuncDecl, func(d ast.Decl) bool {
				funcDecl := d.(*ast.FuncDecl)
				if funcDecl.Recv != nil {
					receiver := funcDecl.Recv.List[0]
					if starExpr, ok := receiver.Type.(*ast.StarExpr); ok {
						if ident, ok := starExpr.X.(*ast.Ident); ok {
							if ident.Name == decl.(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Name.Name {
								return true
							}
						}
					}
				}
				return false
			})

			if funcDecl == nil {
				continue
			}

			// Print the class name from Decl and all functions names from filtered funcDecl
			Logger.Debug(decl.(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Name.Name)
			for _, f := range funcDecl {
				Logger.Debug(f.(*ast.FuncDecl).Name.Name)
			}

			metadataInfo, _ := ExtractClassMetadata(*decl.(*ast.GenDecl), baseStruct)
			ctrlMetadata := definitions.ControllerMetadata{}
			ctrlMetadata.Package = node.Name.Name
			ctrlMetadata.Name = metadataInfo.Name
			ctrlMetadata.Tag = metadataInfo.Tag
			ctrlMetadata.Description = metadataInfo.Description
			ctrlMetadata.RestMetadata = metadataInfo.RestMetadata

			ctrlMetadata.Routes = []definitions.RouteMetadata{}
			for _, f := range funcDecl {
				routeMetadata, _ := ExtractClassRoutesMetaData(*f.(*ast.FuncDecl))
				ctrlMetadata.Routes = append(ctrlMetadata.Routes, *routeMetadata)
			}

			metadata = append(metadata, ctrlMetadata)

		}

		// Walk through the AST and extract relevant information
		//for _, decl := range node.Decls {
		//	switch d := decl.(type) {
		//	case *ast.GenDecl: // For types like structs
		//		metadataInfo, _ := ExtractClassMetadata(*d, baseStruct)
		//		if metadataInfo == nil {
		//			break
		//		}
		//		ctrlMetadata.CtrlName = metadataInfo.CtrlName
		//		ctrlMetadata.CtrlTag = metadataInfo.CtrlTag
		//		ctrlMetadata.CtrlDescription = metadataInfo.CtrlDescription
		//		ctrlMetadata.CtrlRestMetadata = metadataInfo.CtrlRestMetadata
		//	case *ast.FuncDecl: // For functions and methods
		//		ctrlMetadata.Routes, _ = ExtractClassRoutesMetaData(*d, "sdfdf")
		//		//if d.Recv != nil { // Check if it’s a method
		//		//	receiver := d.Recv.List[0]
		//		//	if starExpr, ok := receiver.Type.(*ast.StarExpr); ok {
		//		//		if ident, ok := starExpr.X.(*ast.Ident); ok {
		//		//			// Find the corresponding controller in metadata
		//		//			for i, ctrl := range metadata {
		//		//				if ctrl.CtrlName == ident.Name {
		//		//					// Extract route metadata
		//		//					if d.Doc != nil {
		//		//						for _, comment := range d.Doc.List {
		//		//							commentText := strings.TrimSpace(comment.Text)
		//		//							if strings.HasPrefix(commentText, "// @Route") {
		//		//								//routePath := strings.TrimPrefix(commentText, "// @Route ")
		//		//								//metadata[i].CtrlRestMetadata = append(metadata[i].CtrlRestMetadata, RestMetadata{
		//		//								//	Path: routePath,
		//		//								//})
		//		//							}
		//		//						}
		//		//					}
		//		//				}
		//		//			}
		//		//		}
		//		//	}
		//		//}
		//	}
		//}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func main() {

}
