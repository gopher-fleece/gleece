package metadata

import (
	"github.com/gopher-fleece/gleece/graphs"
)

type IdProvider interface {
	GetIdForKey(key graphs.SymbolKey) uint64
}

// Go's insane package system forces us to get... creative.
type MetaCache interface {
	GetStruct(key graphs.SymbolKey) *StructMeta
	GetEnum(key graphs.SymbolKey) *EnumMeta
}
