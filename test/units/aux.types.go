package units

import "go/ast"

type StructA struct {
}

type InterfaceA interface {
	FuncA()
}

type EnumTypeA string

const (
	EnumValueA EnumTypeA = "A"
	EnumValueB EnumTypeA = "B"
)

type AliasTypeA = string

const ConstA = 1

type IntAlias int

type StructForGetUnderlyingTypeName struct {
	FieldIntAlias     IntAlias
	FieldStringPtr    *string
	FieldIntSlice     []int
	FieldIntArray     [3]int
	FieldStringIntMap map[string]int
	FieldChannelInt   chan int
	FieldFunc         func()
	FieldInterface    any
	FieldStruct       struct{}
	FieldInt          (int)
	FieldComment      ast.Comment
}

func SimpleVariadicFunc(...int) {}

func NotAReceiver() {}

type StructWithReceivers struct{}

func (s StructWithReceivers) ValueReceiverForStructWithReceivers()    {}
func (s *StructWithReceivers) PointerReceiverForStructWithReceivers() {}
