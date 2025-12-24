package swagen31

import (
	"log"
	"strconv"
	"strings"

	"github.com/gopher-fleece/gleece/v2/generator/swagen/swagtool"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"go.yaml.in/yaml/v4"
)

func BuildSchemaValidationV31(schema *base.Schema, validationString string, fieldInterface string) {
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
		case "email", "uuid", "ip", "ipv4", "ipv6", "hostname", "date", "datetime":
			if specType == "string" {
				formatValue := ruleName
				switch ruleName {
				case "ip":
					formatValue = "ipv4"
				case "datetime":
					formatValue = "date-time"
				}
				schema.Format = formatValue
			} else {
				log.Printf("Validation rule '%s' is only applicable to string fields, got %s", ruleName, specType)
			}
		case "gt":
			if specType == "integer" || specType == "number" {
				val := swagtool.ParseNumber(ruleValue)
				schema.ExclusiveMinimum = &base.DynamicValue[bool, float64]{
					B: *val,
					N: 1,
				}
			} else {
				log.Printf("Validation rule '%s' is only applicable to numeric fields, got %s", ruleName, specType)
			}
		case "gte":
			if specType == "integer" || specType == "number" {
				val := swagtool.ParseNumber(ruleValue)
				schema.Minimum = val
			} else {
				log.Printf("Validation rule '%s' is only applicable to numeric fields, got %s", ruleName, specType)
			}
		case "lt":
			if specType == "integer" || specType == "number" {
				val := swagtool.ParseNumber(ruleValue)
				schema.ExclusiveMaximum = &base.DynamicValue[bool, float64]{
					B: *val,
					N: 1,
				}
			} else {
				log.Printf("Validation rule '%s' is only applicable to numeric fields, got %s", ruleName, specType)
			}
		case "lte":
			if specType == "integer" || specType == "number" {
				val := swagtool.ParseNumber(ruleValue)
				schema.Maximum = val
			} else {
				log.Printf("Validation rule '%s' is only applicable to numeric fields, got %s", ruleName, specType)
			}
		case "min":
			if specType == "string" {
				val := swagtool.ParseInteger(ruleValue)
				schema.MinLength = val
			} else if specType == "integer" || specType == "number" {
				schema.Minimum = swagtool.ParseNumber(ruleValue)
			} else {
				log.Printf("Validation rule 'min' is only applicable to string or numeric fields, got %s", specType)
			}
		case "max":
			if specType == "string" {
				val := swagtool.ParseInteger(ruleValue)
				schema.MaxLength = val
			} else if specType == "integer" || specType == "number" {
				schema.Maximum = swagtool.ParseNumber(ruleValue)
			} else {
				log.Printf("Validation rule 'max' is only applicable to string or numeric fields, got %s", specType)
			}
		case "len":
			if specType == "string" {
				length := swagtool.ParseInteger(ruleValue)
				schema.MinLength = length
				schema.MaxLength = length
			} else {
				log.Printf("Validation rule 'len' is only applicable to string fields, got %s", specType)
			}
		case "pattern":
			if specType == "string" {
				schema.Pattern = ruleValue
			} else {
				log.Printf("Validation rule 'pattern' is only applicable to string fields, got %s", specType)
			}
		case "minItems":
			if specType == "array" {
				val := swagtool.ParseInteger(ruleValue)
				schema.MinItems = val
			} else {
				log.Printf("Validation rule 'minItems' is only applicable to array fields, got %s", specType)
			}
		case "maxItems":
			if specType == "array" {
				val := swagtool.ParseInteger(ruleValue)
				schema.MaxItems = val
			} else {
				log.Printf("Validation rule 'maxItems' is only applicable to array fields, got %s", specType)
			}
		case "uniqueItems":
			if specType == "array" {
				val := swagtool.ParseBool(ruleValue)
				schema.UniqueItems = val
			} else {
				log.Printf("Validation rule 'uniqueItems' is only applicable to array fields, got %s", specType)
			}
		case "enum":
			enumValues := strings.Split(ruleValue, "|")
			if len(enumValues) == 0 || enumValues[0] == "" {
				log.Printf("Validation rule 'enum' must have at least one value")
				schema.Enum = nil
			} else {
				schema.Enum = make([]*yaml.Node, 0, len(enumValues))
				for _, v := range enumValues {
					node := &yaml.Node{
						Kind:  yaml.ScalarNode,
						Value: v,
					}
					schema.Enum = append(schema.Enum, node)
				}
			}
		case "oneof":
			oneofValues := strings.Fields(ruleValue)
			if len(oneofValues) == 0 {
				log.Printf("Validation rule 'oneof' must have at least one value")
				continue
			}

			schema.Enum = make([]*yaml.Node, 0, len(oneofValues))

			switch specType {
			case "string":
				for _, v := range oneofValues {
					node := &yaml.Node{
						Kind:  yaml.ScalarNode,
						Value: v,
					}
					schema.Enum = append(schema.Enum, node)
				}
			case "integer":
				for _, v := range oneofValues {
					if _, err := strconv.ParseInt(v, 10, 64); err == nil {
						node := &yaml.Node{
							Kind:  yaml.ScalarNode,
							Value: v,
							Tag:   "!!int",
						}
						schema.Enum = append(schema.Enum, node)
					} else {
						log.Printf("Invalid integer value in oneof: %s", v)
					}
				}
			case "number":
				for _, v := range oneofValues {
					if _, err := strconv.ParseFloat(v, 64); err == nil {
						node := &yaml.Node{
							Kind:  yaml.ScalarNode,
							Value: v,
							Tag:   "!!float",
						}
						schema.Enum = append(schema.Enum, node)
					} else {
						log.Printf("Invalid number value in oneof: %s", v)
					}
				}
			default:
				log.Printf("oneof validation for type %s might not be properly handled", specType)
				for _, v := range oneofValues {
					node := &yaml.Node{
						Kind:  yaml.ScalarNode,
						Value: v,
					}
					schema.Enum = append(schema.Enum, node)
				}
			}
		}
	}
}
