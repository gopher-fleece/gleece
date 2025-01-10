package generator

import (
	"github.com/aymerick/raymond"
	"github.com/iancoleman/strcase"
)

func RegisterHbsHelpers() {
	raymond.RegisterHelper("ToLowerCamel", func(arg string) string {
		return strcase.ToLowerCamel(arg)
	})
}
