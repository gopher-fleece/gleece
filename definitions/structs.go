package definitions

import (
	"go/ast"
	"go/types"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/runtime"
)

type MethodHideOptions struct {
	Type      HideMethodType
	Condition string
}

type DeprecationOptions struct {
	Deprecated  bool
	Description string
}

type RestMetadata struct {
	Path string
}

type OrderedIdent struct {
	Ident   *ast.Ident
	Ordinal int
}

// This struct describes a function parameter's metadata without Gleece's additions
type ParamMeta struct {
	Ordinal   int
	Name      string
	IsContext bool
	TypeMeta  TypeMetadata
}

// This struct describes a function parameter's metadata with Gleece's additions.
type FuncParam struct {
	ParamMeta
	PassedIn           ParamPassedIn
	NameInSchema       string
	Description        string
	UniqueImportSerial uint64
	Validator          string
	Deprecation        DeprecationOptions
}

type FuncReturnValue struct {
	Ordinal int
	TypeMetadata
	UniqueImportSerial uint64
}

type AliasMetadata struct {
	Name      string   // e.g. LengthUnits
	AliasType string   // e.g. string, int, etc.
	Values    []string // e.g. ["Meter", "Kilometer"]
}

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

type TypeMetadata struct {
	Name                string
	PkgPath             string
	DefaultPackageAlias string
	Description         string
	Import              common.ImportType
	IsUniverseType      bool
	IsByAddress         bool
	SymbolKind          common.SymKind
	AliasMetadata       *AliasMetadata
}

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

type ErrorResponse struct {
	HttpStatusCode runtime.HttpStatusCode
	Description    string
}

type TemplateContext struct {
	Options     map[string]any
	Description string
}

type RouteMetadata struct {
	// The handler function's and operation name in the OpenAPI schema
	OperationId string

	// The HTTP verb this method expects (i.e., GET, POST etc.)
	HttpVerb HttpVerb

	// Controls whether the method is hidden in schema and when
	Hiding MethodHideOptions

	// Defines whether the method is considered deprecated
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

func (m RouteMetadata) GetValueReturnType() *TypeMetadata {
	if len(m.Responses) <= 1 {
		// No return value, just error
		return nil
	}

	// We're assuming the controller method's return signature always starts with the value
	return &m.Responses[0].TypeMetadata
}

func (m RouteMetadata) GetErrorReturnType() *TypeMetadata {
	if len(m.Responses) <= 1 {
		// If there is only one return value, it's the error
		return &m.Responses[0].TypeMetadata
	}

	// We're assuming the controller method's return signature always ends with the error
	return &m.Responses[1].TypeMetadata
}

type RouteSecurity struct {
	SecurityAnnotation []SecurityAnnotationComponent `json:"securityMethod" validate:"not_nil_array"` // AND between security methods
}

// SecurityAnnotationComponent is the schema-scopes parts of a security annotation;
// i.e., @Security(AND, [{name: "schema1", scopes: ["read", "write"]}, {name: "schema2", scopes: ["delete"]}])
type SecurityAnnotationComponent struct {
	SchemaName string   `json:"name" validate:"required,starts_with_letter"`
	Scopes     []string `json:"scopes" validate:"not_nil_array"`
}

// ControllerMetadata holds metadata pertaining to a specific controller as a full entity, including its receivers (routes)
type ControllerMetadata struct {
	Name         string
	PkgPath      string
	Tag          string
	Description  string
	RestMetadata RestMetadata
	Routes       []RouteMetadata

	// The default security schema/s used for the controller's operations.
	// May be overridden at the route level
	Security []RouteSecurity
}

type StructMetadata struct {
	Name        string
	PkgPath     string
	Description string
	Fields      []FieldMetadata
	Deprecation DeprecationOptions
}

type EnumMetadata struct {
	Name        string
	PkgPath     string
	Description string
	Values      []string
	Type        string
	Deprecation DeprecationOptions
}

type FieldMetadata struct {
	Name        string
	Type        string
	Description string
	Tag         string
	IsEmbedded  bool
	Deprecation *DeprecationOptions
}

type EnumValidator struct {
	Name   string
	Values []string
}

type Models struct {
	Structs        []StructMetadata
	Enums          []EnumMetadata
	EnumValidators []EnumValidator
}

type SpecGeneratorConfig struct {
	OutputPath string `json:"outputPath" validate:"required"`
}

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
	FieldName        string             `json:"fieldName" validate:"required,starts_with_letter"`
	Type             SecuritySchemeType `json:"type" validate:"required,security_schema_type"` // see SecuritySchemeType
	In               SecuritySchemeIn   `json:"in" validate:"required,security_schema_in"`     // see SecuritySchemeIn
	OpenIdConnectUrl string             `json:"openIdConnectUrl" validate:"omitempty,url"`
}

