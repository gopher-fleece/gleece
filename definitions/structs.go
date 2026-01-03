package definitions

import (
	"github.com/gopher-fleece/gleece/v2/common"
	"github.com/gopher-fleece/runtime"
)

// Describes an API endpoint's visibility in the OpenAPI schema

type MethodHideOptions struct {
	// An API method's visibility in the generated OpenAPI schema
	Type HideMethodType
	// CURRENTLY UNIMPLEMENTED
	//
	// A condition, which, when met, will cause the method to be omitted from the OpenAPI schema (i.e., hidden)
	//
	// Relevant only for HideMethodTypeCondition
	Condition string
}

// Contains information regarding a schema entity's deprecation
type DeprecationOptions struct {
	// Indicates whether the entity is deprecated
	Deprecated bool
	// A description to be included in the schema
	Description string
}

// Additional metadata related to an a controller or API endpoint, such as it's URL
type RestMetadata struct {
	// The entity's API path.
	//
	// Note that final API path depends on whether the entity is a 'leaf' node, i.e.,
	// a controller's path may be extended by the specific route's path
	Path string
}

// Describes a method's parameter base metadata
type ParamMeta struct {
	// The parameter's position
	Ordinal int
	// The parameter's name
	Name string
	// Indicates whether the parameter is a context.Context
	IsContext bool
	// Parameter type information
	TypeMeta TypeMetadata
}

// Describes a function parameter's metadata including any information required for code generation
type FuncParam struct {
	// Parameter base metadata
	ParamMeta
	// The venue in which the parameter is to be passed (e.g., Query, Header, etc.)
	PassedIn ParamPassedIn
	// The parameter's name in the resulting OpenAPI schema
	NameInSchema string
	// The parameter's description
	Description string
	// A unique import serial number.
	//
	// This serial is used to prevent name conflict between packages that share the same 'default' name in the generated code
	UniqueImportSerial uint64
	// The parameter's validator, if any, as obtained from its tag
	Validator string
	// Information about whether this parameter has been deprecated and why
	Deprecation DeprecationOptions
}

// Describes a method's return value
type FuncReturnValue struct {
	// The value's position in the return signature
	Ordinal int
	// Return value type information
	TypeMetadata
	// A unique import serial number.
	//
	// This serial is used to prevent name conflict between packages that share the same 'default' name in the generated code
	UniqueImportSerial uint64
}

// Describes an enumeration or an alias
type AliasMetadata struct {
	// The enumeration/alias name, e.g. "LengthUnits'
	Name string
	// The type of the enumeration/alias values, e.g. "string", "int" etc.
	AliasType string
	// The enumeration/alias values, e.g., "Meter", "Kilometer" etc.
	Values []string
}

// Equals returns a boolean indicating whether the given enumeration is equal to the current one
func (a AliasMetadata) Equals(other AliasMetadata) bool {
	if a.Name != other.Name {
		return false
	}

	if a.AliasType != other.AliasType {
		return false
	}

	if len(a.Values) != len(other.Values) {
		return false
	}

	for i := range a.Values {
		if a.Values[i] != other.Values[i] {
			return false
		}
	}
	return true
}

// Describes a type alias such as
//
//	type Alias string
//
// or
//
//	type Alias = string
type NakedAliasMetadata struct {
	// The alias's name, e.g. "StringAlias"
	Name string
	// The full package path in which the alias is defined
	PkgPath string
	// The alias's underlying type name
	Type string
	// A description for the alias declaration itself
	Description string
	// Information about whether the alias has been deprecated and why
	Deprecation DeprecationOptions
}

// Describe's a type's usage site
type TypeMetadata struct {
	// The type's name
	Name string
	// The full package path in which the type itself is defined
	PkgPath string
	// The default import alias for the type's package
	DefaultPackageAlias string
	// A description for the type declaration itself
	Description string
	// The way in which the type is imported, in the context of a usage
	Import common.ImportType
	// Indicates whether the type is a built-in Universe type
	IsUniverseType bool
	// Indicates whether the type is being used by-address
	IsByAddress bool
	// The type's symbol kind
	SymbolKind common.SymKind
	// Extended Enum/Alias metadata relating to the type.
	//
	// Enums and Aliases fulfil the 'TypeMetadata' struct but also have the AliasMetadata section that
	// contains their their value type and exact values
	AliasMetadata *AliasMetadata
}

