package handlers

import (
	"github.com/gopher-fleece/gleece/v2/cmd"
	"github.com/gopher-fleece/gleece/v2/internal/lsp/common"
	"github.com/gopher-fleece/gleece/v2/internal/lsp/core/analyzer"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func (h *ProtocolHandler) initialize(ctx *glsp.Context, params *protocol.InitializeParams) (any, error) {
	config, err := cmd.LoadGleeceConfig("/mnt/7e91759c-6dd7-4c99-8d38-e6422452a469/git/gleece/gleece.config.json")
	if err != nil {
		return nil, err
	}

	analyzerInstance, err := analyzer.NewGleeceAnalyzer(config)
	if err != nil {
		return nil, err
	}

	h.state.Analyzer = analyzerInstance

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
	err := h.state.Analyzer.Pipeline.GenerateGraph()
	if err != nil {
		return err
	}

	return nil
}
