package extractor

import (
	"regexp"
)

type Attribute struct {
	Name       string
	Value      string
	Properties map[string]string
}

type AttributesHolder struct {
	attributes []Attribute
}

func NewAttributesHolder(comments []string) AttributesHolder {
	holder := AttributesHolder{}

	generalPattern := regexp.MustCompile(`//\s*@(\w+)(?:\s*\(([^)]*)\))?(?:\s+(.+))?`)

	for _, text := range comments {
		matches := generalPattern.FindStringSubmatch(text)

		attrib := Attribute{}
		if len(matches) <= 0 {
			continue
		}

		if len(matches) > 0 {
			attrib.Name = matches[1]
		}

		if matches[2] != "" {
			// Case: Parentheses with key-value pairs

			// Parse key-value pairs
			kvPattern := `(\w+):\s*([^;]+)`
			kvRe := regexp.MustCompile(kvPattern)
			params := kvRe.FindAllStringSubmatch(matches[2], -1)

			// Map for key-value pairs
			attrib.Properties = make(map[string]string)
			for _, keyValueTuple := range params {
				attrib.Properties[keyValueTuple[1]] = keyValueTuple[2]
			}
		}

		// Case: Trailing content after parentheses
		if matches[3] != "" {
			attrib.Value = matches[3]
		}

		holder.attributes = append(holder.attributes, attrib)
	}

	return holder
}

func (holder AttributesHolder) Get(attribute string) *Attribute {
	for _, attrib := range holder.attributes {
		if attrib.Name == attribute {
			return &attrib
		}
	}

	return nil
}

func (holder AttributesHolder) Has(attribute string) bool {
	return holder.Get(attribute) != nil
}
