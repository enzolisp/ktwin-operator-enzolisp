package dtdl

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
)

var (
	ContentCommandType      = "Command"
	ContentComponentType    = "Component"
	ContentPropertyType     = "Property"
	ContentRelationshipType = "Relationship"
	ContentTelemetryType    = "Telemetry"

	ErrContentUnmarshalInvalidType = errors.New("Invalid content @type")
)

func ErrContentUnmarshalTypeNotSupported(typeValue string) error {
	return errors.New(fmt.Sprintf("Content @type %s not supported", typeValue))
}

type Content struct {
	Command      *Command
	Component    *Component
	Property     *Property
	Relationship *Relationship
	Telemetry    *Telemetry
}

func (c *Content) UnmarshalJSON(data []byte) error {
	var jsonObject interface{}
	err := json.Unmarshal(data, &jsonObject)

	if err != nil {
		return err
	}

	objectMap := jsonObject.(map[string]interface{})

	objectType := objectMap["@type"].(string)

	switch objectMap["@type"] {
	case ContentPropertyType:
		c.Property = c.newProperty(objectMap)
		return nil
	case ContentRelationshipType:
		c.Relationship = c.newRelationship(objectMap)
		return nil
	case ContentCommandType:
		c.Command = c.newCommand(objectMap)
		return nil
	case ContentComponentType:
		return ErrContentUnmarshalTypeNotSupported(objectType)
	case ContentTelemetryType:
		return ErrContentUnmarshalTypeNotSupported(objectType)
	default:
		return ErrContentUnmarshalInvalidType
	}
}

func (c Content) MarshalYAML() (interface{}, error) {
	if !reflect.DeepEqual(c.Command, Command{}) {
		return c.Command, nil
	}

	if !reflect.DeepEqual(c.Relationship, Relationship{}) {
		return c.Relationship, nil
	}

	if !reflect.DeepEqual(c.Property, Property{}) {
		return c.Property, nil
	}

	if !reflect.DeepEqual(c.Telemetry, Telemetry{}) {
		return c.Telemetry, nil
	}

	if !reflect.DeepEqual(c.Component, Component{}) {
		return c.Component, nil
	}

	return nil, errors.New("Not possible to marshal Yaml")
}

func (s *Content) newProperty(data interface{}) *Property {
	property := Property{}

	dataByte, err := json.Marshal(data)

	if err != nil {
		log.Fatal("Error in marshaling")
	}

	err = json.Unmarshal(dataByte, &property)

	if err != nil {
		log.Fatal("Error in unmarshaling")
	}

	return &property
}

func (s *Content) newRelationship(data interface{}) *Relationship {
	relationship := Relationship{}

	dataByte, err := json.Marshal(data)

	if err != nil {
		log.Fatal("Error in marshaling")
	}

	err = json.Unmarshal(dataByte, &relationship)

	if err != nil {
		log.Fatal("Error in unmarshaling")
	}

	return &relationship
}

func (s *Content) newCommand(data interface{}) *Command {
	command := Command{}

	dataByte, err := json.Marshal(data)

	if err != nil {
		log.Fatal("Error in marshaling")
	}

	err = json.Unmarshal(dataByte, &command)

	if err != nil {
		log.Fatal("Error in unmarshaling")
	}

	return &command
}
