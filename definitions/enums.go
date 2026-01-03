package definitions

// Enum of HTTP parma type (header, query, path, body)
type ParamPassedIn string

const (
	PassedInHeader ParamPassedIn = "Header"
	PassedInQuery  ParamPassedIn = "Query"
	PassedInPath   ParamPassedIn = "Path"
	PassedInBody   ParamPassedIn = "Body"
	PassedInForm   ParamPassedIn = "Form"
)

// An HTTP verb such as POST or GET
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

// An HTTP Content-Type
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

// An API method's visibility in the generated OpenAPI schema
type HideMethodType string

const (
	// Never hide this endpoint
	HideMethodNever HideMethodType = "Never"
	// Always hide this endpoint
	HideMethodAlways HideMethodType = "Always"
	// ***NOT YET IMPLEMENTED***
	//
	// Hide this method if the given condition is met
	HideMethodCondition HideMethodType = "Condition"
)

// HttpAuthScheme defines valid authentication schemes for HTTP type security in OpenAPI 3.0.
// These values must be registered in the IANA Authentication Scheme registry.
type HttpAuthScheme string

const (
	// HttpAuthSchemeBasic Basic authentication
	HttpAuthSchemeBasic HttpAuthScheme = "basic"
	// HttpAuthSchemeBearer Bearer token authentication (commonly used with JWT)
	HttpAuthSchemeBearer HttpAuthScheme = "bearer"
	// HttpAuthSchemeDigest Digest authentication
	HttpAuthSchemeDigest HttpAuthScheme = "digest"
	// HttpAuthSchemeHoba HTTP Origin-Bound Authentication
	HttpAuthSchemeHoba HttpAuthScheme = "hoba"
	// HttpAuthSchemeMutual Mutual TLS authentication
	HttpAuthSchemeMutual HttpAuthScheme = "mutual"
	// HttpAuthSchemeNegotiate SPNEGO/Negotiate authentication
	HttpAuthSchemeNegotiate HttpAuthScheme = "negotiate"
	// HttpAuthSchemeOauth OAuth authentication
	HttpAuthSchemeOauth HttpAuthScheme = "oauth"
	// HttpAuthSchemeScramSha1 SCRAM-SHA-1 authentication
	HttpAuthSchemeScramSha1 HttpAuthScheme = "scram-sha-1"
	// HttpAuthSchemeScramSha256 SCRAM-SHA-256 authentication
	HttpAuthSchemeScramSha256 HttpAuthScheme = "scram-sha-256"
	// HttpAuthSchemeVapid Voluntary Application Server Identification
	HttpAuthSchemeVapid HttpAuthScheme = "vapid"
)

// The security schema type for API authorization,
// that is, the definition of authentication the API server
type SecuritySchemeType string

const (
	APIKey        SecuritySchemeType = "apiKey"
	OAuth2        SecuritySchemeType = "oauth2"
	OpenIDConnect SecuritySchemeType = "openIdConnect"
	HTTP          SecuritySchemeType = "http"
)

// The way in which consumers are to pass authorization to the API server
type SecuritySchemeIn string

const (
	Empty    SecuritySchemeIn = ""
	InQuery  SecuritySchemeIn = "query"
	InHeader SecuritySchemeIn = "header"
	InCookie SecuritySchemeIn = "cookie"
)

// The API server's underlying routing engine
type RoutingEngineType string

const (
	RoutingEngineGin   RoutingEngineType = "gin"
	RoutingEngineEcho  RoutingEngineType = "echo"
	RoutingEngineMux   RoutingEngineType = "mux"
	RoutingEngineFiber RoutingEngineType = "fiber"
	RoutingEngineChi   RoutingEngineType = "chi"
)
