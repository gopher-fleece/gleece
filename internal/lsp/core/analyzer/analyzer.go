package analyzer

import (
	"fmt"

	"github.com/gopher-fleece/gleece/v2/core/pipeline"
	"github.com/gopher-fleece/gleece/v2/core/validators/diagnostics"
	"github.com/gopher-fleece/gleece/v2/definitions"
	"github.com/gopher-fleece/gleece/v2/gast"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type GleeceAnalyzer struct {
	Pipeline pipeline.GleecePipeline
}

func NewGleeceAnalyzer(config *definitions.GleeceConfig) (*GleeceAnalyzer, error) {
	pipe, err := pipeline.NewGleecePipeline(config)
	if err != nil {
		return nil, fmt.Errorf("the analyzer could not construct a Gleece pipeline - %v", err)
	}

	return &GleeceAnalyzer{Pipeline: pipe}, nil
}

func (analyzer *GleeceAnalyzer) OnFileDidChange(
	ctx *glsp.Context,
	document *protocol.DidChangeTextDocumentParams,
) error {
	return analyzer.invalidateFile(ctx, document.TextDocument.URI, true)
}

func (analyzer *GleeceAnalyzer) OnFileDidOpen(
	ctx *glsp.Context,
	document *protocol.DidOpenTextDocumentParams,
) error {
	return nil
}

func (analyzer *GleeceAnalyzer) invalidateFile(ctx *glsp.Context, filePath string, reAnalyze bool) error {
	fVer, fvErr := gast.NewFileVersion(filePath)
	if fvErr != nil {
		return fvErr
	}

	analyzer.Pipeline.Graph().InvalidateFileVersion(fVer)

	if reAnalyze {
		diags, pipeErr := analyzer.Pipeline.Validate()
		if pipeErr != nil {
			// Logger required
			return pipeErr
		}

		analyzer.pushDiagnosticsAsync(ctx, filePath, diags)
	}

	return nil
}

func (analyzer *GleeceAnalyzer) pushDiagnosticsAsync(ctx *glsp.Context, fileUri string, diagnostics []diagnostics.EntityDiagnostic) {
	go ctx.Notify(
		protocol.ServerTextDocumentPublishDiagnostics,
		protocol.PublishDiagnosticsParams{
			URI:         fileUri,
			Diagnostics: entityDiagsToLsp(diagnostics),
		},
	)
}
