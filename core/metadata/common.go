package metadata

import (
	"go/ast"
	"go/token"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
)

type IdProvider interface {
	GetIdForKey(key graphs.SymbolKey) uint64
}

// Go's insane package system forces us to get... creative.
type MetaCache interface {
	GetStruct(key graphs.SymbolKey) *StructMeta
	GetReceiver(key graphs.SymbolKey) *ReceiverMeta
	GetEnum(key graphs.SymbolKey) *EnumMeta
	GetAlias(key graphs.SymbolKey) *AliasMeta
	HasController(meta *ControllerMeta) bool
	HasReceiver(key graphs.SymbolKey) bool
	HasStruct(key graphs.SymbolKey) bool
	HasEnum(key graphs.SymbolKey) bool
	HasAlias(key graphs.SymbolKey) bool
	HasVisited(key graphs.SymbolKey) bool
	GetFileVersion(file *ast.File, fileSet *token.FileSet) (*gast.FileVersion, error)
	AddController(meta *ControllerMeta) error
	AddReceiver(meta *ReceiverMeta) error
	AddStruct(meta *StructMeta) error
	AddEnum(meta *EnumMeta) error
	AddAlias(meta *AliasMeta) error

	// StartMaterializing claims the key for materialization.
	// Returns true if the caller should proceed (not visited, not already in-progress).
	StartMaterializing(key graphs.SymbolKey) bool

	// FinishMaterializing clears in-progress. If success==true, also mark visited.
	FinishMaterializing(key graphs.SymbolKey, success bool)
}

type ReductionContext struct {
	GleeceConfig   *definitions.GleeceConfig
	MetaCache      MetaCache
	SyncedProvider IdProvider
}
