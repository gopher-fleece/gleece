package metadata

import (
	"fmt"
	"go/types"

	"github.com/gopher-fleece/gleece/common/linq"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/definitions"
)

type EnumValueKind string

const (
	EnumValueKindString  EnumValueKind = "string"
	EnumValueKindInt     EnumValueKind = "int"
	EnumValueKindInt8    EnumValueKind = "int8"
	EnumValueKindInt16   EnumValueKind = "int16"
	EnumValueKindInt32   EnumValueKind = "int32"
	EnumValueKindInt64   EnumValueKind = "int64"
	EnumValueKindUInt    EnumValueKind = "uint"
	EnumValueKindUInt8   EnumValueKind = "uint8"
	EnumValueKindUInt16  EnumValueKind = "uint16"
	EnumValueKindUInt32  EnumValueKind = "uint32"
	EnumValueKindUInt64  EnumValueKind = "uint64"
	EnumValueKindFloat32 EnumValueKind = "float32"
	EnumValueKindFloat64 EnumValueKind = "float64"
	EnumValueKindBool    EnumValueKind = "bool"
)

func NewEnumValueKind(kind types.BasicKind) (EnumValueKind, error) {
	switch kind {
	case types.String:
		return EnumValueKindString, nil
	case types.Int:
		return EnumValueKindInt, nil
	case types.Int8:
		return EnumValueKindInt8, nil
	case types.Int16:
		return EnumValueKindInt16, nil
	case types.Int32:
		return EnumValueKindInt32, nil
	case types.Int64:
		return EnumValueKindInt64, nil
	case types.Uint:
		return EnumValueKindUInt, nil
	case types.Uint8:
		return EnumValueKindUInt8, nil
	case types.Uint16:
		return EnumValueKindUInt16, nil
	case types.Uint32:
		return EnumValueKindUInt32, nil
	case types.Uint64:
		return EnumValueKindUInt64, nil
	case types.Float32:
		return EnumValueKindFloat32, nil
	case types.Float64:
		return EnumValueKindFloat64, nil
	case types.Bool:
		return EnumValueKindBool, nil
	default:
		return "", fmt.Errorf("unsupported basic kind: %v", kind)
	}
}

type EnumValueDefinition struct {
	// The enum's value definition node meta, e.g. EnumValueA SomeEnumType = "Abc"
	SymNodeMeta
	Value any // e.g. ["Meter", "Kilometer"]
	// TODO: An exact textual representation of the value. For example "1 << 2"
	//RawLiteralValue string
}

type EnumMeta struct {
	// The enum's type definition's node meta e.g. type SomeEnumType string
	SymNodeMeta
	ValueKind EnumValueKind // e.g. string, int, etc.
	Values    []EnumValueDefinition
}

func (e EnumMeta) Reduce(_ ReductionContext) (definitions.EnumMetadata, error) {
	stringifiedValues := linq.Map(e.Values, func(value EnumValueDefinition) string {
		return fmt.Sprintf("%v", value.Value)
	})

	return definitions.EnumMetadata{
		Name:        e.Name,
		PkgPath:     e.PkgPath,
		Description: annotations.GetDescription(e.Annotations),
		Values:      stringifiedValues,
		Type:        string(e.ValueKind),
		Deprecation: GetDeprecationOpts(e.Annotations),
	}, nil
}
