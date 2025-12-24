package providers

import (
	"sync/atomic"

	"github.com/gopher-fleece/gleece/v2/graphs"
)

type SyncedProvider struct {
	nextImportId  uint64
	keyToImportId map[graphs.SymbolKey]uint64
}

func NewSyncedProvider() SyncedProvider {
	return SyncedProvider{
		keyToImportId: make(map[graphs.SymbolKey]uint64),
	}
}

func (p *SyncedProvider) GetNextImportId() uint64 {
	return atomic.AddUint64(&p.nextImportId, 1) - 1
}

func (p *SyncedProvider) GetIdForKey(key graphs.SymbolKey) uint64 {
	id, exists := p.keyToImportId[key]
	if exists {
		return id
	}

	id = atomic.AddUint64(&p.nextImportId, 1) - 1
	p.keyToImportId[key] = id
	return id
}
