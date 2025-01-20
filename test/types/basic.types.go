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
	FieldA string
	FieldB uint64
}

type SomeNestedStruct struct {
	FieldA int
	FieldB SomeVeryNestedStruct
}
type HoldsVeryNestedStructs struct {
	FieldA float32
	FieldB uint
	FieldC SomeNestedStruct
}
