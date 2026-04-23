package server

import (
	"fmt"

	"github.com/gopher-fleece/gleece/v2/internal/lsp/common"
	"github.com/gopher-fleece/gleece/v2/internal/lsp/handlers"
	"github.com/gopher-fleece/gleece/v2/internal/lsp/server/ipc"
	"github.com/gopher-fleece/gleece/v2/internal/lsp/state"
	"github.com/tliron/glsp/server"
)

type LangServerOptions interface {
	IpcType() ipc.LangServerIpcType
	Validate() error
}

type LanguageServer struct {
	options LangServerOptions
}

func NewLanguageServer(opts LangServerOptions) (*LanguageServer, error) {
	if opts == nil {
		return nil, fmt.Errorf("options are required")
	}

	if err := opts.Validate(); err != nil {
		return nil, err
	}

	return &LanguageServer{options: opts}, nil
}

func (s *LanguageServer) Run() error {
	state := state.WorkspaceState{}
	handler := handlers.GetProtocolHandler(&state)
	srv := server.NewServer(&handler, common.LangSrvName, false)

	switch opts := s.options.(type) {
	case ipc.StdIoOptions:
		return srv.RunStdio()
	case ipc.TcpOptions:
		return srv.RunTCP(opts.Address)
	default:
		return fmt.Errorf("unknown or unsupported ipc type '%s'", opts.IpcType())
	}
}
