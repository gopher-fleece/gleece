package gin

// This file simply embeds the default template into the compiled EmbedBin binary to allow easy runtime access

import (
	_ "embed"

	"github.com/aymerick/raymond"
)

//go:embed routes.hbs
var RoutesTemplate string

//go:embed request.args.parsing.hbs
var RequestArgsParsingPartial string

//go:embed request.response.partial.hbs
var RequestResponsePartial string

var Partials = map[string]string{
	"RequestArgsParsingPartial": RequestArgsParsingPartial,
	"RequestResponsePartial":    RequestResponsePartial,
}

func RegisterPartials() {
	raymond.RegisterPartials(Partials)
}
