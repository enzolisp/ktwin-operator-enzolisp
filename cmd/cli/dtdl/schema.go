package dtdl

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

var (
	STANDARD_DOUBLE_SCHEMA  = "double"
	STANDARD_FLOAT_SCHEMA   = "float"
	STANDARD_BOOLEAN_SCHEMA = "boolean"
	STANDARD_STRING_SCHEMA  = "string"
	STANDARD_ENUM_SCHEMA    = "enum"

	ErrUnmarshalTypeNotSupported = errors.New("Unmarshal type not supported")
	ErrInvalidSchemaType         = errors.New("Invalid schema type")
	ErrUnmarshalUnknown          = errors.New("Unmarshal unknown error")
)

// https://github.com/Azure/opendigitaltwins-dtdl/blob/master/DTDL/v3/DTDL.v3.md#schema

// Primitive Schemas:
// boolean - a boolean value
// date - a date in ISO 8601 format, per RFC 3339
// dateTime - a date and time in ISO 8601 format, per RFC 3339
// double - a finite numeric value that is expressible in IEEE 754 double-precision floating point format, conformant with xsd:double
// duration - a duration in ISO 8601 format
// float - a finite numeric value that is expressible in IEEE 754 single-precision floating point format, conformant with xsd:float
// integer - a signed integral numeric value that is expressible in 4 bytes
// long	- a signed integral numeric value that is expressible in 8 bytes
// string - a UTF8 string
// time	- a time in ISO 8601 format, per RFC 3339

// Complex Schemas:
// Array - Not supported
// Enum - Supported
// Map - Not supported
// Object - Not supported

type Schema struct {
	IsDefaultSchema    bool
	DefaultSchemaValue string
	EnumSchema         EnumSchema
}

type EnumSchema struct {
	Type        string             `json:"@type" yaml:"type,omitempty"`
	ValueSchema string             `json:"valueSchema" yaml:"valueSchema,omitempty"`
	EnumValues  []EnumSchemaValues `json:"enumValues" yaml:"enumValues,omitempty"`
}

type EnumSchemaValues struct {
	Name        string `json:"name" yaml:"name,omitempty"`
	DisplayName string `json:"displayName" yaml:"displayName,omitempty"`
	EnumValue   string `json:"enumValue" yaml:"enumValue,omitempty"`
}

func (s *Schema) UnmarshalJSON(data []byte) error {
	var jsonObject interface{}
	err := json.Unmarshal(data, &jsonObject)

	if err != nil {
		return err
	}

	switch object := jsonObject.(type) {
	case string:
		// TODO: check if the type is valid
		*s = Schema{
			IsDefaultSchema:    true,
			DefaultSchemaValue: object,
		}
		return nil
	case nil:
		return nil
	case interface{}:
		enumSchema, err := s.processSchemaInterface(jsonObject)

		if err != nil {
			return err
		}

		*s = Schema{
			IsDefaultSchema: false,
			EnumSchema:      enumSchema,
		}

		return nil
	}

	return ErrUnmarshalUnknown
}

func (s *Schema) MarshalJSON() ([]byte, error) {
	if s.IsDefaultSchema {
		return json.Marshal(s.DefaultSchemaValue)
	}

	return json.Marshal(s.EnumSchema)
}

func (s Schema) MarshalYAML() (interface{}, error) {
	if s.IsDefaultSchema {
		fmt.Println("Default schema")
		return s.DefaultSchemaValue, nil
	}
	fmt.Println("Enum schema")
	return s.EnumSchema, nil
}

// Schema
func (s *Schema) processSchemaInterface(jsonObject interface{}) (EnumSchema, error) {

	objectMap, isMapStringInterface := jsonObject.(map[string]interface{})

	if isMapStringInterface {

		schemaType := objectMap["@type"].(string)
		if s.isValidSchemaType(schemaType) {

			valueSchema := objectMap["valueSchema"].(string)
			enumValues := s.processSchemaEnumValues(objectMap["enumValues"])

			enumSchema := EnumSchema{
				Type:        schemaType,
				ValueSchema: valueSchema,
				EnumValues:  enumValues,
			}

			return enumSchema, nil
		}

		log.Fatal("It was not able to process schema. Schema type is invalid.")
		return EnumSchema{}, ErrInvalidSchemaType
	}

	return EnumSchema{}, ErrUnmarshalTypeNotSupported
}

func (s *Schema) processSchemaEnumValues(enumValuesMap interface{}) []EnumSchemaValues {

	enumValues, isValidListMap := enumValuesMap.([]interface{})

	if !isValidListMap {
		log.Fatal("Invalid type for enum values array")
	}

	var enumSchemaValues []EnumSchemaValues = make([]EnumSchemaValues, 0)

	for _, enumValue := range enumValues {
		enumMap := enumValue.(map[string]interface{})
		enumSchemaValue := EnumSchemaValues{
			Name:        enumMap["name"].(string),
			DisplayName: enumMap["displayName"].(string),
			EnumValue:   enumMap["enumValue"].(string),
		}
		enumSchemaValues = append(enumSchemaValues, enumSchemaValue)
	}

	return enumSchemaValues

}

func (s *Schema) isValidSchemaType(schemaType string) bool {
	return schemaType == "Enum"
}
