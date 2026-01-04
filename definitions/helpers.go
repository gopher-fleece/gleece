package definitions

import (
	"fmt"
	"os"
	"slices"
	"strconv"

	"github.com/gopher-fleece/gleece/v2/common"
	"github.com/gopher-fleece/runtime"
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

// A map of HTTP verbs supported by Gleece routes.
// This is expected to eventually become deprecated as support for additional verbs is added
var routeSupportedHttpVerbs = map[string]struct{}{
	string(HttpGet):    {},
	string(HttpPost):   {},
	string(HttpPut):    {},
	string(HttpDelete): {},
	string(HttpPatch):  {},
}

var validHttpStatusCode = map[uint]struct{}{
	uint(runtime.StatusContinue):                      {},
	uint(runtime.StatusSwitchingProtocols):            {},
	uint(runtime.StatusProcessing):                    {},
	uint(runtime.StatusEarlyHints):                    {},
	uint(runtime.StatusOK):                            {},
	uint(runtime.StatusCreated):                       {},
	uint(runtime.StatusAccepted):                      {},
	uint(runtime.StatusNonAuthoritativeInfo):          {},
	uint(runtime.StatusNoContent):                     {},
	uint(runtime.StatusResetContent):                  {},
	uint(runtime.StatusPartialContent):                {},
	uint(runtime.StatusMultiStatus):                   {},
	uint(runtime.StatusAlreadyReported):               {},
	uint(runtime.StatusIMUsed):                        {},
	uint(runtime.StatusMultipleChoices):               {},
	uint(runtime.StatusMovedPermanently):              {},
	uint(runtime.StatusFound):                         {},
	uint(runtime.StatusSeeOther):                      {},
	uint(runtime.StatusNotModified):                   {},
	uint(runtime.StatusUseProxy):                      {},
	uint(runtime.StatusTemporaryRedirect):             {},
	uint(runtime.StatusPermanentRedirect):             {},
	uint(runtime.StatusBadRequest):                    {},
	uint(runtime.StatusUnauthorized):                  {},
	uint(runtime.StatusPaymentRequired):               {},
	uint(runtime.StatusForbidden):                     {},
	uint(runtime.StatusNotFound):                      {},
	uint(runtime.StatusMethodNotAllowed):              {},
	uint(runtime.StatusNotAcceptable):                 {},
	uint(runtime.StatusProxyAuthRequired):             {},
	uint(runtime.StatusRequestTimeout):                {},
	uint(runtime.StatusConflict):                      {},
	uint(runtime.StatusGone):                          {},
	uint(runtime.StatusLengthRequired):                {},
	uint(runtime.StatusPreconditionFailed):            {},
	uint(runtime.StatusRequestEntityTooLarge):         {},
	uint(runtime.StatusRequestURITooLong):             {},
	uint(runtime.StatusUnsupportedMediaType):          {},
	uint(runtime.StatusRequestedRangeNotSatisfiable):  {},
	uint(runtime.StatusExpectationFailed):             {},
	uint(runtime.StatusTeapot):                        {},
	uint(runtime.StatusMisdirectedRequest):            {},
	uint(runtime.StatusUnprocessableEntity):           {},
	uint(runtime.StatusLocked):                        {},
	uint(runtime.StatusFailedDependency):              {},
	uint(runtime.StatusTooEarly):                      {},
	uint(runtime.StatusUpgradeRequired):               {},
	uint(runtime.StatusPreconditionRequired):          {},
	uint(runtime.StatusTooManyRequests):               {},
	uint(runtime.StatusRequestHeaderFieldsTooLarge):   {},
	uint(runtime.StatusUnavailableForLegalReasons):    {},
	uint(runtime.StatusInternalServerError):           {},
	uint(runtime.StatusNotImplemented):                {},
	uint(runtime.StatusBadGateway):                    {},
	uint(runtime.StatusServiceUnavailable):            {},
	uint(runtime.StatusGatewayTimeout):                {},
	uint(runtime.StatusHTTPVersionNotSupported):       {},
	uint(runtime.StatusVariantAlsoNegotiates):         {},
	uint(runtime.StatusInsufficientStorage):           {},
	uint(runtime.StatusLoopDetected):                  {},
	uint(runtime.StatusNotExtended):                   {},
	uint(runtime.StatusNetworkAuthenticationRequired): {},
}

// GetValidHttpVerbs returns a list of all valid HTTP verbs
func GetValidHttpVerbs() []string {
	verbs := make([]string, 0, len(validHttpVerbs))
	for verb := range validHttpVerbs {
		verbs = append(verbs, verb)
	}
	return verbs
}

// GetRouteSupportedHttpVerbs returns a list of all HTTP verbs that are valid in the context of an API Endpoint
func GetRouteSupportedHttpVerbs() []string {
	verbs := common.MapKeys(routeSupportedHttpVerbs)
	slices.Sort(verbs)
	return verbs
}

// GetValidHttpStatusCodes returns a list of all known/valid HTTP Status Codes
func GetValidHttpStatusCodes() []uint {
	codes := make([]uint, 0, len(validHttpStatusCode))
	for code := range validHttpStatusCode {
		codes = append(codes, code)
	}
	return codes
}

// IsValidHttpVerb determines whether a given string is a valid HTTP verb
func IsValidHttpVerb(verb string) bool {
	_, exists := validHttpVerbs[verb]
	return exists
}

// IsValidRouteHttpVerb determines whether a given string is an HTTP verb
// that is valid in the context of an API endpoint
func IsValidRouteHttpVerb(verb string) bool {
	_, exists := routeSupportedHttpVerbs[verb]
	return exists
}

// IsValidHttpStatusCode determines whether the given code is a known, valid HTTP Status Code
func IsValidHttpStatusCode(code uint) bool {
	_, exists := validHttpStatusCode[code]
	return exists
}

// ConvertToHttpStatus attempts to convert the given code into an HTTP Status Code.
// Returns an error if the given code is invalid.
func ConvertToHttpStatus(code string) (runtime.HttpStatusCode, error) {
	parsed, err := strconv.ParseUint(code, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to convert HTTP status code string '%s' to an integer - %v", code, err)
	}

	parsedCode := uint(parsed)
	if !IsValidHttpStatusCode(parsedCode) {
		return 0, fmt.Errorf("'%s' is not a valid HTTP status code", code)
	}
	return runtime.HttpStatusCode(parsedCode), nil
}

// PermissionStringToFileMod converts the given permission string into a FileMode.
//
// Input is expected to be a valid octal permission value like '0777'
func PermissionStringToFileMod(permissionString string) (os.FileMode, error) {
	permission, err := strconv.ParseUint(permissionString, 8, 32)
	// A proper mask needs to account for sticky/setuid/setgid bitflags
	if err != nil || permission&^uint64(0o7777) != 0 {
		return 0, fmt.Errorf("must be a valid UNIX FileMode value")
	}
	return os.FileMode(permission), err
}
