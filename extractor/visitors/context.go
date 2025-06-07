package visitors

import (
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor/visitors/providers"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
)

type VisitContext struct {
	// Provides abstractions over AST and packages
	ArbitrationProvider *providers.ArbitrationProvider

	// Provides thread-safe answers for question like "What's the next unique import ID?"
	// It can be provided and chained from visitor to visitor to ensure no import naming collisions or omitted which
	// may be useful for certain use-cases
	SyncedProvider *providers.SyncedProvider

	// An interface used to build a symbol graph for the processed code
	GraphBuilder symboldg.SymbolGraphBuilder

	// The project's configuration, as specified in the user's gleece.config.json
	GleeceConfig *definitions.GleeceConfig
}
