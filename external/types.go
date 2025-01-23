package external

type HttpStatusCode uint

// Taken from net/http
// Prefer to have this typed in-house as an 'enum'
const (
	StatusContinue           HttpStatusCode = 100 // RFC 9110, 15.2.1
	StatusSwitchingProtocols HttpStatusCode = 101 // RFC 9110, 15.2.2
	StatusProcessing         HttpStatusCode = 102 // RFC 2518, 10.1
	StatusEarlyHints         HttpStatusCode = 103 // RFC 8297

	StatusOK                   HttpStatusCode = 200 // RFC 9110, 15.3.1
	StatusCreated              HttpStatusCode = 201 // RFC 9110, 15.3.2
	StatusAccepted             HttpStatusCode = 202 // RFC 9110, 15.3.3
	StatusNonAuthoritativeInfo HttpStatusCode = 203 // RFC 9110, 15.3.4
	StatusNoContent            HttpStatusCode = 204 // RFC 9110, 15.3.5
	StatusResetContent         HttpStatusCode = 205 // RFC 9110, 15.3.6
	StatusPartialContent       HttpStatusCode = 206 // RFC 9110, 15.3.7
	StatusMultiStatus          HttpStatusCode = 207 // RFC 4918, 11.1
	StatusAlreadyReported      HttpStatusCode = 208 // RFC 5842, 7.1
	StatusIMUsed               HttpStatusCode = 226 // RFC 3229, 10.4.1

	StatusMultipleChoices  HttpStatusCode = 300 // RFC 9110, 15.4.1
	StatusMovedPermanently HttpStatusCode = 301 // RFC 9110, 15.4.2
	StatusFound            HttpStatusCode = 302 // RFC 9110, 15.4.3
	StatusSeeOther         HttpStatusCode = 303 // RFC 9110, 15.4.4
	StatusNotModified      HttpStatusCode = 304 // RFC 9110, 15.4.5
	StatusUseProxy         HttpStatusCode = 305 // RFC 9110, 15.4.6

	StatusTemporaryRedirect HttpStatusCode = 307 // RFC 9110, 15.4.8
	StatusPermanentRedirect HttpStatusCode = 308 // RFC 9110, 15.4.9

	StatusBadRequest                   HttpStatusCode = 400 // RFC 9110, 15.5.1
	StatusUnauthorized                 HttpStatusCode = 401 // RFC 9110, 15.5.2
	StatusPaymentRequired              HttpStatusCode = 402 // RFC 9110, 15.5.3
	StatusForbidden                    HttpStatusCode = 403 // RFC 9110, 15.5.4
	StatusNotFound                     HttpStatusCode = 404 // RFC 9110, 15.5.5
	StatusMethodNotAllowed             HttpStatusCode = 405 // RFC 9110, 15.5.6
	StatusNotAcceptable                HttpStatusCode = 406 // RFC 9110, 15.5.7
	StatusProxyAuthRequired            HttpStatusCode = 407 // RFC 9110, 15.5.8
	StatusRequestTimeout               HttpStatusCode = 408 // RFC 9110, 15.5.9
	StatusConflict                     HttpStatusCode = 409 // RFC 9110, 15.5.10
	StatusGone                         HttpStatusCode = 410 // RFC 9110, 15.5.11
	StatusLengthRequired               HttpStatusCode = 411 // RFC 9110, 15.5.12
	StatusPreconditionFailed           HttpStatusCode = 412 // RFC 9110, 15.5.13
	StatusRequestEntityTooLarge        HttpStatusCode = 413 // RFC 9110, 15.5.14
	StatusRequestURITooLong            HttpStatusCode = 414 // RFC 9110, 15.5.15
	StatusUnsupportedMediaType         HttpStatusCode = 415 // RFC 9110, 15.5.16
	StatusRequestedRangeNotSatisfiable HttpStatusCode = 416 // RFC 9110, 15.5.17
	StatusExpectationFailed            HttpStatusCode = 417 // RFC 9110, 15.5.18
	StatusTeapot                       HttpStatusCode = 418 // RFC 9110, 15.5.19 (Unused)
	StatusMisdirectedRequest           HttpStatusCode = 421 // RFC 9110, 15.5.20
	StatusUnprocessableEntity          HttpStatusCode = 422 // RFC 9110, 15.5.21
	StatusLocked                       HttpStatusCode = 423 // RFC 4918, 11.3
	StatusFailedDependency             HttpStatusCode = 424 // RFC 4918, 11.4
	StatusTooEarly                     HttpStatusCode = 425 // RFC 8470, 5.2.
	StatusUpgradeRequired              HttpStatusCode = 426 // RFC 9110, 15.5.22
	StatusPreconditionRequired         HttpStatusCode = 428 // RFC 6585, 3
	StatusTooManyRequests              HttpStatusCode = 429 // RFC 6585, 4
	StatusRequestHeaderFieldsTooLarge  HttpStatusCode = 431 // RFC 6585, 5
	StatusUnavailableForLegalReasons   HttpStatusCode = 451 // RFC 7725, 3

	StatusInternalServerError           HttpStatusCode = 500 // RFC 9110, 15.6.1
	StatusNotImplemented                HttpStatusCode = 501 // RFC 9110, 15.6.2
	StatusBadGateway                    HttpStatusCode = 502 // RFC 9110, 15.6.3
	StatusServiceUnavailable            HttpStatusCode = 503 // RFC 9110, 15.6.4
	StatusGatewayTimeout                HttpStatusCode = 504 // RFC 9110, 15.6.5
	StatusHTTPVersionNotSupported       HttpStatusCode = 505 // RFC 9110, 15.6.6
	StatusVariantAlsoNegotiates         HttpStatusCode = 506 // RFC 2295, 8.1
	StatusInsufficientStorage           HttpStatusCode = 507 // RFC 4918, 11.5
	StatusLoopDetected                  HttpStatusCode = 508 // RFC 5842, 7.2
	StatusNotExtended                   HttpStatusCode = 510 // RFC 2774, 7
	StatusNetworkAuthenticationRequired HttpStatusCode = 511 // RFC 6585, 6
)

