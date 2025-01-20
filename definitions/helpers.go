package definitions

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gopher-fleece/gleece/external"
)

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
	uint(external.StatusContinue):                      {},
	uint(external.StatusSwitchingProtocols):            {},
	uint(external.StatusProcessing):                    {},
	uint(external.StatusEarlyHints):                    {},
	uint(external.StatusOK):                            {},
	uint(external.StatusCreated):                       {},
	uint(external.StatusAccepted):                      {},
	uint(external.StatusNonAuthoritativeInfo):          {},
	uint(external.StatusNoContent):                     {},
	uint(external.StatusResetContent):                  {},
	uint(external.StatusPartialContent):                {},
	uint(external.StatusMultiStatus):                   {},
	uint(external.StatusAlreadyReported):               {},
	uint(external.StatusIMUsed):                        {},
	uint(external.StatusMultipleChoices):               {},
	uint(external.StatusMovedPermanently):              {},
	uint(external.StatusFound):                         {},
	uint(external.StatusSeeOther):                      {},
	uint(external.StatusNotModified):                   {},
	uint(external.StatusUseProxy):                      {},
	uint(external.StatusTemporaryRedirect):             {},
	uint(external.StatusPermanentRedirect):             {},
	uint(external.StatusBadRequest):                    {},
	uint(external.StatusUnauthorized):                  {},
	uint(external.StatusPaymentRequired):               {},
	uint(external.StatusForbidden):                     {},
	uint(external.StatusNotFound):                      {},
	uint(external.StatusMethodNotAllowed):              {},
	uint(external.StatusNotAcceptable):                 {},
	uint(external.StatusProxyAuthRequired):             {},
	uint(external.StatusRequestTimeout):                {},
	uint(external.StatusConflict):                      {},
	uint(external.StatusGone):                          {},
	uint(external.StatusLengthRequired):                {},
	uint(external.StatusPreconditionFailed):            {},
	uint(external.StatusRequestEntityTooLarge):         {},
	uint(external.StatusRequestURITooLong):             {},
	uint(external.StatusUnsupportedMediaType):          {},
	uint(external.StatusRequestedRangeNotSatisfiable):  {},
	uint(external.StatusExpectationFailed):             {},
	uint(external.StatusTeapot):                        {},
	uint(external.StatusMisdirectedRequest):            {},
	uint(external.StatusUnprocessableEntity):           {},
	uint(external.StatusLocked):                        {},
	uint(external.StatusFailedDependency):              {},
	uint(external.StatusTooEarly):                      {},
	uint(external.StatusUpgradeRequired):               {},
	uint(external.StatusPreconditionRequired):          {},
	uint(external.StatusTooManyRequests):               {},
	uint(external.StatusRequestHeaderFieldsTooLarge):   {},
	uint(external.StatusUnavailableForLegalReasons):    {},
	uint(external.StatusInternalServerError):           {},
	uint(external.StatusNotImplemented):                {},
	uint(external.StatusBadGateway):                    {},
	uint(external.StatusServiceUnavailable):            {},
	uint(external.StatusGatewayTimeout):                {},
	uint(external.StatusHTTPVersionNotSupported):       {},
	uint(external.StatusVariantAlsoNegotiates):         {},
	uint(external.StatusInsufficientStorage):           {},
	uint(external.StatusLoopDetected):                  {},
	uint(external.StatusNotExtended):                   {},
	uint(external.StatusNetworkAuthenticationRequired): {},
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

func EnsureHttpStatusCode(code uint) external.HttpStatusCode {
	if IsValidHttpStatusCode(code) {
		return external.HttpStatusCode(code)
	}
	panic(fmt.Sprintf("'%d' is not a valid HTTP status code", code))
}

func ConvertToHttpStatus(code string) (external.HttpStatusCode, error) {
	parsed, err := strconv.ParseUint(code, 10, 32)
	if err != nil {
		return 0, err
	}

	parsedCode := uint(parsed)
	if !IsValidHttpStatusCode(parsedCode) {
		return 0, fmt.Errorf("'%s' is not a valid HTTP status code", code)
	}
	return external.HttpStatusCode(parsedCode), nil
}

func PermissionStringToFileMod(permissionString string) (os.FileMode, error) {
	permission, err := strconv.ParseUint(permissionString, 8, 32)
	if err != nil || permission&^uint64(os.ModePerm) != 0 {
		return 0, fmt.Errorf("must be a valid UNIX FileMode value")
	}
	return os.FileMode(permission), err
}
