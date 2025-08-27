package annotations

const (
	PropertyName            = "name"
	PropertySecurityScopes  = "scopes"
	PropertyValidatorString = "validate"
)

type GleeceAnnotation = string

const (
	GleeceAnnotationTag             GleeceAnnotation = "Tag"
	GleeceAnnotationQuery           GleeceAnnotation = "Query"
	GleeceAnnotationPath            GleeceAnnotation = "Path"
	GleeceAnnotationBody            GleeceAnnotation = "Body"
	GleeceAnnotationHeader          GleeceAnnotation = "Header"
	GleeceAnnotationFormField       GleeceAnnotation = "FormField"
	GleeceAnnotationDeprecated      GleeceAnnotation = "Deprecated"
	GleeceAnnotationHidden          GleeceAnnotation = "Hidden"
	GleeceAnnotationSecurity        GleeceAnnotation = "Security"
	GleeceAnnotationRoute           GleeceAnnotation = "Route"
	GleeceAnnotationResponse        GleeceAnnotation = "Response"
	GleeceAnnotationDescription     GleeceAnnotation = "Description"
	GleeceAnnotationMethod          GleeceAnnotation = "Method"
	GleeceAnnotationErrorResponse   GleeceAnnotation = "ErrorResponse"
	GleeceAnnotationTemplateContext GleeceAnnotation = "TemplateContext"
)

type CommentSource string

// controller, route, schema, property

const (
	CommentSourceController CommentSource = "controller"
	CommentSourceRoute      CommentSource = "route"
	CommentSourceSchema     CommentSource = "schema"
	CommentSourceProperty   CommentSource = "property"
)
