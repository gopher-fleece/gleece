package common

type SymKind string

const (
	SymKindUnknown   SymKind = "Unknown"
	SymKindPackage   SymKind = "Package"
	SymKindStruct    SymKind = "Struct"
	SymKindInterface SymKind = "Interface"
	SymKindAlias     SymKind = "Alias"
	SymKindFunction  SymKind = "Function"
	SymKindMethod    SymKind = "Method"
	SymKindField     SymKind = "Field"
	SymKindParameter SymKind = "Parameter"
	SymKindVariable  SymKind = "Variable"
	SymKindConstant  SymKind = "Constant"
)

type AstNodeKind string

const (
	AstNodeKindUnknown     AstNodeKind = "Unknown"
	AstNodeKindInterface   AstNodeKind = "Interface"
	AstNodeKindStruct      AstNodeKind = "Struct"
	AstNodeKindIdent       AstNodeKind = "Identifier"
	AstNodeKindSelector    AstNodeKind = "SelectorExpr"
	AstNodeKindPointer     AstNodeKind = "Pointer"
	AstNodeKindArray       AstNodeKind = "Array"
	AstNodeKindMap         AstNodeKind = "Map"
	AstNodeKindChannel     AstNodeKind = "Channel"
	AstNodeKindFunction    AstNodeKind = "Function"
	AstNodeKindVariadic    AstNodeKind = "Variadic"
	AstNodeKindParenthesis AstNodeKind = "Parenthesis"
)
