package definitions

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/haimkastner/gleece/external"
)

// Enum of HTTP parma type (header, query, path, body)
type ParamPassedIn string

const (
	PassedInHeader ParamPassedIn = "Header"
	PassedInQuery  ParamPassedIn = "Query"
	PassedInPath   ParamPassedIn = "Path"
	PassedInBody   ParamPassedIn = "Body"
)

type HttpVerb string

const (
	HttpGet     HttpVerb = "GET"
	HttpPost    HttpVerb = "POST"
	HttpPut     HttpVerb = "PUT"
	HttpDelete  HttpVerb = "DELETE"
	HttpPatch   HttpVerb = "PATCH"
	HttpOptions HttpVerb = "OPTIONS"
	HttpHead    HttpVerb = "HEAD"
	HttpTrace   HttpVerb = "TRACE"
	HttpConnect HttpVerb = "CONNECT"
)

type ContentType string

const (
	ContentTypeJSON           ContentType = "application/json"
	ContentTypeXML            ContentType = "application/xml"
	ContentTypeHTML           ContentType = "text/html"
	ContentTypePlainText      ContentType = "text/plain"
	ContentTypeFormURLEncoded ContentType = "application/x-www-form-urlencoded"
	ContentTypeMultipartForm  ContentType = "multipart/form-data"
	ContentTypeOctetStream    ContentType = "application/octet-stream"
	ContentTypePDF            ContentType = "application/pdf"
	ContentTypePNG            ContentType = "image/png"
	ContentTypeJPEG           ContentType = "image/jpeg"
	ContentTypeGIF            ContentType = "image/gif"
	ContentTypeCSV            ContentType = "text/csv"
	ContentTypeJavaScript     ContentType = "application/javascript"
	ContentTypeCSS            ContentType = "text/css"
)

type HideMethodType string

const (
	HideMethodNever     HideMethodType = "Never"
	HideMethodAlways    HideMethodType = "Always"
	HideMethodCondition HideMethodType = "Condition"
)

type MethodHideOptions struct {
	Type      HideMethodType
	Condition string
}

type DeprecationOptions struct {
	Deprecated  bool
	Description string
}

type ImportType string

const (
	ImportTypeNone  ImportType = "None"
	ImportTypeAlias ImportType = "Alias"
	ImportTypeDot   ImportType = "Dot"
)

type RestMetadata struct {
	Path string
}

type ParamMeta struct {
	Name     string
	TypeMeta TypeMetadata
}

type FuncParam struct {
	ParamMeta
	PassedIn           ParamPassedIn
	NameInSchema       string
	Description        string
	UniqueImportSerial uint64
	Validator          string
	Deprecation        *DeprecationOptions
}

type FuncReturnValue struct {
	TypeMetadata
	UniqueImportSerial uint64
}

type TypeMetadata struct {
	Name                  string
	FullyQualifiedPackage string
	DefaultPackageAlias   string
	Description           string
	Import                ImportType
	IsUniverseType        bool
}

type ErrorResponse struct {
	HttpStatusCode external.HttpStatusCode
	Description    string
}

type RouteMetadata struct {
	// The handler function's and operation name in the OpenAPI schema
	OperationId string

	// The HTTP verb this method expects (i.e., GET, POST etc.)
	HttpVerb HttpVerb

	// Controls whether the method is hidden in schema and when
	Hiding MethodHideOptions

	// Defines whether the method is considered deprecated
	Deprecation DeprecationOptions

	// The operation's description
	Description string

	// Additional metadata related to the operation such as it's URL
	RestMetadata RestMetadata

	// Metadata on the handler function's parameters
	FuncParams []FuncParam

	// Metadata on the handler function's return values
	Responses []FuncReturnValue

	// Indicates whether the operation returns a value on success.
	//
	// Note that the framework enforces at-least an error return value from all controller methods
	HasReturnValue bool

	// A description for success responses
	ResponseDescription string

	// The HTTP code expected to be returned from a successful call
	//
	// TODO: Needs to be an array
	ResponseSuccessCode external.HttpStatusCode

	// Metadata on the type of errors that may be returned from the operation
	ErrorResponses []ErrorResponse

	// The expected request content type.
	//
	// Currently hard-coded to application/json.
	RequestContentType ContentType

	// The expected response content type.
	//
	// Currently hard-coded to application/json.
	ResponseContentType ContentType

	// The security schema/s used for the operation
	Security []RouteSecurity // OR between security routes
}

func (m RouteMetadata) GetValueReturnType() *TypeMetadata {
	if len(m.Responses) <= 1 {
		// No return value, just error
		return nil
	}

	// We're assuming the controller method's return signature always starts with the value
	return &m.Responses[0].TypeMetadata
}

func (m RouteMetadata) GetErrorReturnType() *TypeMetadata {
	if len(m.Responses) <= 1 {
		// If there is only one return value, it's the error
		return &m.Responses[0].TypeMetadata
	}

	// We're assuming the controller method's return signature always ends with the error
	return &m.Responses[1].TypeMetadata
}

type RouteSecurity struct {
	SecurityMethod []SecurityMethod `json:"securityMethod" validate:"not_nil_array"` // AND between security methods
}

type SecurityMethod struct {
	Name        string   `json:"name" validate:"required,starts_with_letter"`
	Permissions []string `json:"permissions" validate:"not_nil_array"`
}

type ControllerMetadata struct {
	Name                  string
	Package               string
	FullyQualifiedPackage string
	Tag                   string
	Description           string
	RestMetadata          RestMetadata
	Routes                []RouteMetadata
}

type ModelMetadata struct {
	Name                  string
	Package               string
	FullyQualifiedPackage string
	Description           string
	Fields                []FieldMetadata
	Deprecation           DeprecationOptions
}

type FieldMetadata struct {
	Name        string
	Type        string
	Description string
	Validator   string
	Deprecation *DeprecationOptions
}

type SecuritySchemeType string

const (
	APIKey        SecuritySchemeType = "apiKey"
	OAuth2        SecuritySchemeType = "oauth2"
	OpenIDConnect SecuritySchemeType = "openIdConnect"
	HTTP          SecuritySchemeType = "http"
)

type SecuritySchemeIn string

const (
	InQuery  SecuritySchemeIn = "query"
	InHeader SecuritySchemeIn = "header"
	InCookie SecuritySchemeIn = "cookie"
)

type SpecGeneratorConfig struct {
	OutputPath string `json:"outputPath" validate:"required"`
}

type SecuritySchemeConfig struct {
	Description  string             `json:"description" validate:"required"`
	SecurityName string             `json:"name" validate:"required,starts_with_letter"`
	FieldName    string             `json:"fieldName" validate:"required,starts_with_letter"`
	Type         SecuritySchemeType `json:"type" validate:"required,security_schema_type"` // see SecuritySchemeType
	In           SecuritySchemeIn   `json:"in" validate:"required,security_schema_in"`     // see SecuritySchemeIn
}

type OpenAPIGeneratorConfig struct {
	Info                 openapi3.Info          `json:"info" validate:"required"`
	BaseURL              string                 `json:"base_url" validate:"required,url"`
	SecuritySchemes      []SecuritySchemeConfig `json:"securitySchemes" validate:"not_nil_array"`
	DefaultRouteSecurity []RouteSecurity        `json:"defaultSecurity" validate:"not_nil_array"`
	SpecGeneratorConfig  SpecGeneratorConfig    `json:"specGeneratorConfig" validate:"required"`
}
type GleeceConfig struct {
	OpenAPIGeneratorConfig OpenAPIGeneratorConfig `json:"openAPIGeneratorConfig" validate:"required"`
}
