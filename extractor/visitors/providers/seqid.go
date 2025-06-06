package providers

import (
	"sync/atomic"
)

type SyncedProvider struct {
	nextImportId uint64
}

func (p *SyncedProvider) GetNextImportId() uint64 {
	return atomic.AddUint64(&p.nextImportId, 1) - 1
}
