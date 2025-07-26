package metadata

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
)

type IdProvider interface {
	GetIdForKey(key graphs.SymbolKey) uint64
}

// Go's insane package system forces us to get... creative.
type MetaCache interface {
	GetStruct(key graphs.SymbolKey) *StructMeta
	GetEnum(key graphs.SymbolKey) *EnumMeta
}

type EnumValueKind string

const (
	EnumValueKindString  EnumValueKind = "string"
	EnumValueKindInt     EnumValueKind = "int"
	EnumValueKindInt8    EnumValueKind = "int8"
	EnumValueKindInt16   EnumValueKind = "int16"
	EnumValueKindInt32   EnumValueKind = "int32"
	EnumValueKindInt64   EnumValueKind = "int64"
	EnumValueKindUInt    EnumValueKind = "uint"
	EnumValueKindUInt8   EnumValueKind = "uint8"
	EnumValueKindUInt16  EnumValueKind = "uint16"
	EnumValueKindUInt32  EnumValueKind = "uint32"
	EnumValueKindUInt64  EnumValueKind = "uint64"
	EnumValueKindFloat32 EnumValueKind = "float32"
	EnumValueKindFloat64 EnumValueKind = "float64"
	EnumValueKindBool    EnumValueKind = "bool"
)

func NewEnumValueKind(kind types.BasicKind) (EnumValueKind, error) {
	switch kind {
	case types.String:
		return EnumValueKindString, nil
	case types.Int:
		return EnumValueKindInt, nil
	case types.Int8:
		return EnumValueKindInt8, nil
	case types.Int16:
		return EnumValueKindInt16, nil
	case types.Int32:
		return EnumValueKindInt32, nil
	case types.Int64:
		return EnumValueKindInt64, nil
	case types.Uint:
		return EnumValueKindUInt, nil
	case types.Uint8:
		return EnumValueKindUInt8, nil
	case types.Uint16:
		return EnumValueKindUInt16, nil
	case types.Uint32:
		return EnumValueKindUInt32, nil
	case types.Uint64:
		return EnumValueKindUInt64, nil
	case types.Float32:
		return EnumValueKindFloat32, nil
	case types.Float64:
		return EnumValueKindFloat64, nil
	case types.Bool:
		return EnumValueKindBool, nil
	default:
		return "", fmt.Errorf("unsupported basic kind: %v", kind)
	}
}

type SymNodeMeta struct {
	Name        string
	Node        ast.Node
	SymbolKind  common.SymKind
	PkgPath     string
	Annotations *annotations.AnnotationHolder
	FVersion    *gast.FileVersion
}

type StructMeta struct {
	SymNodeMeta
	Fields []FieldMeta
}

func (s StructMeta) Reduce() definitions.StructMetadata {
	reducedFields := make([]definitions.FieldMetadata, len(s.Fields))
	for idx, field := range s.Fields {
		reducedFields[idx] = field.Reduce()
	}

	return definitions.StructMetadata{
		Name:        s.Name,
		PkgPath:     s.PkgPath,
		Description: annotations.GetDescription(s.Annotations),
		Fields:      reducedFields,
		Deprecation: GetDeprecationOpts(s.Annotations),
	}
}

type ConstMeta struct {
	SymNodeMeta
	Value any
	Type  TypeUsageMeta
}

type ControllerMeta struct {
	Struct    StructMeta
	Receivers []ReceiverMeta
}

