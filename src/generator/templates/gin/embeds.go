package gin

// This file simply embeds the default template into the compiled EmbedBin binary to allow easy runtime access

import (
	_ "embed"
	"github.com/aymerick/raymond"
)

//go:routes.hbs
var RoutesTemplate string

//go:controller.response.partial.hbs
var RoutesControllerResponsePartial string

var Partials = map[string]string{
	"RoutesControllerResponsePartial": RoutesControllerResponsePartial,
}

func RegisterPartials() {
	raymond.RegisterPartials(Partials)
}
