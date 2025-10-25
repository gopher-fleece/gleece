package visitors

import (
	"errors"
	"fmt"
	"go/ast"
)

type VisitorOrchestrator struct {
	ctx *VisitContext

	typeDeclVisitor  *TypeDeclVisitor
	typeUsageVisitor *TypeUsageVisitor
	structVisitor    *StructVisitor
	enumVisitor      *EnumVisitor

	controllerVisitor *ControllerVisitor
}

func NewVisitorOrchestrator(ctx *VisitContext) (*VisitorOrchestrator, error) {
	err := validateContext(ctx)
	if err != nil {
		return nil, err
	}

	typeDeclVisitor, err := NewTypeDeclVisitor(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to construct a TypeDeclVisitor instance - %v", err)
	}

	typeUsageVisitor, err := NewTypeUsageVisitor(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to construct a TypeUsageVisitor instance - %v", err)
	}

	structVisitor, err := NewStructVisitor(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to construct a StructVisitor instance - %v", err)
	}

	enumVisitor, err := NewEnumVisitor(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to construct an EnumVisitor instance - %v", err)
	}

	fieldVisitor, err := NewFieldVisitor(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to construct an FieldVisitor instance - %v", err)
	}

	controllerVisitor, err := NewControllerVisitor(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to construct a ControllerVisitor instance - %v", err)
	}

	typeDeclVisitor.setStructVisitor(structVisitor)
	typeDeclVisitor.setEnumVisitor(enumVisitor)

	typeUsageVisitor.setDeclVisitor(typeDeclVisitor)

	structVisitor.setTypeUsageVisitor(typeUsageVisitor)

	fieldVisitor.setTypeUsageVisitor(typeUsageVisitor)

	controllerVisitor.setFieldVisitor(fieldVisitor)

	orchestrator := VisitorOrchestrator{
		ctx:               ctx,
		controllerVisitor: controllerVisitor,
		structVisitor:     structVisitor,
		enumVisitor:       enumVisitor,
		typeUsageVisitor:  typeUsageVisitor,
		typeDeclVisitor:   typeDeclVisitor,
	}

	return &orchestrator, nil
}

// Visit implements the AST Visitor interface via the internal ControllerVisitor
func (o *VisitorOrchestrator) Visit(node ast.Node) ast.Visitor {
	return o.controllerVisitor.Visit(node)
}

// GetLastError retrieves the last visitor error via the the internal ControllerVisitor.
// Note that errors are not diagnostics - they are breakages in the processing pipeline.
func (o *VisitorOrchestrator) GetLastError() error {
	return o.controllerVisitor.GetLastError()
}

// GetFormattedDiagnosticStack retrieves the current diagnostic call stack via the the internal ControllerVisitor.
// This is used for outputting human-readable error information originating from deep within the visitor hierarchy
func (o *VisitorOrchestrator) GetFormattedDiagnosticStack() string {
	return o.controllerVisitor.GetFormattedDiagnosticStack()
}

func validateContext(ctx *VisitContext) error {
	if ctx == nil {
		return errors.New("nil context was provided to VisitorOrchestrator")
	}

	errs := []error{}
	if ctx.ArbitrationProvider == nil {
		errs = append(errs, errors.New("VisitContext does not have an arbitration provider"))
	}

	if ctx.GleeceConfig == nil {
		errs = append(errs, errors.New("VisitContext does not have a Gleece Config"))
	}

	if ctx.GraphBuilder == nil {
		errs = append(errs, errors.New("VisitContext does not have a graph builder"))
	}

	if ctx.MetadataCache == nil {
		errs = append(errs, errors.New("VisitContext does not have a metadata cache"))
	}

	if ctx.SyncedProvider == nil {
		errs = append(errs, errors.New("VisitContext does not have a synchronized provider"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
