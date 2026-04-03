package handlers

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func shutdown(context *glsp.Context) error {
	protocol.SetTraceValue(protocol.TraceValueOff)
	return nil
}
