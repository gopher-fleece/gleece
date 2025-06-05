package symboldag

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor/annotations"
)

// ControllerNodeData represents a controller in the DAG.
type ControllerNodeData struct {
	TypeSpec    *ast.TypeSpec                        // The AST node for the controller struct
	Annotations *annotations.AnnotationHolder        // Comments associated with the controller
	Metadata    *definitions.ControllerLevelMetadata // Reduced metadata with just relevant graph content
}

// RouteNodeData represents an individual route handler.
type RouteNodeData struct {
	FuncDecl    *ast.FuncDecl                   // The function declaration for the route
	Annotations *annotations.AnnotationHolder   // Route-level comments
	Metadata    *definitions.RouteLevelMetadata // Slim version of route metadata (includes params/retvals by reference)
}

// ParameterNodeData represents a parameter to a route.
type ParameterNodeData struct {
	Field    *ast.Field
	Metadata *definitions.FuncParam
}

// ReturnValueNodeData represents a route's return value.
type ReturnValueNodeData struct {
	Field    *ast.Field
	Metadata *definitions.FuncReturnValue
}

// StructNodeData represents a standalone type/struct used in parameters or return values.
type StructNodeData struct {
	TypeSpec    *ast.TypeSpec
	Annotations *annotations.AnnotationHolder
	Metadata    *definitions.TypeMetadata // Simplified if needed
}

// FieldNodeData represents a field in a struct (optional granularity).
type FieldNodeData struct {
	Field    *ast.Field
	Parent   *ast.TypeSpec              // Optional link to parent struct
	Metadata *definitions.FieldMetadata // If such exists
}
