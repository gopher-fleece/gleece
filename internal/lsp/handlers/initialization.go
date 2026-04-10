package handlers

import (
	"github.com/gopher-fleece/gleece/v2/internal/lsp/common"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func (h *ProtocolHandler) initialize(ctx *glsp.Context, params *protocol.InitializeParams) (any, error) {
	return protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: protocol.TextDocumentSyncKindFull,
		},
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name: common.LangSrvName,
		},
	}, nil
}

func (h *ProtocolHandler) initialized(ctx *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}
