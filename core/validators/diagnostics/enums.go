package diagnostics

type DiagnosticSeverity int

// Enumeration based on VS-Code's internal typing though values are different to adhere to LSP
const (
	/**
	 * Something not allowed by the rules of a language or other means.
	 */
	DiagnosticError DiagnosticSeverity = 1

	/**
	 * Something suspicious but allowed.
	 */
	DiagnosticWarning DiagnosticSeverity = 2

	/**
	 * Something to inform about but not a problem.
	 */
	DiagnosticInformation DiagnosticSeverity = 3

	/**
	 * Something to hint to a better way of doing it, like proposing
	 * a refactoring.
	 */
	DiagnosticHint DiagnosticSeverity = 4
)

type DiagnosticCode string

const (
	DiagAnnotationValueShouldNotExist          DiagnosticCode = "annotation-value-should-not-exist"
	DiagAnnotationValueMustExist               DiagnosticCode = "annotation-value-must-exist"
	DiagAnnotationValueInvalid                 DiagnosticCode = "annotation-value-invalid"
	DiagAnnotationPropertiesShouldNotExist     DiagnosticCode = "annotation-properties-should-not-exist"
	DiagAnnotationPropertiesMustExist          DiagnosticCode = "annotation-properties-must-exist"
	DiagAnnotationPropertiesInvalid            DiagnosticCode = "annotation-properties-invalid"
	DiagAnnotationPropertiesMissingKey         DiagnosticCode = "annotation-properties-missing-key"
	DiagAnnotationPropertiesUnknownKey         DiagnosticCode = "annotation-properties-unknown-key"
	DiagAnnotationPropertiesInvalidValueForKey DiagnosticCode = "annotation-properties-invalid-value-for-key"
	DiagAnnotationDescriptionShouldExist       DiagnosticCode = "annotation-description-should-exist"
	DiagMethodLevelTooManyOfAnnotation         DiagnosticCode = "method-too-many-of-annotation"
	DiagMethodLevelMissingRequiredAnnotation   DiagnosticCode = "method-missing-required-annotation"
	DiagMethodLevelAnnotationNotAllowed        DiagnosticCode = "method-annotation-not-allowed"
	DiagLinkerRouteMissingPath                 DiagnosticCode = "linker-route-missing-path-reference"
	DiagLinkerUnreferencedParameter            DiagnosticCode = "linker-unreferenced-parameter"
	DiagLinkerMultipleParameterRefs            DiagnosticCode = "linker-multiple-parameter-refs"
	DiagLinkerPathInvalidRef                   DiagnosticCode = "linker-path-annotation-invalid-reference"
	DiagLinkerDuplicatePathParamRef            DiagnosticCode = "linker-duplicate-path-param-ref"
	DiagLinkerDuplicatePathAliasRef            DiagnosticCode = "linker-duplicate-path-alias-ref"
	DiagLinkerIncompleteAttribute              DiagnosticCode = "linker-incomplete-attribute"
	DiagControllerLevelMissingTag              DiagnosticCode = "controller-missing-tag"
	DiagControllerLevelAnnotationNotAllowed    DiagnosticCode = "controller-annotation-not-allowed"
	DiagReceiverInvalidBody                    DiagnosticCode = "receiver-invalid-body"
	DiagReceiverParamNotPrimitive              DiagnosticCode = "receiver-parameter-not-primitive"
	DiagReceiverRetValsInvalidSignature        DiagnosticCode = "receiver-return-values-invalid-signature"
	DiagReceiverRetValsIsNotError              DiagnosticCode = "receiver-return-value-is-not-an-error"
	DiagFeatureUnsupported                     DiagnosticCode = "unsupported-feature"
)
