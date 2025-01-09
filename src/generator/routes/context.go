package routes

import (
	"github.com/haimkastner/gleece/cmd"
	"github.com/haimkastner/gleece/src/generator/structs"
)

type BodyMetadata struct {
}

type QueryParams struct {
	Name string
}

type PathArguments struct {
}

type Route struct {
	Path                string
	Verb                structs.HttpVerb
	Body                BodyMetadata
	Query               QueryParams
	PathArgs            PathArguments
	RequestContentType  structs.ContentType
	ResponseContentType structs.ContentType
}

type ControllerMeta struct {
	Routes []Route
}

type RoutesContext struct {
	PackageName string
	Imports     []string
	Routes      []Route
}

func GetTemplateContext(cmd.RoutesConfig) (RoutesContext, error) {
	return RoutesContext{}, nil
}