func (m ControllerMeta) Reduce(
	gleeceConfig *definitions.GleeceConfig,
	metaCache MetaCache,
	syncedProvider IdProvider,
) (definitions.ControllerMetadata, error) {
	// Parse any explicit Security annotations
	security, err := GetSecurityFromContext(m.Struct.Annotations)
	if err != nil {
		return definitions.ControllerMetadata{}, err
	}

	// If there are no explicitly defined securities, check for inherited ones
	if len(security) <= 0 {
		logger.Debug("Controller %s does not have explicit security; Using user-defined defaults", m.Struct.Name)
		security = GetDefaultSecurity(gleeceConfig)
	}

	var reducedReceivers []definitions.RouteMetadata
	for _, rec := range m.Receivers {
		reduced, err := rec.Reduce(metaCache, syncedProvider, security)
		if err != nil {
			logger.Error("Failed to reduce receiver '%s' of controller '%s' - %w", rec.Name, m.Struct.Name, err)
			return definitions.ControllerMetadata{}, err
		}
		reducedReceivers = append(reducedReceivers, reduced)
	}

	meta := definitions.ControllerMetadata{
		Name:        m.Struct.Name,
		PkgPath:     m.Struct.PkgPath,
		Tag:         m.Struct.Annotations.GetFirstValueOrEmpty(annotations.GleeceAnnotationTag),
		Description: m.Struct.Annotations.GetDescription(),
		RestMetadata: definitions.RestMetadata{
			Path: m.Struct.Annotations.GetFirstValueOrEmpty(annotations.GleeceAnnotationRoute),
		},
		Routes:   reducedReceivers,
		Security: security,
	}

	return meta, nil
}

type ReceiverMeta struct {
	SymNodeMeta
	Params  []FuncParam
	RetVals []FuncReturnValue
}

func (m ReceiverMeta) Reduce(
	metaCache MetaCache,
	syncedProvider IdProvider,
	parentSecurity []definitions.RouteSecurity,
) (definitions.RouteMetadata, error) {

	verbAnnotation := m.Annotations.GetFirst(annotations.GleeceAnnotationMethod)
	if verbAnnotation == nil || verbAnnotation.Value == "" {
		// Not ideal- we'd like to separate visitation, reduction and validation but typing currently doesn't
		// cleanly allow it so we have to embed a bit of validation in a few other places as well
		return definitions.RouteMetadata{}, fmt.Errorf("receiver %s has not @Method annotation", m.Name)
	}

	security, err := GetRouteSecurityWithInheritance(m.Annotations, parentSecurity)
	if err != nil {
		return definitions.RouteMetadata{}, err
	}

	templateCtx, err := GetTemplateContextMetadata(m.Annotations)
	if err != nil {
		return definitions.RouteMetadata{}, err
	}

	hasReturnValue := len(m.RetVals) > 1

	responses := []definitions.FuncReturnValue{}
	for _, fRetVal := range m.RetVals {
		response, err := fRetVal.Reduce(metaCache, syncedProvider)
		if err != nil {
			return definitions.RouteMetadata{}, err
		}
		responses = append(responses, response)
	}

	reducedParams := []definitions.FuncParam{}
	for _, param := range m.Params {
		reducedParam, err := param.Reduce(metaCache, syncedProvider)
		if err != nil {
			return definitions.RouteMetadata{}, err
		}
		reducedParams = append(reducedParams, reducedParam)
	}

	successResponseCode, successResponseDescription, err := GetResponseStatusCodeAndDescription(m.Annotations, hasReturnValue)
	if err != nil {
		return definitions.RouteMetadata{}, err
	}

	errorResponses, err := GetErrorResponses(m.Annotations)
	if err != nil {
		return definitions.RouteMetadata{}, err
	}

	return definitions.RouteMetadata{
		OperationId: m.Name,
		HttpVerb:    definitions.HttpVerb(verbAnnotation.Value),
		Hiding:      GetMethodHideOpts(m.Annotations),
		Deprecation: GetDeprecationOpts(m.Annotations),
		Description: m.Annotations.GetDescription(),
		RestMetadata: definitions.RestMetadata{
			Path: m.Annotations.GetFirstValueOrEmpty(annotations.GleeceAnnotationRoute),
		},
		HasReturnValue:      hasReturnValue,
		RequestContentType:  definitions.ContentTypeJSON, // Hardcoded for now, should be supported via annotations later on
		ResponseContentType: definitions.ContentTypeJSON, // Hardcoded for now, should be supported via annotations later on
		Security:            security,
		TemplateContext:     templateCtx,
		ResponseSuccessCode: successResponseCode,
		ResponseDescription: successResponseDescription,
		FuncParams:          reducedParams,
		Responses:           responses,
		ErrorResponses:      errorResponses,
	}, nil
}