type OpenAPIContact struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Email string `json:"email" validate:"email"`
}

type OpenAPILicense struct {
	Name string `json:"name" validate:"required"`
	URL  string `json:"url"`
}

type OpenAPIInfo struct {
	Title          string          `json:"title" validate:"required"`
	Description    string          `json:"description"`
	TermsOfService string          `json:"termsOfService"`
	Contact        *OpenAPIContact `json:"contact"`
	License        *OpenAPILicense `json:"license"`
	Version        string          `json:"version" validate:"required"`
}

type OpenAPIGeneratorConfig struct {
	OpenAPI              string                       `json:"openapi" validate:"required,oneof=3.0.0 3.1.0"` // only 3.0.0 is fully supported
	Info                 OpenAPIInfo                  `json:"info" validate:"required"`
	BaseURL              string                       `json:"baseUrl" validate:"required,url"`
	SecuritySchemes      []SecuritySchemeConfig       `json:"securitySchemes" validate:"dive"`
	DefaultRouteSecurity *SecurityAnnotationComponent `json:"defaultSecurity"`
	SpecGeneratorConfig  SpecGeneratorConfig          `json:"specGeneratorConfig" validate:"required"`
}

type RoutesConfig struct {
	Engine                  RoutingEngineType   `json:"engine" validate:"required,oneof=gin echo mux fiber chi"`
	PackageName             string              `json:"packageName"`
	OutputPath              string              `json:"outputPath" validate:"required,filepath"`
	OutputFilePerms         string              `json:"outputFilePerms" validate:"regex=^(0?[0-7]{3})?$"`
	AuthorizationConfig     AuthorizationConfig `json:"authorizationConfig" validate:"required"`
	TemplateOverrides       map[string]string   `json:"templateOverrides"`
	TemplateExtensions      map[string]string   `json:"templateExtensions"`
	ValidateResponsePayload bool                `json:"validateResponsePayload"`
	SkipGenerateDateComment bool                `json:"skipGenerateDateComment"`
}

type AuthorizationConfig struct {
	AuthFileFullPackageName    string `json:"authFileFullPackageName" validate:"required,filepath"`
	EnforceSecurityOnAllRoutes bool   `json:"enforceSecurityOnAllRoutes"`
}

type CommonConfig struct {
	ControllerGlobs          []string `json:"controllerGlobs" validate:"omitempty,min=1"`
	AllowPackageLoadFailures bool     `json:"failOnAnyPackageLoadError"`
}

type ExperimentalConfig struct {
	ValidateTopLevelOnlyEnum bool `json:"validateTopLevelOnlyEnum"`
	GenerateEnumValidator    bool `json:"generateEnumValidator"`
}

type GleeceConfig struct {
	CommonConfig           CommonConfig           `json:"commonConfig" validate:"required"`
	RoutesConfig           RoutesConfig           `json:"routesConfig" validate:"required"`
	OpenAPIGeneratorConfig OpenAPIGeneratorConfig `json:"openapiGeneratorConfig" validate:"required"`
	ExperimentalConfig     ExperimentalConfig     `json:"experimentalConfig"` // TODO add docs
}

type Iterable interface {
	Elem() types.Type
	Underlying() types.Type
	String() string
}
