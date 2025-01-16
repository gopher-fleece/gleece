package gin

// This file simply embeds the default template into the compiled EmbedBin binary to allow easy runtime access

import (
	_ "embed"

	"github.com/aymerick/raymond"
)

//go:embed routes.hbs
var RoutesTemplate string

//////////////////
//   Partials   //
//////////////////

//go:embed partials/imports.hbs
var Imports string

//go:embed partials/request.args.parsing.hbs
var RequestArgsParsing string

//go:embed partials/request.switch.param.type..hbs
var RequestSwitchParamType string

//go:embed partials/reply.response.hbs
var ReplyResponse string

//go:embed partials/json.response.hbs
var JsonResponse string

var Partials = map[string]string{
	"Imports":                Imports,
	"RequestArgsParsing":     RequestArgsParsing,
	"JsonResponse":           JsonResponse,
	"ReplyResponse":          ReplyResponse,
	"RequestSwitchParamType": RequestSwitchParamType,
}

func RegisterPartials() {
	raymond.RegisterPartials(Partials)
}
