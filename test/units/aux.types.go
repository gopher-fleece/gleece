package units

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
