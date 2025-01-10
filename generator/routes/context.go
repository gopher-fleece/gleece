package routes

import (
	"github.com/haimkastner/gleece/cmd"
	"github.com/haimkastner/gleece/definitions"
)

type Argument struct {
	Type      definitions.ParamType
	Name      string
	ValueType any
}

type Route struct {
	Path                string
	Verb                definitions.HttpVerb
	RequestContentType  definitions.ContentType
	ResponseContentType definitions.ContentType
	Arguments           []Argument
}

type ControllerMeta struct {
	Routes []Route
}

type PackageImport struct {
	FullPath string
	Alias    string
}

type RoutesContext struct {
	PackageName string
	Imports     []PackageImport
	Routes      []Route
}

func GetTemplateContext(
	cmd.RoutesConfig,
	definitions.ApiMetadata,
) (RoutesContext, error) {
	return RoutesContext{}, nil
}
