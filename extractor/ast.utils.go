package extractor

import "go/ast"

func IsFuncDeclReceiverForStruct(structName string, funcDecl *ast.FuncDecl) bool {
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) <= 0 {
		return false
	}

	receiver := funcDecl.Recv.List[0].Type
	ident, ok := receiver.(*ast.Ident)
	if ok && ident.Name == structName {
		return true
	}

	if starExpr, ok := receiver.(*ast.StarExpr); ok {
		if ident, ok := starExpr.X.(*ast.Ident); ok {
			if ident.Name == structName {
				return true
			}
		}
	}
	return false
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