// Equals returns a boolean indicating whether the given type metadata is equal to the current one
func (t TypeMetadata) Equals(other TypeMetadata) bool {
	if t.Name != other.Name {
		return false
	}
	if t.PkgPath != other.PkgPath {
		return false
	}
	if t.DefaultPackageAlias != other.DefaultPackageAlias {
		return false
	}
	if t.Description != other.Description {
		return false
	}
	if t.Import != other.Import {
		return false
	}
	if t.IsUniverseType != other.IsUniverseType {
		return false
	}
	if t.IsByAddress != other.IsByAddress {
		return false
	}
	if t.SymbolKind != other.SymbolKind {
		return false
	}

	if (t.AliasMetadata == nil) != (other.AliasMetadata == nil) {
		return false
	}

	if t.AliasMetadata != nil && !t.AliasMetadata.Equals(*other.AliasMetadata) {
		return false
	}

	return true
}

// An error response from an API endpoint, in the context of the OpenAPI schema
type ErrorResponse struct {
	// The response's HTTP status code
	HttpStatusCode runtime.HttpStatusCode
	// A description for the response.
	//
	// Should contain useful information such as when consumers should expect this response
	Description string
}

// Additional context to be made available when rendering the routing template.
//
// This is used via the @TemplateContext annotation to allow for injection of custom behaviors at the route level.
//
// An example could be something like:
//
//	// @TemplateContext(DEBUG, {enabled: true, name: "Something"}) Enables debug output for the specific route
//
// Which can be used, in the template itself to enable debug logs using a specific name.
//
// This feature is mostly aimed at enterprise users which often require extensive and unforeseen customizations
type TemplateContext struct {
	// A map of data to include in the template renderer context.
	//
	// This data can be referenced directly from the template code.
	Options map[string]any
	// A description for the context.
	//
	// Currently unused.
	Description string
}

// Encapsulates information about a particular API endpoint.
//
// This structure is used to represent an API endpoint and contains all information necessary to
// generate both routing and schema for it.
type RouteMetadata struct {
	// The handler function's and operation name in the OpenAPI schema
	OperationId string

	// The HTTP verb this method expects (i.e., GET, POST etc.)
	HttpVerb HttpVerb

	// Controls whether the method is hidden in schema and when
	Hiding MethodHideOptions

	// Information about whether the method has been deprecated and why
	Deprecation DeprecationOptions

	// The operation's description
	Description string

	// Additional metadata related to the operation such as it's URL
	RestMetadata RestMetadata

	// Metadata on the handler function's parameters
	FuncParams []FuncParam

	// Metadata on the handler function's return values
	Responses []FuncReturnValue

	// Indicates whether the operation returns a value on success.
	//
	// Note that the framework enforces at-least an error return value from all controller methods
	HasReturnValue bool

	// A description for success responses
	ResponseDescription string

	// The HTTP code expected to be returned from a successful call
	//
	// TODO: Needs to be an array
	ResponseSuccessCode runtime.HttpStatusCode

	// Metadata on the type of errors that may be returned from the operation
	ErrorResponses []ErrorResponse

	// The expected request content type.
	//
	// Currently hard-coded to application/json.
	RequestContentType ContentType

	// The expected response content type.
	//
	// Currently hard-coded to application/json.
	ResponseContentType ContentType

	// The security schema/s used for the operation
	Security []RouteSecurity // OR between security routes

	// Custom template context for the operation, provided by the route developer, used template extension/override
	TemplateContext map[string]TemplateContext
}

// GetValueReturnType returns the TypeMetadata for the API endpoint's primary (success) response type.
//
// Returns nil if the endpoint returns only an error.
// Examples:
//
// For:
//
//	func (r *Controller) OperationA() error { ... }
//
// returns nil.
//
// For:
//
//	func (r *Controller) OperationB() (string, error) { ... }
//
// returns the type metadata for 'string'.
func (m RouteMetadata) GetValueReturnType() *TypeMetadata {
	if len(m.Responses) <= 1 {
		// No return value, just error
		return nil
	}

	// We're assuming the controller method's return signature always starts with the value,
	// as defined at the logic level and validated during analysis/validation
	return &m.Responses[0].TypeMetadata
}

// GetErrorReturnType returns the TypeMetadata for the error return value of the API endpoint
func (m RouteMetadata) GetErrorReturnType() *TypeMetadata {
	if len(m.Responses) <= 1 {
		// If there is only one return value, it's the error
		return &m.Responses[0].TypeMetadata
	}

	// We're assuming the controller method's return signature always ends with the error
	return &m.Responses[1].TypeMetadata
}

