package extractor

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/haimkastner/gleece/definitions"
	. "github.com/haimkastner/gleece/definitions"
)

func ExtractClassMetadata(d ast.GenDecl, baseStruct string) (*ControllerMetadata, error) {
	for _, spec := range d.Specs {
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				if EmbedsBaseStruct(structType, baseStruct) {
					// Initialize metadata for this controller
					ctrlMetadata := ControllerMetadata{
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

func ExtractFuncReturnTypes(d ast.FuncDecl) (string, error) {
	var returnTypes []string

	// Check if the function has results (return values)
	if d.Type.Results != nil {
		for _, field := range d.Type.Results.List {
			if ident, ok := field.Type.(*ast.Ident); ok {
				returnTypes = append(returnTypes, ident.Name)
			}
		}
	}

	// If it's empty array, print error
	if len(returnTypes) == 0 {
		println("No return types found for ", d.Name.Name)
		return "", nil
	}

	// If it's more then 2, print error, no more them tow are supported
	if len(returnTypes) > 2 {
		println("More then 2 return types found for ", d.Name.Name)
		return "", nil
	}

	if len(returnTypes) == 1 {

		// Validate that the only return type is not an error
		if returnTypes[0] != "error" {
			println("No error return type found for ", d.Name.Name)
			return "", nil
		}
		return "void", nil
	}

	// Validate that the first return type is NOT an error
	if returnTypes[0] == "error" {
		println("No error return type found for ", d.Name.Name)
		return "", nil
	}

	// Validate that the second return type is an error
	if returnTypes[1] != "error" {
		println("No error return type found for ", d.Name.Name)
		return "", nil
	}

	return returnTypes[0], nil
}

func ExtractClassRoutesMetaData(d ast.FuncDecl) (*RouteMetadata, error) {
	routeMetadata := RouteMetadata{}
	routeMetadata.OperationId = d.Name.Name
	comments := MapDocListToStrings(d.Doc.List)

	routeMetadata.HttpVerb = EnsureValidHttpVerb(FindAndExtract(comments, "@Method"))
	routeMetadata.Description = FindAndExtract(comments, "@Description")
	routeMetadata.RestMetadata = BuildRestMetadata(comments)
	routeMetadata.ResponseInterface, _ = ExtractFuncReturnTypes(d)

	responseCode := FindAndExtract(comments, "@ResponseCode")

	if responseCode != "" {
		routeMetadata.ResponseSuccessCode = responseCode
	} else if routeMetadata.ResponseInterface != "void" {
		routeMetadata.ResponseSuccessCode = "200"
	} else {
		routeMetadata.ResponseSuccessCode = "204"
	}

	// Extract function parameters
	if d.Type.Params != nil {
		for _, field := range d.Type.Params.List {
			for _, name := range field.Names {
				param := FuncParam{
					Name: name.Name,
				}
				line := SearchForParamTerm(comments, name.Name)
				if line == "" {
					println("No line found for ", name.Name)
					continue
				}

				// Check whenever the actual name in the HTTP request is different from the function parameter name
				if httpName := ExtractParenthesesContent(line); httpName != "" {
					param.Name = httpName
				}

				if pType := strings.ToLower(ExtractParamTerm(line)); pType != "" {
					switch pType {
					case "query":
						param.ParamType = definitions.Query
					case "header":
						param.ParamType = definitions.Header
					case "path":
						param.ParamType = definitions.Path
					case "body":
						param.ParamType = definitions.Body
					}
				}

				// Extract the rest of the line as the description
				param.Description = strings.TrimSpace(GetTextAfterParenthesis(line, " "+name.Name+" "))

				param.ParamInterface = field.Type.(*ast.Ident).Name
				routeMetadata.FuncParams = append(routeMetadata.FuncParams, param)
			}
		}
	}

	// Extract function results
	if d.Type.Results != nil {

	}
	return &routeMetadata, nil
}

func ExtractMetadata() ([]ControllerMetadata, error) {
	// Define the directory containing the Go files
	dir := "./ctrl"

	// Define the name of the base struct we are looking for
	baseStruct := "GleeceController"

	// Array to hold metadata
	var metadata []ControllerMetadata

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
			fmt.Println(decl.(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Name.Name)
			for _, f := range funcDecl {
				fmt.Println(f.(*ast.FuncDecl).Name.Name)
			}

			metadataInfo, _ := ExtractClassMetadata(*decl.(*ast.GenDecl), baseStruct)
			ctrlMetadata := ControllerMetadata{}
			ctrlMetadata.Package = node.Name.Name
			ctrlMetadata.Name = metadataInfo.Name
			ctrlMetadata.Tag = metadataInfo.Tag
			ctrlMetadata.Description = metadataInfo.Description
			ctrlMetadata.RestMetadata = metadataInfo.RestMetadata

			ctrlMetadata.Routes = []RouteMetadata{}
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
