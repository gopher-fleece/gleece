package gin

// This file simply embeds the default template into the compiled EmbedBin binary to allow easy runtime access

import (
	_ "embed"

	"github.com/aymerick/raymond"
)

//go:embed routes.hbs
var RoutesTemplate string

//go:embed response.partial.hbs
var RoutesControllerResponsePartial string

//go:embed request.args.parsing.hbs
var RoutesRequestArgsParsing string

var Partials = map[string]string{
	"RoutesControllerResponsePartial": RoutesControllerResponsePartial,
	"RoutesRequestArgsParsing":        RoutesRequestArgsParsing,
}

func RegisterPartials() {
	raymond.RegisterPartials(Partials)
}