// Describes the security that applies to an API endpoint
type RouteSecurity struct {
	// The security annotations for the endpoint.
	//
	// Note that a route may have multiple route securities and each one may have multiple components.
	//
	// In the context of SecurityAnnotationComponent, the relation is 'AND'
	SecurityAnnotation []SecurityAnnotationComponent `json:"securityMethod" validate:"not_nil_array"` // AND between security methods
}

// SecurityAnnotationComponent is the schema-scopes parts of a security annotation;
// i.e., @Security(AND, [{name: "schema1", scopes: ["read", "write"]}, {name: "schema2", scopes: ["delete"]}])
type SecurityAnnotationComponent struct {
	// The security schema's name.
	// This schema will be included in the generated OpenAPI schema and may be used by route @Security annotations
	SchemaName string `json:"name" validate:"required,starts_with_letter"`

	// The scopes passed along this security schema.
	//
	// Scopes are a way to allow for authorization granularity & permissions - for example,
	// some endpoints may require only READ permissions while others require WRITE.
	// In this way, the schema is akin to the 'type' of authentication and the scopes control authorization.
	Scopes []string `json:"scopes" validate:"not_nil_array"`
}

// ControllerMetadata holds metadata pertaining to a specific controller as a full entity, including its receivers (routes)
type ControllerMetadata struct {
	// The controller's name
	Name string
	// The controller's full package path
	PkgPath string

	// The controller's OpenAPI tag
	//
	// Provided using the @Tag annotation, tags are emitted to the OpenAPI schemas
	// and are mostly used by visualizers to group API endpoints in the UI.
	Tag string

	// The controller's description
	//
	// Provided via the @Description annotation, if one exists, or the first contiguous lines of the struct's standard Go description
	Description string

	// Additional controller metadata
	RestMetadata RestMetadata

	// The routes (API endpoints) exposed by this controller
	Routes []RouteMetadata

	// The default security schema/s used for the controller's operations.
	// Inherited from configuration and may be overridden at either controller or route levels
	Security []RouteSecurity
}

// Encapsulates information about a particular structure.
//
// This structure is generally used to represent and generate an OpenAPI model
type StructMetadata struct {
	// The structure's name
	Name string

	// The structure's full package path
	PkgPath string

	// The structure's description
	//
	// Provided via the @Description annotation, if one exists, or the first contiguous lines of the struct's standard Go description
	Description string

	// The structure's fields
	Fields []FieldMetadata

	// Information about whether the structure has been deprecated and why
	Deprecation DeprecationOptions
}

// Clone returns a copy of the structure's metadata
func (s StructMetadata) Clone() StructMetadata {
	fields := make([]FieldMetadata, 0, len(s.Fields))
	for _, field := range s.Fields {
		var deprecationOpts *DeprecationOptions
		if field.Deprecation != nil {
			deprecationOpts = &DeprecationOptions{
				Deprecated:  field.Deprecation.Deprecated,
				Description: field.Deprecation.Description,
			}
		}

		fields = append(fields, FieldMetadata{
			Name:        field.Name,
			Type:        field.Type,
			Description: field.Description,
			Tag:         field.Tag,
			IsEmbedded:  field.IsEmbedded,
			Deprecation: deprecationOpts,
		})
	}

	return StructMetadata{
		Name:        s.Name,
		PkgPath:     s.PkgPath,
		Description: s.Description,
		Fields:      fields,
		Deprecation: s.Deprecation,
	}
}

// Describes an enumeration and its values
type EnumMetadata struct {
	// The name of the enumeration type.
	//
	// e.g.,
	// 	"SomeEnum"
	// for
	// 	type SomeEnum = string
	//	const V1 SomeEnum = "abc"
	Name string

	// The enum's full package path
	PkgPath string

	// The enum's declaration's description
	//
	// Provided via the @Description annotation, if one exists, or the first contiguous lines of the type's standard Go description
	Description string

	// The enum's values in string form
	Values []string

	// The type of the enum's values, e.g. string, int float etc.
	Type string

	// Information about whether the enumeration has been deprecated
	Deprecation DeprecationOptions
}

// Describe's a structure's field
type FieldMetadata struct {
	// The field name.
	//
	// For anonymous fields, this value is the type's name
	Name string

	// The field type's name
	Type string

	// The field's description
	//
	// Provided via the @Description annotation, if one exists, or the first contiguous lines of the field's standard Go description
	Description string

	// A Go tag for the field.
	//
	// Field tags control the final naming for the field, at the API level as well as (optional) validation.
	//
	// Example:
	//	SomeField `json:"someOtherName" validate:"required"`
	Tag string

	// Specified whether the field is an embedded one
	IsEmbedded bool

	// Information about whether the field has been deprecated
	Deprecation *DeprecationOptions
}