type FuncParam struct {
	SymNodeMeta
	Ordinal int
	Type    TypeUsageMeta
}

func (v FuncParam) Reduce(metaCache MetaCache, syncedProvider IdProvider) (definitions.FuncParam, error) {
	typeMeta, err := v.Type.Resolve(metaCache)
	if err != nil {
		return definitions.FuncParam{}, err
	}

	var nameInSchema string
	var passedIn definitions.ParamPassedIn
	var validator string

	isContext := v.Type.Name == "Context" && v.Type.PkgPath == "context"

	if !isContext {
		nameInSchema, err = GetParameterSchemaName(v.Name, v.Annotations)
		if err != nil {
			return definitions.FuncParam{}, err
		}

		passedIn, err = GetParamPassedIn(v.Name, v.Annotations)
		if err != nil {
			return definitions.FuncParam{}, err
		}

		validator, err = GetParamValidator(v.Name, v.Annotations, passedIn, v.Type.IsByAddress())
		if err != nil {
			return definitions.FuncParam{}, err
		}
	}

	typeRef, err := v.Type.GetBaseTypeRefKey()
	if err != nil {
		return definitions.FuncParam{}, err
	}

	// Find the parameter's attribute in the receiver's annotations
	var paramDescription string
	paramAttrib := v.Annotations.FindFirstByValue(v.Name)
	if paramAttrib != nil {
		// Note that nil here is not valid and should be rejected at the validation stage
		paramDescription = paramAttrib.Description
	}

	return definitions.FuncParam{
		ParamMeta: definitions.ParamMeta{
			Name:      v.Name,
			Ordinal:   v.Ordinal,
			TypeMeta:  typeMeta,
			IsContext: isContext,
		},
		PassedIn:           passedIn,
		NameInSchema:       nameInSchema,
		Description:        paramDescription,
		UniqueImportSerial: syncedProvider.GetIdForKey(typeRef),
		Validator:          validator,
		Deprecation:        GetDeprecationOpts(v.Annotations),
	}, nil
}

type FuncReturnValue struct {
	SymNodeMeta
	Ordinal int
	Type    TypeUsageMeta
}

func (v FuncReturnValue) Reduce(metaCache MetaCache, syncedProvider IdProvider) (definitions.FuncReturnValue, error) {
	typeMeta, err := v.Type.Resolve(metaCache)
	if err != nil {
		return definitions.FuncReturnValue{}, err
	}

	typeRef, err := v.Type.GetBaseTypeRefKey()
	if err != nil {
		return definitions.FuncReturnValue{}, err
	}

	return definitions.FuncReturnValue{
		Ordinal:            v.Ordinal,
		UniqueImportSerial: syncedProvider.GetIdForKey(typeRef),
		TypeMetadata:       typeMeta,
	}, nil
}

type EnumValueDefinition struct {
	// The enum's value definition node meta, e.g. EnumValueA SomeEnumType = "Abc"
	SymNodeMeta
	Value any // e.g. ["Meter", "Kilometer"]
	// TODO: An exact textual representation of the value. For example "1 << 2"
	//RawLiteralValue string
}

type EnumMeta struct {
	// The enum's type definition's node meta e.g. type SomeEnumType string
	SymNodeMeta
	ValueKind EnumValueKind // e.g. string, int, etc.
	Values    []EnumValueDefinition
}

