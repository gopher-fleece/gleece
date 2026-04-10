package state

import "github.com/gopher-fleece/gleece/v2/gast"

type DocumentState struct {
	FVersion gast.FileVersion
}

type DocumentsState struct {
	OpenDocuments map[string]DocumentState
}
