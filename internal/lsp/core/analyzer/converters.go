package analyzer

import (
	gleeceCommon "github.com/gopher-fleece/gleece/v2/common"
	gleeceDiag "github.com/gopher-fleece/gleece/v2/core/validators/diagnostics"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func gleeceDiagRangeToLsp(rangeValue gleeceCommon.ResolvedRange) protocol.Range {
	return protocol.Range{
		Start: protocol.Position{
			Line:      protocol.UInteger(rangeValue.StartLine),
			Character: protocol.UInteger(rangeValue.StartCol),
		},
		End: protocol.Position{
			Line:      protocol.UInteger(rangeValue.EndLine),
			Character: protocol.UInteger(rangeValue.EndCol),
		},
	}
}

func resolvedGleeceDiagToLsp(diag gleeceDiag.ResolvedDiagnostic) protocol.Diagnostic {
	return protocol.Diagnostic{
		Message:  diag.Message,
		Severity: gleeceCommon.Ptr(protocol.DiagnosticSeverity(diag.Severity)),
		Range:    gleeceDiagRangeToLsp(diag.Range),
		Code:     gleeceCommon.Ptr(protocol.IntegerOrString{Value: diag.Code}),
		Source:   gleeceCommon.Ptr(diag.Source),
	}
}

func entityDiagToLsp(entityDiag gleeceDiag.EntityDiagnostic) []protocol.Diagnostic {
	var lspDiags []protocol.Diagnostic

	for _, selfDiag := range entityDiag.Diagnostics {
		lspDiags = append(lspDiags, resolvedGleeceDiagToLsp(selfDiag))
	}

	for _, childDiag := range entityDiag.Children {
		if childDiag == nil {
			continue
		}
		lspDiags = append(lspDiags, entityDiagToLsp(*childDiag)...)
	}

	return lspDiags
}

func entityDiagsToLsp(diags []gleeceDiag.EntityDiagnostic) []protocol.Diagnostic {
	var lspDiags []protocol.Diagnostic
	for _, diag := range diags {
		lspDiags = append(lspDiags, entityDiagToLsp(diag)...)
	}
	return lspDiags
}
