package handlers

import protocol "github.com/tliron/glsp/protocol_3_16"

func GetProtocolHandler() protocol.Handler {
	return protocol.Handler{
		Initialize:  initialize,
		Initialized: initialized,
		Shutdown:    shutdown,
		SetTrace:    setTrace,
	}
}
