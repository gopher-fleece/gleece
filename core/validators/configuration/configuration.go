package configuration

import "github.com/gopher-fleece/gleece/core/annotations"

var ValidatorConfigMap = map[string]AnnotationConfigDefinition{
	// Controller (Class-Level) Annotations
	annotations.GleeceAnnotationTag: {
		Contexts:            []annotations.CommentSource{"controller"},
		RequiresValue:       true,
		AllowedProperties:   map[string]PropertyDefinition{},
		AllowsMultiple:      false,
		RequiresUniqueValue: false,
	},
	annotations.GleeceAnnotationRoute: {
		Contexts:            []annotations.CommentSource{"controller", "route"},
		RequiresValue:       true,
		AllowedProperties:   map[string]PropertyDefinition{},
		AllowsMultiple:      false,
		RequiresUniqueValue: false,
	},
	annotations.GleeceAnnotationSecurity: {
		Contexts:      []annotations.CommentSource{"controller", "route"},
		RequiresValue: true,
		AllowedProperties: map[string]PropertyDefinition{
			"scopes": {
				Required:     false,
				Type:         "array",
				DefaultValue: []any{},
			},
		},
		AllowsMultiple:      true,
		RequiresUniqueValue: false,
	},
	annotations.GleeceAnnotationDescription: {
		Contexts:            []annotations.CommentSource{"controller", "route", "schema", "property"},
		RequiresValue:       false,
		AllowedProperties:   map[string]PropertyDefinition{},
		AllowsMultiple:      false,
		RequiresUniqueValue: false,
	},
	annotations.GleeceAnnotationDeprecated: {
		Contexts:            []annotations.CommentSource{"controller", "route", "schema", "property"},
		RequiresValue:       false,
		AllowedProperties:   map[string]PropertyDefinition{},
		AllowsMultiple:      false,
		RequiresUniqueValue: false,
	},

	// Route (Function-Level) Annotations
	annotations.GleeceAnnotationMethod: {
		Contexts:            []annotations.CommentSource{"route"},
		RequiresValue:       true,
		AllowedProperties:   map[string]PropertyDefinition{},
		AllowsMultiple:      false,
		RequiresUniqueValue: false,
	},
	annotations.GleeceAnnotationQuery: {
		Contexts:      []annotations.CommentSource{"route"},
		RequiresValue: true,
		AllowedProperties: map[string]PropertyDefinition{
			"name": {
				Required:     false,
				Type:         "string",
				DefaultValue: "",
			},
			"validate": {
				Required:     false,
				Type:         "string",
				DefaultValue: "",
			},
		},
		AllowsMultiple:      true,
		RequiresUniqueValue: true, // Values must be unique across all HTTP params annotations
	},
	annotations.GleeceAnnotationHeader: {
		Contexts:      []annotations.CommentSource{"route"},
		RequiresValue: true,
		AllowedProperties: map[string]PropertyDefinition{
			"name": {
				Required:     false,
				Type:         "string",
				DefaultValue: "",
			},
			"validate": {
				Required:     false,
				Type:         "string",
				DefaultValue: "",
			},
		},
		AllowsMultiple:      true,
		RequiresUniqueValue: true, // Values must be unique across all HTTP params annotations
	},
	annotations.GleeceAnnotationPath: {
		Contexts:      []annotations.CommentSource{"route"},
		RequiresValue: true,
		AllowedProperties: map[string]PropertyDefinition{
			"name": {
				Required:     false,
				Type:         "string",
				DefaultValue: "",
			},
			"validate": {
				Required:     false,
				Type:         "string",
				DefaultValue: "",
			},
		},
		AllowsMultiple:      true,
		RequiresUniqueValue: true, // Values must be unique across all HTTP params annotations
	},
	annotations.GleeceAnnotationBody: {
		Contexts:      []annotations.CommentSource{"route"},
		RequiresValue: true,
		AllowedProperties: map[string]PropertyDefinition{
			"validate": {
				Required:     false,
				Type:         "string",
				DefaultValue: "",
			},
		},
		AllowsMultiple:      false,
		MutuallyExclusive:   []string{annotations.GleeceAnnotationFormField},
		RequiresUniqueValue: true, // Values must be unique across all HTTP params annotations
	},
	annotations.GleeceAnnotationFormField: {
		Contexts:      []annotations.CommentSource{"route"},
		RequiresValue: true,
		AllowedProperties: map[string]PropertyDefinition{
			"name": {
				Required:     false,
				Type:         "string",
				DefaultValue: "",
			},
			"validate": {
				Required:     false,
				Type:         "string",
				DefaultValue: "",
			},
		},
		AllowsMultiple:      true,
		MutuallyExclusive:   []string{annotations.GleeceAnnotationBody},
		RequiresUniqueValue: true, // Values must be unique across all HTTP params annotations
	},
	annotations.GleeceAnnotationResponse: {
		Contexts:            []annotations.CommentSource{"route"},
		RequiresValue:       true,
		AllowedProperties:   map[string]PropertyDefinition{},
		AllowsMultiple:      true,
		RequiresUniqueValue: false,
	},
	annotations.GleeceAnnotationErrorResponse: {
		Contexts:            []annotations.CommentSource{"route"},
		RequiresValue:       true,
		AllowedProperties:   map[string]PropertyDefinition{},
		AllowsMultiple:      true,
		RequiresUniqueValue: false,
	},
	annotations.GleeceAnnotationHidden: {
		Contexts:            []annotations.CommentSource{"route"},
		RequiresValue:       false,
		AllowedProperties:   map[string]PropertyDefinition{},
		AllowsMultiple:      false,
		RequiresUniqueValue: false,
	},
	annotations.GleeceAnnotationTemplateContext: {
		Contexts:            []annotations.CommentSource{"route"},
		RequiresValue:       true,
		AllowedProperties:   nil, // Any properties are allowed
		AllowsMultiple:      true,
		RequiresUniqueValue: false,
	},
}