// Contains models for the OpenAPI schema
type Models struct {
	// Struct-type models
	Structs []StructMetadata

	// Enum-type models, that is, a representation for enumerations like:
	// 	type SomeEnum = string
	//	const V1 SomeEnum = "abc"
	Enums []EnumMetadata

	// Alias-type models, e.g.
	//	type StringAlias = string
	// or
	//	type StringAlias string
	Aliases []NakedAliasMetadata
}

// Configuration for the OpenAPI schema generator
type SpecGeneratorConfig struct {
	// The path output the OpenAPI schema to
	OutputPath string `json:"outputPath" validate:"required"`
}

// Describes an OAuth authorization flow
type OAuthFlow struct {
	Extensions       map[string]any    `json:"-" yaml:"-"`
	AuthorizationURL string            `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty" yaml:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes" yaml:"scopes"` // required
}

type OAuthFlows struct {
	Extensions map[string]any `json:"-" yaml:"-"`

	Implicit          *OAuthFlow `json:"implicit,omitempty" yaml:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty" yaml:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
}

type SecuritySchemeConfig struct {
	Description      string             `json:"description" validate:"required"`
	SecurityName     string             `json:"name" validate:"required,starts_with_letter"`
	Scheme           HttpAuthScheme     `json:"scheme" validate:"omitempty,oneof=basic bearer digest hoba mutual negotiate oauth scram-sha-1 scram-sha-256 vapid"`
	Flows            *OAuthFlows        `json:"flows"`
	FieldName        string             `json:"fieldName" validate:"starts_with_letter"`
	Type             SecuritySchemeType `json:"type" validate:"required,security_schema_type"` // see SecuritySchemeType
	In               SecuritySchemeIn   `json:"in" validate:"security_schema_in"`              // see SecuritySchemeIn
	OpenIdConnectUrl string             `json:"openIdConnectUrl" validate:"omitempty,url"`
}

// Contact information to be included with the OpenAPI schema
type OpenAPIContact struct {
	// The contact's name
	Name string `json:"name"`
	// A contact URL, commonly a product or documentation page
	URL string `json:"url"`
	// The contact's e-mail
	Email string `json:"email" validate:"email"`
}

// Information about the API's license
type OpenAPILicense struct {
	// The license's name, e.g. "Apache 2.0"
	Name string `json:"name" validate:"required"`
	// A URL in which extended license information can be found, e.g. "http://www.apache.org/licenses/LICENSE-2.0.html"
	URL string `json:"url"`
}

// General information to be included in the OpenAPI schema
type OpenAPIInfo struct {
	// The API's title
	Title string `json:"title" validate:"required"`
	// The API's description
	Description string `json:"description"`
	// Terms of Service, if any
	TermsOfService string `json:"termsOfService"`
	// Contact information
	Contact *OpenAPIContact `json:"contact"`
	// License information
	License *OpenAPILicense `json:"license"`
	// API version
	Version string `json:"version" validate:"required"`
}

// Configuration for the OpenAPI generator
type OpenAPIGeneratorConfig struct {
	// The OpenAPI schema version, e.g., "3.0.0"
	OpenAPI string `json:"openapi" validate:"required,oneof=3.0.0 3.1.0"`
	// General information to include in the schema
	Info OpenAPIInfo `json:"info" validate:"required"`
	// The API's base URL.
	//
	// Final API endpoint URL is comprised of
	//	`{BASE_URL}/{CONTROLLER_URL}/{ROUTE_URL}`
	BaseURL string `json:"baseUrl" validate:"required,url"`
	// The security schema definitions for the API.
	//
	// Controllers and routes may specify which of the schemas they adhere to
	SecuritySchemes []SecuritySchemeConfig `json:"securitySchemes" validate:"dive"`
	// The default security for routes that do not have any explicit or inherited @Security annotations.
	//
	// This setting is used to ensure all API endpoints are secured
	DefaultRouteSecurity *SecurityAnnotationComponent `json:"defaultSecurity"`
	// Configuration for the OpenAPI schema generator
	SpecGeneratorConfig SpecGeneratorConfig `json:"specGeneratorConfig" validate:"required"`
}

// Configuration for the routing code generator
type RoutesConfig struct {
	// The underlying routing engine to use
	Engine RoutingEngineType `json:"engine" validate:"required,oneof=gin echo mux fiber chi"`
	// The package name for the generated code
	PackageName string `json:"packageName"`
	// The path to output the generated routing code to
	OutputPath string `json:"outputPath" validate:"required,filepath"`
	// The permissions for the generated routing code.
	//
	// Not relevant for Windows-based machines
	OutputFilePerms string `json:"outputFilePerms" validate:"regex=^(0?[0-7]{3})?$"`
	// Configuration relating to API authentication/authorization
	AuthorizationConfig AuthorizationConfig `json:"authorizationConfig" validate:"required"`
	// A map of template-name/file-path.
	//
	// This property is used to specify overrides for the built-in routing templates
	// to allow for in-depth customization of the generated routing code.
	TemplateOverrides map[string]string `json:"templateOverrides"`

	// A map of template-hook-name/file-path.
	//
	// This property is used to insert 'hooks' into pre-defined locations in the generated routing code.
	// Examples of such places include before a request is processed, after a request finishes etc.
	//
	// Each routing engine template exposes its own hooks; For specifics, consult the 'embeds.go' file for the relevant engine (e.g. /gleece/generator/templates/gin/embeds.go)
	TemplateExtensions map[string]string `json:"templateExtensions"`
	// Determines whether API responses are validated before being returned to consumers.
	//
	// Response validation ensures the returned payload matches the expected schema.
	// This includes and constraints imposed via validator tags.
	ValidateResponsePayload bool `json:"validateResponsePayload"`
	// Determines whether the generated code will contain a generation timestamp.
	//
	// Generation timestamps, while handy, can create 'fake' changes to the generated routing code.
	SkipGenerateDateComment bool `json:"skipGenerateDateComment"`
}

// Configuration pertaining to the API authentication/authorization
type AuthorizationConfig struct {
	// The full package name for the file containing the authentication middleware
	AuthFileFullPackageName string `json:"authFileFullPackageName" validate:"required,filepath"`
	// Determines whether the generation pipeline should fail if any route does not have any security.
	//
	// This feature meant to ensure all routes are secured by default regardless of developer familiarity and care
	EnforceSecurityOnAllRoutes bool `json:"enforceSecurityOnAllRoutes"`
}

// Common Gleece pipeline configurations
type CommonConfig struct {
	// Glob expression/s used to locate controller files.
	//
	// Expression evaluation is done by the bmatcuk/doublestar package.
	ControllerGlobs []string `json:"controllerGlobs" validate:"omitempty,min=1"`
	// Determines whether any package load failures trigger a pipeline failure.
	//
	// Generally, when Gleece fails to load a package, it stops the pipeline.
	//
	// In some scenarios, especially those involving generated code that may not be present this behavior is unwanted.
	//
	// Setting this flag changes the behavior to 'warn' instead of 'error'.
	AllowPackageLoadFailures bool `json:"allowPackageLoadFailures"`
}

// Configuration of experimental or otherwise advanced features
type ExperimentalConfig struct {
	// Controls automatic validation of top-level (*only*) enums.
	//
	// I/O validation is delegated to go-playground/validator which does not, to the date of writing,
	// support automatic enum validation.
	//
	// This means that in order to validate enumeration values, developers must explicitly use a 'one-of' directive in a Go tag.
	//
	// To help make this process less manual, this feature allows validating top-level enumerations, i.e., those passed
	// directly via a query, URL an so forth.
	//
	// Note that this *DOES NOT* validate enumerations in fields!
	//
	// Use with care - usage may create a false sense of security in which developers might think they're fully covered when they're not.
	ValidateTopLevelOnlyEnum bool `json:"validateTopLevelOnlyEnum"`
	// Controls support for generated, tag based enum validators.
	//
	// This option effectively generates validators for all known enumerations.
	// Generated validators will still need to be manually referenced via a tag.
	// The generated validators's naming convention is the enumeration's name in snake-case plus an _enum suffix, e.g.
	//	LengthUnits
	// becomes
	//	length_units_enum
	//
	// To use these generated validators:
	//
	//	type SomeStruct struct {
	//		Length LengthUnits `validate:"required,length_units_enum"`
	//	}
	//
	GenerateEnumValidator bool `json:"generateEnumValidator"`
}

// Gleece's main configuration
type GleeceConfig struct {
	// Common Gleece pipeline configurations
	CommonConfig CommonConfig `json:"commonConfig" validate:"required"`
	// Configuration for the routing code generator
	RoutesConfig RoutesConfig `json:"routesConfig" validate:"required"`
	// Configuration for the OpenAPI generator
	OpenAPIGeneratorConfig OpenAPIGeneratorConfig `json:"openapiGeneratorConfig" validate:"required"`
	// Configuration of experimental or otherwise advanced features
	ExperimentalConfig ExperimentalConfig `json:"experimentalConfig"` // TODO add docs
}
