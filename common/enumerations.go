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

type PrimitiveType string

const (
	PrimitiveTypeBool   PrimitiveType = "bool"
	PrimitiveTypeString PrimitiveType = "string"

	// Signed integers
	PrimitiveTypeInt   PrimitiveType = "int"
	PrimitiveTypeInt8  PrimitiveType = "int8"
	PrimitiveTypeInt16 PrimitiveType = "int16"
	PrimitiveTypeInt32 PrimitiveType = "int32"
	PrimitiveTypeInt64 PrimitiveType = "int64"

	// Unsigned integers
	PrimitiveTypeUint    PrimitiveType = "uint"
	PrimitiveTypeUint8   PrimitiveType = "uint8"
	PrimitiveTypeUint16  PrimitiveType = "uint16"
	PrimitiveTypeUint32  PrimitiveType = "uint32"
	PrimitiveTypeUint64  PrimitiveType = "uint64"
	PrimitiveTypeUintptr PrimitiveType = "uintptr"

	// Aliases
	PrimitiveTypeByte PrimitiveType = "byte" // alias for uint8
	PrimitiveTypeRune PrimitiveType = "rune" // alias for int32

	// Floating point numbers
	PrimitiveTypeFloat32 PrimitiveType = "float32"
	PrimitiveTypeFloat64 PrimitiveType = "float64"

	// Complex numbers
	PrimitiveTypeComplex64  PrimitiveType = "complex64"
	PrimitiveTypeComplex128 PrimitiveType = "complex128"
)

// ToPrimitiveType checks if the given string represents a valid PrimitiveType.
// If it does, it returns (corresponding PrimitiveType, true).
func ToPrimitiveType(typeName string) (PrimitiveType, bool) {
	switch typeName {
	case
		string(PrimitiveTypeBool),
		string(PrimitiveTypeString),

		string(PrimitiveTypeInt),
		string(PrimitiveTypeInt8),
		string(PrimitiveTypeInt16),
		string(PrimitiveTypeInt32),
		string(PrimitiveTypeInt64),

		string(PrimitiveTypeUint),
		string(PrimitiveTypeUint8),
		string(PrimitiveTypeUint16),
		string(PrimitiveTypeUint32),
		string(PrimitiveTypeUint64),
		string(PrimitiveTypeUintptr),

		string(PrimitiveTypeByte),
		string(PrimitiveTypeRune),

		string(PrimitiveTypeFloat32),
		string(PrimitiveTypeFloat64),

		string(PrimitiveTypeComplex64),
		string(PrimitiveTypeComplex128):
		return PrimitiveType(typeName), true
	default:
		return "", false
	}
}

type SpecialType string

const (
	SpecialTypeError          SpecialType = "error"
	SpecialTypeEmptyInterface SpecialType = "interface{}"
	SpecialTypeContext        SpecialType = "context.Context"
	SpecialTypeTime           SpecialType = "time.Time"
	SpecialTypeAny            SpecialType = "any" // alias of interface{}
	SpecialTypeUnsafePointer  SpecialType = "unsafe.Pointer"
)

func (s SpecialType) IsUniverse() bool {
	switch s {
	case SpecialTypeError, SpecialTypeEmptyInterface, SpecialTypeAny:
		return true
	}

	return false
}

func ToSpecialType(s string) (SpecialType, bool) {
	switch s {
	case "error":
		return SpecialTypeError, true
	case "interface{}":
		return SpecialTypeEmptyInterface, true
	case "any":
		return SpecialTypeAny, true
	case "context.Context":
		return SpecialTypeContext, true
	case "time.Time":
		return SpecialTypeTime, true
	case "unsafe.Pointer":
		return SpecialTypeUnsafePointer, true
	default:
		return "", false
	}
}
