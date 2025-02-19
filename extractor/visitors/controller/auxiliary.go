package controller

import (
	"fmt"
	"go/ast"
	"strings"

	MapSet "github.com/deckarep/golang-set/v2"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/extractor"
	"github.com/gopher-fleece/gleece/extractor/annotations"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"github.com/gopher-fleece/runtime"
)

func (v *ControllerVisitor) getNextImportId() uint64 {
	value := v.importIdCounter
	v.importIdCounter++
	return value
}

func (v *ControllerVisitor) getAllSourceFiles() []*ast.File {
	result := []*ast.File{}
	for _, file := range v.sourceFiles {
		result = append(result, file)
	}
	return result
}

func (v ControllerVisitor) isAnErrorEmbeddingType(meta definitions.TypeMetadata) (bool, error) {
	v.enter(fmt.Sprintf("Type %s (%s)", meta.Name, meta.FullyQualifiedPackage))
	defer v.exit()

	if meta.Name == "error" {
		return true, nil
	}

	pkg := extractor.FilterPackageByFullName(v.packages, meta.FullyQualifiedPackage)
	embeds, err := extractor.DoesStructEmbedType(pkg, meta.Name, "", "error")
	if err != nil {
		return false, err
	}

	return embeds, nil
}

func (v ControllerVisitor) getMethodHideOpts(attributes *annotations.AnnotationHolder) definitions.MethodHideOptions {
	attr := attributes.GetFirst(annotations.AttributeHidden)
	if attr == nil {
		// No '@Hidden' attribute
		return definitions.MethodHideOptions{Type: definitions.HideMethodNever}
	}

	if attr.Properties == nil || len(attr.Properties) <= 0 {
		return definitions.MethodHideOptions{Type: definitions.HideMethodAlways}
	}

	// Technically a bit redundant since we know by length whether there's a condition defined
	// but nothing stops user from adding text to the comment so this mostly serves as a validation
	if len(attr.Value) <= 0 {
		// Standard '@Hidden' attribute; Always hide.
		return definitions.MethodHideOptions{Type: definitions.HideMethodAlways}
	}

	// A '@Hidden(condition)' attribute
	return definitions.MethodHideOptions{Type: definitions.HideMethodCondition, Condition: attr.Value}
}

func (v ControllerVisitor) getDeprecationOpts(attributes *annotations.AnnotationHolder) definitions.DeprecationOptions {
	attr := attributes.GetFirst(annotations.AttributeDeprecated)
	if attr == nil {
		return definitions.DeprecationOptions{Deprecated: false}
	}

	if len(attr.Description) <= 0 {
		// '@Deprecated' with no description
		return definitions.DeprecationOptions{Deprecated: true}
	}

	// '@Deprecated' with a comment
	return definitions.DeprecationOptions{Deprecated: true, Description: attr.Description}
}

func (v ControllerVisitor) getErrorResponseMetadata(attributes *annotations.AnnotationHolder) ([]definitions.ErrorResponse, error) {
	responseAttributes := attributes.GetAll(annotations.AttributeErrorResponse)

	responses := []definitions.ErrorResponse{}
	encounteredCodes := MapSet.NewSet[runtime.HttpStatusCode]()

	for _, attr := range responseAttributes {
		code, err := definitions.ConvertToHttpStatus(attr.Value)
		if err != nil {
			return responses, err
		}

		if encounteredCodes.ContainsOne(code) {
			logger.Warn(
				"Status code '%d' appears multiple time on a controller receiver. Ignoring. Original Comment: %s",
				code,
				attr,
			)
			continue
		}
		responses = append(responses, definitions.ErrorResponse{HttpStatusCode: code, Description: attr.Description})
		encounteredCodes.Add(code)
	}

	return responses, nil
}

func (v ControllerVisitor) getCustomContextMetadata(attributes *annotations.AnnotationHolder) []definitions.CustomContext {
	customAttributes := attributes.GetAll(annotations.AttributeCustomContext)

	customContexts := []definitions.CustomContext{}

	for _, attr := range customAttributes {

		customContexts = append(customContexts, definitions.CustomContext{
			Value:       attr.Value,
			Options:     attr.Properties,
			Description: attr.Description,
		})
	}

	return customContexts
}

func (v *ControllerVisitor) getResponseStatusCodeAndDescription(
	attributes *annotations.AnnotationHolder,
	hasReturnValue bool,
) (runtime.HttpStatusCode, string, error) {
	// Set the success attrib code based on whether function returns a value or only error (200 vs 204)
	attrib := attributes.GetFirst(annotations.AttributeResponse)
	if attrib == nil {
		if hasReturnValue {
			return runtime.StatusOK, "", nil
		}

		return runtime.StatusNoContent, "", nil
	}

	var statusCode runtime.HttpStatusCode
	if len(attrib.Value) > 0 {
		code, err := definitions.ConvertToHttpStatus(attrib.Value)
		if err != nil {
			return 0, "", v.frozenError(err)
		}
		statusCode = code
	}

	return statusCode, attrib.Description, nil
}

// For now, all params are required, later we will support nil for pointers and slices params
func appendParamRequiredValidation(validation *string, isPointer bool, paramPassedIn definitions.ParamPassedIn) string {
	// For a pointer, we do allow to be optional and it's pending user decision via validate tag
	// BUT, as openapi specification, path params are always required
	if isPointer && paramPassedIn != definitions.PassedInPath {
		return *validation
	}
	if validation == nil || *validation == "" {
		return "required"
	}

	// Split the validation string into individual tags
	tags := strings.Split(*validation, ",")

	// Check if "required" is already present
	for _, tag := range tags {
		if tag == "required" {
			return *validation
		}
	}

	// Append "required" to the validation string
	return *validation + ",required"
}
