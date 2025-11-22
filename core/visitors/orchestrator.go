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
	fieldVisitor     *FieldVisitor

	controllerVisitor *ControllerVisitor
}

func NewVisitorOrchestrator(ctx *VisitContext) (*VisitorOrchestrator, error) {
	err := validateContext(ctx)
	if err != nil {
		return nil, err
	}

	aliasVisitor, err := NewAliasVisitor(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to construct an AliasVisitor instance - %v", err)
	}

	typeDeclVisitor, err := NewTypeDeclVisitor(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to construct a TypeDeclVisitor instance - %v", err)
	}

	typeUsageVisitor, err := NewTypeUsageVisitor(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to construct a TypeUsageVisitor instance - %v", err)
	}

	nonMaterializingTypeUsageVisitor, err := NewTypeUsageVisitor(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to construct a non-materializing TypeUsageVisitor instance - %v", err)
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

	// Need to re-think this whole thing. Maybe use the orchestrator itself for DI?

	aliasVisitor.setTypeDeclVisitor(typeDeclVisitor)
	aliasVisitor.setTypeUsageVisitor(typeUsageVisitor)

	typeDeclVisitor.setStructVisitor(structVisitor)
	typeDeclVisitor.setEnumVisitor(enumVisitor)
	typeDeclVisitor.setAliasVisitor(aliasVisitor)

	typeUsageVisitor.setDeclVisitor(typeDeclVisitor)

	structVisitor.setTypeUsageVisitor(typeUsageVisitor)
	structVisitor.setNonMaterializingTypeUsageVisitor(nonMaterializingTypeUsageVisitor)

	fieldVisitor.setTypeUsageVisitor(typeUsageVisitor)

	controllerVisitor.setFieldVisitor(fieldVisitor)

	orchestrator := VisitorOrchestrator{
		ctx:               ctx,
		controllerVisitor: controllerVisitor,
		structVisitor:     structVisitor,
		fieldVisitor:      fieldVisitor,
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

func (o *VisitorOrchestrator) GetAllSourceFiles() []*ast.File {
	return o.ctx.ArbitrationProvider.GetAllSourceFiles()
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

func (o *VisitorOrchestrator) GetFieldVisitor() *FieldVisitor {
	return o.fieldVisitor
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

	if ctx.Graph == nil {
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