type Rfc7807Error struct {
	Type       string            `json:"type"`
	Title      string            `json:"title"`
	Detail     string            `json:"detail"`
	Status     int               `json:"status"`
	Instance   string            `json:"instance"`
	Extensions map[string]string `json:"extensions"`
}

// GleeceController provides common functionality for controllers.
type GleeceController struct {
	statusCode *HttpStatusCode
	headers    map[string]interface{}
	// Request is the HTTP request from the underlying routing engine (gin, echo etc.)
	request any
}

func (gc *GleeceController) SetRequest(request any) {
	gc.request = request
}

// SetStatus sets the status code for the GleeceController.
func (gc *GleeceController) SetStatus(statusCode HttpStatusCode) {
	gc.statusCode = &statusCode
}

// GetStatus gets the status code for the GleeceController.
func (gc *GleeceController) GetStatus() *HttpStatusCode {
	return gc.statusCode
}

// SetHeader sets a header for the GleeceController.
func (gc *GleeceController) SetHeader(name string, value interface{}) {
	gc.headers[name] = value
}

type Controller interface {
	SetRequest(request any)

	// SetStatus sets the status code for the GleeceController.
	SetStatus(statusCode HttpStatusCode)

	// GetStatus gets the status code for the GleeceController.
	GetStatus() *HttpStatusCode

	// SetHeader sets a header for the GleeceController.
	SetHeader(name string, value interface{})
}

type SecurityCheck struct {
	SchemaName string   `json:"name" validate:"required,starts_with_letter"`
	Scopes     []string `json:"scopes" validate:"not_nil_array"`
}

type CustomError struct {
	Payload any
}

type SecurityError struct {
	Message     string
	StatusCode  HttpStatusCode
	CustomError *CustomError
}
