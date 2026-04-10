package handlers

import (
	"github.com/gopher-fleece/gleece/v2/internal/lsp/state"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type ProtocolHandler struct {
	state *state.WorkspaceState
}

func NewProtocolHandler(state *state.WorkspaceState) *ProtocolHandler {
	return &ProtocolHandler{state: state}
}

func GetProtocolHandler(state *state.WorkspaceState) protocol.Handler {
	h := NewProtocolHandler(state)

	return protocol.Handler{
		Initialize:            h.initialize,
		Initialized:           h.initialized,
		Shutdown:              h.shutdown,
		SetTrace:              h.setTrace,
		TextDocumentDidOpen:   h.didOpen,
		TextDocumentDidChange: h.didChange,
		TextDocumentDidSave:   h.didSave,
		TextDocumentDidClose:  h.didClose,
	}
}
