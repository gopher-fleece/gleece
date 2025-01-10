package definitions

// Enum of HTTP parma type (header, query, path, body)
type ParamType string

const (
	Header ParamType = "Header"
	Query  ParamType = "Query"
	Path   ParamType = "Path"
	Body   ParamType = "Body"
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

type RestMetadata struct {
	Path string
}

type FuncParam struct {
	Name           string
	ParamType      ParamType
	ParamInterface string
	Description    string
}

type RouteMetadata struct {
	OperationId         string
	HttpVerb            HttpVerb
	Description         string
	RestMetadata        RestMetadata
	FuncParams          []FuncParam
	ResponseInterface   string
	ResponseSuccessCode string
}

type ControllerMetadata struct {
	Name         string
	Package      string
	Tag          string
	Description  string
	RestMetadata RestMetadata
	Routes       []RouteMetadata
}

type ApiMetadata struct {
	Controllers [] ControllerMetadata
}