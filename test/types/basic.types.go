package types

type StringAlias string

type ImportedWithDefaultAlias struct {
	FieldA uint
	FieldB string
}
type ImportedWithCustomAlias struct {
	FieldA uint
	FieldB string
}
type ImportedWithDot struct {
	FieldA uint
	FieldB string
}

type SomeVeryNestedStruct struct {
	// This is a very nested field A
	FieldA string
	FieldB uint64
}

type SomeNestedStruct struct {
	FieldA int
	// This is field B
	FieldB SomeVeryNestedStruct
}

// This struct holds some nested structs
// @Deprecated
// This comment should not be included in the description
type HoldsVeryNestedStructs struct {
	FieldA float32
	FieldB uint `json:"fieldB" validator:"required"`
	FieldC SomeNestedStruct
}
