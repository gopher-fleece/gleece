package state

import (
	"sync"
	"time"
)

// Session holds per-client state. Keep it minimal to start.
type Session struct {
	mu      sync.Mutex
	buffers map[string]string // uri -> full text
	backend Backend           // interface to your core (see below)
	// you can add cancellation contexts, debounce timers etc.
	lastDiag time.Time
}

// NewSessionFactory returns a function that creates sessions (you can inject core backend)
func NewSessionFactory() func() *Session {
	return func() *Session {
		return &Session{
			buffers: make(map[string]string),
			backend: NewCoreBackend(), // implement this to wrap your core
		}
	}
}

func (s *Session) Open(uri, text string) {
	s.mu.Lock()
	s.buffers[uri] = text
	s.mu.Unlock()
	s.backend.Open(uri, text)
}

func (s *Session) Update(uri, text string) {
	s.mu.Lock()
	s.buffers[uri] = text
	s.mu.Unlock()
	s.backend.Update(uri, text)
}

func (s *Session) GetText(uri string) (string, bool) {
	s.mu.Lock()
	t, ok := s.buffers[uri]
	s.mu.Unlock()
	return t, ok
}
