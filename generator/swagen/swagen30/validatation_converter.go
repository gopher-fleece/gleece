package swagen30

import (
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gopher-fleece/gleece/v2/generator/swagen/swagtool"
	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
)

func BuildSchemaValidation(schema *openapi3.SchemaRef, validationString string, fieldInterface string) {
	// Parse and apply validation rules from the Validator field
	validationRules := strings.Split(validationString, ",")
	for _, rule := range validationRules {
		parts := strings.SplitN(rule, "=", 2)
		ruleName := parts[0]
		var ruleValue string
		if len(parts) > 1 {
			ruleValue = parts[1]
		}

		specType := swagtool.ToOpenApiType(fieldInterface)
		switch ruleName {
		case "email":
			if specType == "string" {
				schema.Value.Format = "email"
			} else {
				logger.Warn("Validation rule 'email' is only applicable to string fields, got %s", specType)
			}
		case "uuid":
			if specType == "string" {
				schema.Value.Format = "uuid"
			} else {
				logger.Warn("Validation rule 'uuid' is only applicable to string fields, got %s", specType)
			}
		case "ip":
			if specType == "string" {
				schema.Value.Format = "ipv4"
			} else {
				logger.Warn("Validation rule 'ip' is only applicable to string fields, got %s", specType)
			}
		case "ipv4":
			if specType == "string" {
				schema.Value.Format = "ipv4"
			} else {
				logger.Warn("Validation rule 'ipv4' is only applicable to string fields, got %s", specType)
			}
		case "ipv6":
			if specType == "string" {
				schema.Value.Format = "ipv6"
			} else {
				logger.Warn("Validation rule 'ipv6' is only applicable to string fields, got %s", specType)
			}
		case "hostname":
			if specType == "string" {
				schema.Value.Format = "hostname"
			} else {
				logger.Warn("Validation rule 'hostname' is only applicable to string fields, got %s", specType)
			}
		case "date":
			if specType == "string" {
				schema.Value.Format = "date"
			} else {
				logger.Warn("Validation rule 'date' is only applicable to string fields, got %s", specType)
			}
		case "datetime":
			if specType == "string" {
				schema.Value.Format = "date-time"
			} else {
				logger.Warn("Validation rule 'datetime' is only applicable to string fields, got %s", specType)
			}
		case "gt":
			if specType == "integer" || specType == "number" {
				schema.Value.Min = swagtool.ParseNumber(ruleValue)
				schema.Value.ExclusiveMin = true
			} else {
				logger.Warn("Validation rule 'gt' is only applicable to numeric fields, got %s", specType)
			}
		case "gte":
			if specType == "integer" || specType == "number" {
				schema.Value.Min = swagtool.ParseNumber(ruleValue)
				schema.Value.ExclusiveMin = false
			} else {
				logger.Warn("Validation rule 'gte' is only applicable to numeric fields, got %s", specType)
			}
		case "lt":
			if specType == "integer" || specType == "number" {
				schema.Value.Max = swagtool.ParseNumber(ruleValue)
				schema.Value.ExclusiveMax = true
			} else {
				logger.Warn("Validation rule 'lt' is only applicable to numeric fields, got %s", specType)
			}
		case "lte":
			if specType == "integer" || specType == "number" {
				schema.Value.Max = swagtool.ParseNumber(ruleValue)
				schema.Value.ExclusiveMax = false
			} else {
				logger.Warn("Validation rule 'lte' is only applicable to numeric fields, got %s", specType)
			}
		case "min":
			if specType == "string" {
				schema.Value.MinLength = *swagtool.ParseUInteger(ruleValue)
			} else if specType == "integer" || specType == "number" {
				schema.Value.Min = swagtool.ParseNumber(ruleValue)
				schema.Value.ExclusiveMin = false
			} else {
				logger.Warn("Validation rule 'min' is only applicable to string or numeric fields, got %s", specType)
			}
		case "max":
			if specType == "string" {
				schema.Value.MaxLength = swagtool.ParseUInteger(ruleValue)
			} else if specType == "integer" || specType == "number" {
				schema.Value.Max = swagtool.ParseNumber(ruleValue)
				schema.Value.ExclusiveMax = false
			} else {
				logger.Warn("Validation rule 'max' is only applicable to string or numeric fields, got %s", specType)
			}
		case "len":
			if specType == "string" {
				length := swagtool.ParseUInteger(ruleValue)
				schema.Value.MinLength = *length
				schema.Value.MaxLength = length
			} else {
				logger.Warn("Validation rule 'len' is only applicable to string fields, got %s", specType)
			}
		case "pattern":
			if specType == "string" {
				schema.Value.Pattern = ruleValue
			} else {
				logger.Warn("Validation rule 'pattern' is only applicable to string fields, got %s", specType)
			}
		case "minItems":
			if specType == "array" {
				schema.Value.MinItems = *swagtool.ParseUInteger(ruleValue)
			} else {
				logger.Warn("Validation rule 'minItems' is only applicable to array fields, got %s", specType)
			}
		case "maxItems":
			if specType == "array" {
				schema.Value.MaxItems = swagtool.ParseUInteger(ruleValue)
			} else {
				logger.Warn("Validation rule 'maxItems' is only applicable to array fields, got %s", specType)
			}
		case "uniqueItems":
			if specType == "array" {
				schema.Value.UniqueItems = *swagtool.ParseBool(ruleValue)
			} else {
				logger.Warn("Validation rule 'uniqueItems' is only applicable to array fields, got %s", specType)
			}
		case "enum":
			enumValues := strings.Split(ruleValue, "|")
			if len(enumValues) == 0 || enumValues[0] == "" {
				logger.Warn("Validation rule 'enum' must have at least one value")
				schema.Value.Enum = nil
			} else {
				for _, v := range enumValues {
					schema.Value.Enum = append(schema.Value.Enum, v)
				}
			}
		case "oneof":
			oneofValues := strings.Fields(ruleValue)
			if len(oneofValues) == 0 {
				logger.Warn("Validation rule 'oneof' must have at least one value")
				continue
			}

			schema.Value.Enum = make([]interface{}, 0, len(oneofValues))

			switch specType {
			case "string":
				for _, v := range oneofValues {
					schema.Value.Enum = append(schema.Value.Enum, v)
				}
			case "integer":
				for _, v := range oneofValues {
					if val, err := strconv.ParseInt(v, 10, 64); err == nil {
						schema.Value.Enum = append(schema.Value.Enum, val)
					} else {
						logger.Warn("Invalid integer value in oneof: %s", v)
					}
				}
			case "number":
				for _, v := range oneofValues {
					if val, err := strconv.ParseFloat(v, 64); err == nil {
						schema.Value.Enum = append(schema.Value.Enum, val)
					} else {
						logger.Warn("Invalid number value in oneof: %s", v)
					}
				}
			default:
				logger.Warn("oneof validation for type %s might not be properly handled", specType)
				for _, v := range oneofValues {
					schema.Value.Enum = append(schema.Value.Enum, v)
				}
			}
		}
	}
}
