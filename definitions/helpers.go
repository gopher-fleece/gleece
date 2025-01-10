package definitions

import "fmt"

var validHttpVerbs = map[string]struct{}{
	string(HttpGet):     {},
	string(HttpPost):    {},
	string(HttpPut):     {},
	string(HttpDelete):  {},
	string(HttpPatch):   {},
	string(HttpOptions): {},
	string(HttpHead):    {},
	string(HttpTrace):   {},
	string(HttpConnect): {},
}

var validHttpStatusCode = map[uint]struct{}{
	uint(StatusContinue):                      {},
	uint(StatusSwitchingProtocols):            {},
	uint(StatusProcessing):                    {},
	uint(StatusEarlyHints):                    {},
	uint(StatusOK):                            {},
	uint(StatusCreated):                       {},
	uint(StatusAccepted):                      {},
	uint(StatusNonAuthoritativeInfo):          {},
	uint(StatusNoContent):                     {},
	uint(StatusResetContent):                  {},
	uint(StatusPartialContent):                {},
	uint(StatusMultiStatus):                   {},
	uint(StatusAlreadyReported):               {},
	uint(StatusIMUsed):                        {},
	uint(StatusMultipleChoices):               {},
	uint(StatusMovedPermanently):              {},
	uint(StatusFound):                         {},
	uint(StatusSeeOther):                      {},
	uint(StatusNotModified):                   {},
	uint(StatusUseProxy):                      {},
	uint(StatusTemporaryRedirect):             {},
	uint(StatusPermanentRedirect):             {},
	uint(StatusBadRequest):                    {},
	uint(StatusUnauthorized):                  {},
	uint(StatusPaymentRequired):               {},
	uint(StatusForbidden):                     {},
	uint(StatusNotFound):                      {},
	uint(StatusMethodNotAllowed):              {},
	uint(StatusNotAcceptable):                 {},
	uint(StatusProxyAuthRequired):             {},
	uint(StatusRequestTimeout):                {},
	uint(StatusConflict):                      {},
	uint(StatusGone):                          {},
	uint(StatusLengthRequired):                {},
	uint(StatusPreconditionFailed):            {},
	uint(StatusRequestEntityTooLarge):         {},
	uint(StatusRequestURITooLong):             {},
	uint(StatusUnsupportedMediaType):          {},
	uint(StatusRequestedRangeNotSatisfiable):  {},
	uint(StatusExpectationFailed):             {},
	uint(StatusTeapot):                        {},
	uint(StatusMisdirectedRequest):            {},
	uint(StatusUnprocessableEntity):           {},
	uint(StatusLocked):                        {},
	uint(StatusFailedDependency):              {},
	uint(StatusTooEarly):                      {},
	uint(StatusUpgradeRequired):               {},
	uint(StatusPreconditionRequired):          {},
	uint(StatusTooManyRequests):               {},
	uint(StatusRequestHeaderFieldsTooLarge):   {},
	uint(StatusUnavailableForLegalReasons):    {},
	uint(StatusInternalServerError):           {},
	uint(StatusNotImplemented):                {},
	uint(StatusBadGateway):                    {},
	uint(StatusServiceUnavailable):            {},
	uint(StatusGatewayTimeout):                {},
	uint(StatusHTTPVersionNotSupported):       {},
	uint(StatusVariantAlsoNegotiates):         {},
	uint(StatusInsufficientStorage):           {},
	uint(StatusLoopDetected):                  {},
	uint(StatusNotExtended):                   {},
	uint(StatusNetworkAuthenticationRequired): {},
}

func IsValidHttpVerb(verb string) bool {
	_, exists := validHttpVerbs[verb]
	return exists
}

func EnsureValidHttpVerb(verb string) HttpVerb {
	if IsValidHttpVerb(verb) {
		return HttpVerb(verb)
	}
	panic(fmt.Sprintf("'%s' is not a valid HTTP verb", verb))
}

func IsValidHttpStatusCode(code uint) bool {
	_, exists := validHttpStatusCode[code]
	return exists
}

func EnsureHttpStatusCode(code uint) HttpStatusCode {
	if IsValidHttpStatusCode(code) {
		return HttpStatusCode(code)
	}
	panic(fmt.Sprintf("'%d' is not a valid HTTP status code", code))
}
