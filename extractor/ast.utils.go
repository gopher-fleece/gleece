package extractor

import "go/ast"

func IsFuncDeclReceiverForStruct(structName string, funcDecl *ast.FuncDecl) bool {
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) <= 0 {
		return false
	}

	switch expr := funcDecl.Recv.List[0].Type.(type) {
	case *ast.Ident:
		return expr.Name == structName
	case *ast.StarExpr:
		return expr.X.(*ast.Ident).Name == structName
	default:
		return false
	}
}

func DoesStructEmbedStruct(structNode *ast.StructType, embeddedStructName string) bool {
	for _, field := range structNode.Fields.List {
		if ident, isOk := field.Type.(*ast.Ident); isOk {
			if ident.Name == embeddedStructName {
				return true
			}
		}
	}
	return false
}
