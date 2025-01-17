package definitions

import "github.com/haimkastner/gleece/external"

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

type FuncParamLegacy struct {
	// The type of the parameter e.g. string, int, etc.
	ParamInterface        string
	Name                  string
	ParamType             ParamPassedIn
	ParamExpressionName   string
	Description           string
	FullyQualifiedPackage string
	Validator             string
}

type RouteMetadata struct {
	OperationId         string
	HttpVerb            HttpVerb
	Description         string
	RestMetadata        RestMetadata
	FuncParams          []FuncParam
	Responses           []FuncReturnValue
	HasReturnValue      bool
	ResponseDescription string
	ResponseSuccessCode external.HttpStatusCode
	ErrorResponses      []ErrorResponse
	RequestContentType  ContentType
	ResponseContentType ContentType
	Security            []RouteSecurity // OR between security routes
}

func (m RouteMetadata) GetValueReturnType() *TypeMetadata {
	if len(m.Responses) <= 1 {
		// No return value, just error
		return nil
	}

	// We're assuming the controller method's return signature always starts with the value
	return &m.Responses[0].TypeMetadata
}

type RouteSecurity struct {
	SecurityMethod []SecurityMethod // AND between security methods
}

type SecurityMethod struct {
	Name        string
	Permissions []string
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
}

type FieldMetadata struct {
	Name        string
	Type        string
	Description string
	Validator   string
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
