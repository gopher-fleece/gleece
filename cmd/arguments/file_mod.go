package arguments

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type FileModeArg struct {
	RawValue   string
	OctalValue uint32
}

func NewFileModeArg(value string) FileModeArg {
	fileMode := FileModeArg{}
	fileMode.Set(value)
	return fileMode
}

func (e *FileModeArg) String() string {
	return fmt.Sprintf("%#o", e.OctalValue)
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (e *FileModeArg) Set(v string) error {

	permission, err := strconv.ParseUint(v, 8, 32)
	if err != nil || permission&^uint64(os.ModePerm) != 0 {
		return errors.New("must be a valid UNIX FileMode value")
	}

	e.RawValue = v
	e.OctalValue = uint32(permission)

	return nil
}

// Type is only used in help text
func (e *FileModeArg) Type() string {
	return "FileModeArg"
}

func (e *FileModeArg) FileMode() os.FileMode {
	return os.FileMode(e.OctalValue)
}
