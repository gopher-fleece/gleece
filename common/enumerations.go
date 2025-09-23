package common

import "fmt"

type SymKind string

const (
	SymKindUnknown        SymKind = "Unknown"
	SymKindPackage        SymKind = "Package"
	SymKindStruct         SymKind = "Struct"
	SymKindController     SymKind = "Controller"
	SymKindInterface      SymKind = "Interface"
	SymKindAlias          SymKind = "Alias"
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

func ToSymbolKind(value string) (SymKind, error) {
	switch value {
	case string(SymKindUnknown):
		return SymKindUnknown, nil
	case string(SymKindPackage):
		return SymKindPackage, nil
	case string(SymKindStruct):
		return SymKindStruct, nil
	case string(SymKindController):
		return SymKindController, nil
	case string(SymKindInterface):
		return SymKindInterface, nil
	case string(SymKindAlias):
		return SymKindAlias, nil
	case string(SymKindEnum):
		return SymKindEnum, nil
	case string(SymKindEnumValue):
		return SymKindEnumValue, nil
	case string(SymKindFunction):
		return SymKindFunction, nil
	case string(SymKindReceiver):
		return SymKindReceiver, nil
	case string(SymKindField):
		return SymKindField, nil
	case string(SymKindParameter):
		return SymKindParameter, nil
	case string(SymKindVariable):
		return SymKindVariable, nil
	case string(SymKindConstant):
		return SymKindConstant, nil
	case string(SymKindReturnType):
		return SymKindReturnType, nil
	case string(SymKindBuiltin):
		return SymKindBuiltin, nil
	case string(SymKindSpecialBuiltin):
		return SymKindSpecialBuiltin, nil
	default:
		return "", fmt.Errorf("invalid SymKind: %q", value)
	}
}

type ImportType string

const (
	ImportTypeNone  ImportType = "None"
	ImportTypeAlias ImportType = "Alias"
	ImportTypeDot   ImportType = "Dot"
)
