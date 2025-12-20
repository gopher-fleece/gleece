package visitors

import (
	"fmt"

	"github.com/gopher-fleece/gleece/common"
)

// An error that indicates the given entity was not expected in this context.
//
// This error may be used to re-direct code flow between different heuristics, like with aliases and enums
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
