package handlers

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func (h *ProtocolHandler) didOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	if params == nil {
		return nil
	}

	return h.state.Analyzer.OnFileDidOpen(ctx, params)
}

func (h *ProtocolHandler) didChange(ctx *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	if params == nil {
		return nil
	}

	return h.state.Analyzer.OnFileDidChange(ctx, params)
}

func (h *ProtocolHandler) didSave(ctx *glsp.Context, params *protocol.DidSaveTextDocumentParams) error {
	return nil
}

func (h *ProtocolHandler) didClose(ctx *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	if params == nil {
		return nil
	}

	// Clear state if needed
	return nil
}
