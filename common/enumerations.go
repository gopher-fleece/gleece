package common

type SymKind string

const (
	SymKindUnknown        SymKind = "Unknown"
	SymKindPackage        SymKind = "Package"
	SymKindStruct         SymKind = "Struct"
	SymKindController     SymKind = "Controller"
	SymKindInterface      SymKind = "Interface"
	SymKindAlias          SymKind = "Alias"
	SymKindComposite      SymKind = "Composite"
	SymKindTypeParam      SymKind = "TypeParam"
	SymKindEnum           SymKind = "Enum"
	SymKindEnumValue      SymKind = "EnumValue"
	SymKindFunction       SymKind = "Function"
	SymKindReceiver       SymKind = "Receiver"
	SymKindField          SymKind = "Field"
	SymKindParameter      SymKind = "Parameter"
	SymKindVariable       SymKind = "Variable"
	SymKindConstant       SymKind = "Constant"
	SymKindReturnType     SymKind = "RetType"
	SymKindBuiltin        SymKind = "Builtin"
	SymKindSpecialBuiltin SymKind = "Special"
)

// IsBuiltin returns a boolean indicating whether the SymKind is a 'built-in' type, such as 'string' or 'error'.
// Note that this includes both SymKindBuiltin and SymKindSpecialBuiltin kinds
func (k SymKind) IsBuiltin() bool {
	switch k {
	case SymKindBuiltin, SymKindSpecialBuiltin:
		return true
	default:
		return false
	}
}

type ImportType string

const (
	ImportTypeNone  ImportType = "None"
	ImportTypeAlias ImportType = "Alias"
	ImportTypeDot   ImportType = "Dot"
)

