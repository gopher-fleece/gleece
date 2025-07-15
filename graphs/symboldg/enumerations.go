package symboldg

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

	// Special case
	PrimitiveTypeUnsafePointer PrimitiveType = "unsafe.Pointer"
)
