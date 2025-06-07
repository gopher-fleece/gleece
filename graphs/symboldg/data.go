package symboldg

import (
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/runtime"
)

// Note to self - We may want to shield pointers here from later modification by caller by performing some deep copy.

type ControllerSymbolicMetadata struct {
	Name         string
	Package      string
	PkgPath      string
	Tag          string
	Description  string
	RestMetadata definitions.RestMetadata
	Security     []definitions.RouteSecurity
	FVersion     *gast.FileVersion
}

type RouteSymbolicMetadata struct {
	// The handler function's and operation name in the OpenAPI schema
	OperationId string

	// The HTTP verb this method expects (i.e., GET, POST etc.)
	HttpVerb definitions.HttpVerb

	// Controls whether the method is hidden in schema and when
	Hiding definitions.MethodHideOptions

	// Defines whether the method is considered deprecated
	Deprecation definitions.DeprecationOptions

	// The operation's description
	Description string

	// Additional metadata related to the operation such as it's URL
	RestMetadata definitions.RestMetadata

	// Indicates whether the operation returns a value on success.
	//
	// Note that the framework enforces at-least an error return value from all controller methods
	HasReturnValue bool

	// A description for success responses
	ResponseDescription string

	// The HTTP code expected to be returned from a successful call
	//
	// TODO: Needs to be an array
	ResponseSuccessCode runtime.HttpStatusCode

	// Metadata on the type of errors that may be returned from the operation
	ErrorResponses []definitions.ErrorResponse

	// The expected request content type.
	//
	// Currently hard-coded to application/json.
	RequestContentType definitions.ContentType

	// The expected response content type.
	//
	// Currently hard-coded to application/json.
	ResponseContentType definitions.ContentType

	// The security schema/s used for the operation
	Security []definitions.RouteSecurity // OR between security routes

	// Custom template context for the operation, provided by the route developer, used template extension/override
	TemplateContext map[string]definitions.TemplateContext

	FVersion *gast.FileVersion
}

type FuncParamSymbolicMetadata struct {
	definitions.OrderedIdent
	Name               string
	IsContext          bool
	PassedIn           definitions.ParamPassedIn
	NameInSchema       string
	Description        string
	UniqueImportSerial uint64
	Validator          string
	Deprecation        *definitions.DeprecationOptions
}

type FuncReturnValueSymbolicMetadata struct {
	definitions.OrderedIdent
	UniqueImportSerial uint64
}
