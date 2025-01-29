package echo

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

//go:embed partials/type.declarations.hbs
var TypeDeclarations string

//go:embed partials/function.declarations.hbs
var FunctionDeclarations string

//go:embed partials/authorization.call.hbs
var AuthorizationCall string

//go:embed partials/request.args.parsing.hbs
var RequestArgsParsing string

//go:embed partials/request.switch.param.type.hbs
var RequestSwitchParamType string

//go:embed partials/reply.response.hbs
var ReplyResponse string

//go:embed partials/json.response.hbs
var JsonResponse string

//go:embed partials/json.validation.error.response.hbs
var JsonValidationErrorResponse string

//go:embed partials/json.body.validation.error.response.hbs
var JsonBodyValidationErrorResponse string

//go:embed partials/response.headers.hbs
var ResponseHeaders string

//go:embed partials/json.error.response.hbs
var JsonErrorResponse string

//go:embed partials/run.validator.hbs
var RunValidator string

var Partials = map[string]string{
	"Imports":                         Imports,
	"TypeDeclarations":                TypeDeclarations,
	"FunctionDeclarations":            FunctionDeclarations,
	"AuthorizationCall":               AuthorizationCall,
	"RequestArgsParsing":              RequestArgsParsing,
	"JsonResponse":                    JsonResponse,
	"JsonValidationErrorResponse":     JsonValidationErrorResponse,
	"JsonBodyValidationErrorResponse": JsonBodyValidationErrorResponse,
	"JsonErrorResponse":               JsonErrorResponse,
	"ReplyResponse":                   ReplyResponse,
	"ResponseHeaders":                 ResponseHeaders,
	"RequestSwitchParamType":          RequestSwitchParamType,
	"RunValidator":                    RunValidator,
}

func RegisterPartials() {
	raymond.RegisterPartials(Partials)
}
