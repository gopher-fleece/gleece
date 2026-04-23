package state

import "github.com/gopher-fleece/gleece/v2/internal/lsp/core/analyzer"

type WorkspaceState struct {
	Analyzer  *analyzer.GleeceAnalyzer
	Documents DocumentsState
	Config    ExtensionConfig
}
