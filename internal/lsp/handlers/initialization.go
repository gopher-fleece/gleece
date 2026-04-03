package handlers

import (
	"github.com/gopher-fleece/gleece/v2/internal/lsp/common"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func initialize(ctx *glsp.Context, params *protocol.InitializeParams) (any, error) {
	caps := protocol.ServerCapabilities{
		TextDocumentSync: protocol.TextDocumentSyncKindFull, // Can move to incremental sync later on
	}

	return protocol.InitializeResult{
		Capabilities: caps,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name: common.LangSrvName,
		},
	}, nil
}

func initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}
