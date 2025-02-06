package definitions

import (
	"github.com/gopher-fleece/runtime"
)

const Rfc7807ErrorName = "Rfc7807Error"
const Rfc7807ErrorFullPackage = "github.com/gopher-fleece/runtime"

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
	IsByAddress           bool
	EntityKind            AstNodeKind
}

type ErrorResponse struct {
	HttpStatusCode runtime.HttpStatusCode
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
	ResponseSuccessCode runtime.HttpStatusCode

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
	SecurityAnnotation []SecurityAnnotationComponent `json:"securityMethod" validate:"not_nil_array"` // AND between security methods
}

// SecurityAnnotationComponent is the schema-scopes parts of a security annotation;
// i.e., @Security(AND, [{name: "schema1", scopes: ["read", "write"]}, {name: "schema2", scopes: ["delete"]}])
type SecurityAnnotationComponent struct {
	SchemaName string   `json:"name" validate:"required,starts_with_letter"`
	Scopes     []string `json:"scopes" validate:"not_nil_array"`
}

type ControllerMetadata struct {
	Name                  string
	Package               string
	FullyQualifiedPackage string
	Tag                   string
	Description           string
	RestMetadata          RestMetadata
	Routes                []RouteMetadata

	// The default security schema/s used for the controller's operations.
	// May be overridden at the route level
	Security []RouteSecurity
}

type ModelMetadata struct {
	Name                  string
	FullyQualifiedPackage string
	Description           string
	Fields                []FieldMetadata
	Deprecation           DeprecationOptions
}

type FieldMetadata struct {
	Name        string
	Type        string
	Description string
	Tag         string
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

type OpenAPIContact struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Email string `json:"email" validate:"email"`
}

type OpenAPILicense struct {
	Name string `json:"name" validate:"required"`
	URL  string `json:"url"`
}

type OpenAPIInfo struct {
	Title          string          `json:"title" validate:"required"`
	Description    string          `json:"description"`
	TermsOfService string          `json:"termsOfService"`
	Contact        *OpenAPIContact `json:"contact"`
	License        *OpenAPILicense `json:"license"`
	Version        string          `json:"version" validate:"required"`
}

type OpenAPIGeneratorConfig struct {
	OpenAPI              string                       `json:"openAPI" validate:"required,oneof=3.0.0 3.1.0"` // only 3.0.0 is fully supported
	Info                 OpenAPIInfo                  `json:"info" validate:"required"`
	BaseURL              string                       `json:"baseUrl" validate:"required,url"`
	SecuritySchemes      []SecuritySchemeConfig       `json:"securitySchemes" validate:"dive"`
	DefaultRouteSecurity *SecurityAnnotationComponent `json:"defaultSecurity"`
	SpecGeneratorConfig  SpecGeneratorConfig          `json:"specGeneratorConfig" validate:"required"`
}

type RoutingEngineType string

const (
	RoutingEngineGin   RoutingEngineType = "gin"
	RoutingEngineEcho  RoutingEngineType = "echo"
	RoutingEngineMux   RoutingEngineType = "mux"
	RoutingEngineFiber RoutingEngineType = "fiber"
	RoutingEngineChi   RoutingEngineType = "chi"
)

type CustomValidators struct {
	FunctionName    string `json:"functionName" validate:"required,starts_with_letter"`
	FullPackageName string `json:"fullPackageName" validate:"required"`
	ValidateTagName string `json:"validateTagName" validate:"required,starts_with_letter"`
}

type MiddlewareExecutionType string

const (
	MiddlewareExecutionBeforeOperation       MiddlewareExecutionType = "beforeOperation"
	MiddlewareExecutionAfterOperationSuccess MiddlewareExecutionType = "afterOperationSuccess"
	MiddlewareExecutionOnError               MiddlewareExecutionType = "onError"
)

type Middleware struct {
	FullPackageName string                  `json:"fullPackageName" validate:"required"`
	Execution       MiddlewareExecutionType `json:"execution" validate:"required,oneof=beforeOperation afterOperationSuccess onError"`
	FunctionName    string                  `json:"functionName" validate:"required,starts_with_letter"`
}

type RoutesConfig struct {
	Engine              RoutingEngineType   `json:"engine" validate:"required,oneof=gin echo mux fiber chi"`
	PackageName         string              `json:"packageName"`
	OutputPath          string              `json:"outputPath" validate:"required,filepath"`
	OutputFilePerms     string              `json:"outputFilePerms" validate:"regex=^(0?[0-7]{3})?$"`
	AuthorizationConfig AuthorizationConfig `json:"authorizationConfig" validate:"required"`
	TemplateOverrides   map[string]string   `json:"templateOverrides"`
	CustomValidators    []CustomValidators  `json:"customValidators" validate:"dive"`
	Middlewares         []Middleware        `json:"middlewares" validate:"dive"`
}

type AuthorizationConfig struct {
	AuthFileFullPackageName    string `json:"authFileFullPackageName" validate:"required,filepath"`
	EnforceSecurityOnAllRoutes bool   `json:"enforceSecurityOnAllRoutes"`
}

type CommonConfig struct {
	ControllerGlobs []string `json:"controllerGlobs" validate:"omitempty,min=1"`
}

type GleeceConfig struct {
	CommonConfig           CommonConfig           `json:"commonConfig" validate:"required"`
	RoutesConfig           RoutesConfig           `json:"routesConfig" validate:"required"`
	OpenAPIGeneratorConfig OpenAPIGeneratorConfig `json:"openAPIGeneratorConfig" validate:"required"`
}

type AstNodeKind string

const (
	AstNodeKindNone        AstNodeKind = "None"
	AstNodeKindUnknown     AstNodeKind = "Unknown"
	AstNodeKindInterface   AstNodeKind = "Interface"
	AstNodeKindStruct      AstNodeKind = "Struct"
	AstNodeKindIdent       AstNodeKind = "Identifier"
	AstNodeKindSelector    AstNodeKind = "SelectorExpr"
	AstNodeKindPointer     AstNodeKind = "Pointer"
	AstNodeKindArray       AstNodeKind = "Array"
	AstNodeKindMap         AstNodeKind = "Map"
	AstNodeKindChannel     AstNodeKind = "Channel"
	AstNodeKindFunction    AstNodeKind = "Function"
	AstNodeKindVariadic    AstNodeKind = "Variadic"
	AstNodeKindParenthesis AstNodeKind = "Parenthesis"
)
