package handlers

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func setTrace(context *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}
