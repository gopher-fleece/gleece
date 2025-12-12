package visitors

import (
	"fmt"

	"github.com/gopher-fleece/gleece/common"
)

type UnexpectedEntityError struct {
	error
	Expected common.SymKind
	Received common.SymKind
}

func NewUnexpectedEntityError(expected, received common.SymKind, messageFormat string, args ...any) UnexpectedEntityError {
	return UnexpectedEntityError{
		error:    fmt.Errorf(messageFormat, args...),
		Expected: expected,
		Received: received,
	}
}