func (e EnumMeta) Reduce() definitions.EnumMetadata {
	stringifiedValues := common.Map(e.Values, func(value EnumValueDefinition) string {
		return fmt.Sprintf("%v", value.Value)
	})

	return definitions.EnumMetadata{
		Name:        e.Name,
		PkgPath:     e.PkgPath,
		Description: annotations.GetDescription(e.Annotations),
		Values:      stringifiedValues,
		Type:        string(e.ValueKind),
		Deprecation: GetDeprecationOpts(e.Annotations),
	}
}

type TypeUsageMeta struct {
	SymNodeMeta
	Import common.ImportType
	Layers []TypeLayer
}

func (t TypeUsageMeta) GetBaseTypeRefKey() (graphs.SymbolKey, error) {
	if len(t.Layers) == 0 {
		return graphs.SymbolKey{}, fmt.Errorf("TypeUsageMeta has no layers")
	}
	baseRef := t.Layers[len(t.Layers)-1].BaseTypeRef
	if baseRef == nil {
		return graphs.SymbolKey{}, fmt.Errorf("BaseTypeRef is nil on last TypeLayer")
	}
	return *baseRef, nil
}

func (t TypeUsageMeta) GetArrayLayersString() string {
	// Currently we only use arrays for spec generation
	arrayCount := 0
	for _, layer := range t.Layers {
		if layer.Kind == TypeLayerKindArray {
			arrayCount++
		}
	}

	return strings.Repeat("[]", arrayCount)
}

func (t TypeUsageMeta) Resolve(metaCache MetaCache) (definitions.TypeMetadata, error) {
	typeRef, err := t.GetBaseTypeRefKey()
	if err != nil {
		return definitions.TypeMetadata{}, err
	}

	underlyingEnum := metaCache.GetEnum(typeRef)

	alias := definitions.AliasMetadata{}
	if underlyingEnum != nil {
		alias.Name = underlyingEnum.Name
		alias.AliasType = string(underlyingEnum.ValueKind)

		values := []string{}
		for _, v := range underlyingEnum.Values {
			values = append(values, fmt.Sprintf("%v", v.Value))
		}
		alias.Values = values
	}

	description := ""
	if t.Annotations != nil {
		description = t.Annotations.GetDescription()
	}

	// Join the actual type name with its "[]" prefixes, as necessary.
	// Ugly, but the spec generator uses that - for now.
	name := t.GetArrayLayersString() + t.Name

	return definitions.TypeMetadata{
		Name:                name,
		PkgPath:             t.PkgPath,
		DefaultPackageAlias: gast.GetDefaultPkgAliasByName(t.PkgPath),
		Description:         description,
		Import:              t.Import,
		IsUniverseType:      t.PkgPath == "" && gast.IsUniverseType(t.Name),
		IsByAddress:         t.IsByAddress(),
		SymbolKind:          t.SymbolKind,
		AliasMetadata:       &alias,
	}, nil
}

type FieldMeta struct {
	SymNodeMeta
	Type       TypeUsageMeta
	IsEmbedded bool
}

func (f FieldMeta) Reduce() definitions.FieldMetadata {
	fieldNode, ok := f.Node.(*ast.Field)
	if !ok {
		// Reduce has a pretty nice signature so pretty reluctant to hole it with an added error
		panic("field %s has a non-field node type")
	}

	var tag string
	if fieldNode != nil && fieldNode.Tag != nil {
		tag = strings.Trim(fieldNode.Tag.Value, "`")
	}

	decoratedType := f.Type.GetArrayLayersString() + f.Type.Name
	return definitions.FieldMetadata{
		Name:        f.Name,
		Type:        decoratedType,
		Description: annotations.GetDescription(f.Annotations),
		Tag:         tag,
		IsEmbedded:  f.IsEmbedded,
		Deprecation: common.Ptr(GetDeprecationOpts(f.Annotations)),
	}
}

func (m TypeUsageMeta) IsUniverseType() bool {
	return gast.IsUniverseType(m.Name)
}

func (m TypeUsageMeta) IsByAddress() bool {
	_, isStar := m.Node.(*ast.StarExpr)
	return isStar
}
