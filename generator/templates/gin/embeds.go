package gin

// This file simply embeds the default template into the compiled EmbedBin binary to allow easy runtime access

import (
	_ "embed"
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

//go:embed partials/json.body.validation.error.response.hbs
var JsonBodyValidationErrorResponse string

//go:embed partials/params.validation.error.response.hbs
var ParamsValidationErrorResponse string

//go:embed partials/response.headers.hbs
var ResponseHeaders string

//go:embed partials/json.error.response.hbs
var JsonErrorResponse string

//go:embed partials/run.validator.hbs
var RunValidator string

//go:embed partials/middleware.hbs
var Middleware string

//go:embed partials/register.middleware.hbs
var RegisterMiddleware string

//go:embed partials/method.parameter.list.hbs
var MethodParameterList string

// Those are the extension that made to allow *extend* Gleece's routes logic. as default they are all empty.
var TemplateExtensions = map[string]string{
	// On routes.hbs
	"RegisterRoutesExtension":        "// register routes extension placeholder \n",
	"RouteStartRoutesExtension":      "// route start routes extension placeholder \n",
	"BeforeOperationRoutesExtension": "// before operation routes extension placeholder \n",
	"AfterOperationRoutesExtension":  "// after operation routes extension placeholder \n",
	"RouteEndRoutesExtension":        "// route end routes extension placeholder \n",

	// On partials
	"ImportsExtension":                         "// import extension placeholder \n",
	"TypeDeclarationsExtension":                "// type declarations extension placeholder \n",
	"FunctionDeclarationsExtension":            "// function declarations extension placeholder \n",
	"JsonResponseExtension":                    "// json response extension placeholder \n",
	"RunValidatorExtension":                    "// validation error response extension placeholder \n",
	"JsonBodyValidationErrorResponseExtension": "// json body validation error response extension placeholder \n",
	"ParamsValidationErrorResponseExtension":   "// params validation error response extension placeholder \n",
	"JsonErrorResponseExtension":               "// json error response extension placeholder \n",
	"ResponseHeadersExtension":                 "// response headers extension placeholder \n",
}

var Partials = map[string]string{
	"Imports":                         Imports,
	"TypeDeclarations":                TypeDeclarations,
	"FunctionDeclarations":            FunctionDeclarations,
	"AuthorizationCall":               AuthorizationCall,
	"RequestArgsParsing":              RequestArgsParsing,
	"JsonResponse":                    JsonResponse,
	"JsonBodyValidationErrorResponse": JsonBodyValidationErrorResponse,
	"ParamsValidationErrorResponse":   ParamsValidationErrorResponse,
	"JsonErrorResponse":               JsonErrorResponse,
	"ReplyResponse":                   ReplyResponse,
	"ResponseHeaders":                 ResponseHeaders,
	"RequestSwitchParamType":          RequestSwitchParamType,
	"RunValidator":                    RunValidator,
	"Middleware":                      Middleware,
	"RegisterMiddleware":              RegisterMiddleware,
	"MethodParameterList":             MethodParameterList,
}
