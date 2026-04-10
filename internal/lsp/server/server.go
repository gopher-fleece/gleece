package server

import (
	"fmt"

	"github.com/gopher-fleece/gleece/v2/internal/lsp/common"
	"github.com/gopher-fleece/gleece/v2/internal/lsp/handlers"
	"github.com/gopher-fleece/gleece/v2/internal/lsp/state"
	"github.com/tliron/glsp/server"
)

type LangServerOptions struct {
	Ipc LangServerIpcType
}

type LanguageServer struct {
	ipcType LangServerIpcType
}

func NewLanguageServer(opts LangServerOptions) (*LanguageServer, error) {
	if opts.Ipc == "" || (opts.Ipc != LangServerIpcStdIo) {
		return nil, fmt.Errorf("ipc type '%s' is not currently supported", opts.Ipc)
	}

	return &LanguageServer{
		ipcType: opts.Ipc,
	}, nil
}

func (s *LanguageServer) Run() error {
	state := state.WorkspaceState{}
	handler := handlers.GetProtocolHandler(&state)
	srv := server.NewServer(&handler, common.LangSrvName, false)

	switch s.ipcType {
	case LangServerIpcStdIo:
		return srv.RunStdio()
	default:
		return fmt.Errorf("unknown or unsupported ipc type '%s'", s.ipcType)
	}
}
