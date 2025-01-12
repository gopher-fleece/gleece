package swagen

import (
	"strconv"

	"github.com/haimkastner/gleece/definitions"
)

func HttpStatusCodeToString(httpStatusCode definitions.HttpStatusCode) string {
	statusCode := uint64(httpStatusCode)
	return strconv.FormatUint(statusCode, 10)
}
