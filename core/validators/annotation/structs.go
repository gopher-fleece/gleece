package annotation

import "github.com/gopher-fleece/gleece/core/annotations"

// PropertyDefinition defines the allowed properties for an annotation
type PropertyDefinition struct {
	Required      bool   // Whether the property is required
	DefaultValue  any    // Default value if not provided
	AllowedValues []any  // List of allowed values, if empty any value is allowed
	Type          string // Expected type (string, number, boolean, array, object)
}

type AnnotationConfigDefinition struct {
	Contexts            []annotations.CommentSource   // Where this annotation can be used (controller, route, schema, property)
	RequiresValue       bool                          // Whether the annotation requires a basic parameter
	AllowedProperties   map[string]PropertyDefinition // Allowed properties and their definitions
	AllowsMultiple      bool                          // Whether multiple instances of this annotation are allowed
	MutuallyExclusive   []string                      // Names of annotations that cannot be used together with this annotation
	RequiresUniqueValue bool                          // Whether the annotation value must be unique across all annotations
}
