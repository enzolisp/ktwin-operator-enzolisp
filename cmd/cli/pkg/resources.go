package pkg

import (
	"reflect"

	apiv0 "github.com/Open-Digital-Twin/ktwin-operator/api/dtd/v0"
	dtdl "github.com/Open-Digital-Twin/ktwin-operator/cmd/cli/dtdl"
	"github.com/Open-Digital-Twin/ktwin-operator/pkg/naming"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Open-Digital-Twin/ktwin-operator/cmd/cli/utils"
)

const (
	INSTANCE_SUFFIX = "-001"
)

type ResourceBuilder interface {
	CreateTwinInterface(tInterface dtdl.Interface) apiv0.TwinInterface
	CreateTwinInstance(twinInterface apiv0.TwinInterface, parentTwinInterfaces []apiv0.TwinInterface) apiv0.TwinInstance
}

func NewResourceBuilder() ResourceBuilder {
	return &resourceBuilder{
		hostUtils: utils.NewHostUtils(),
	}
}

type resourceBuilder struct {
	hostUtils utils.HostUtils
}

// TODO: renew TwinInterface to TwinInstance
func (r *resourceBuilder) CreateTwinInterface(tInterface dtdl.Interface) apiv0.TwinInterface {
	var properties []apiv0.TwinProperty
	var relationships []apiv0.TwinRelationship
	var telemetries []apiv0.TwinTelemetry
	var commands []apiv0.TwinCommand
	var interfaceExtends string

	for _, content := range tInterface.Contents {
		if content.Property != nil {
			properties = r.processProperty(*content.Property, properties)
		}
		if content.Relationship != nil {
			relationships = r.processRelationship(*content.Relationship, relationships)
		}
		if content.Telemetry != nil {
			telemetries = r.processTelemetry(*content.Telemetry, telemetries)
		}
		if content.Command != nil {
			commands = r.processCommand(*content.Command, commands)
		}
	}

	// Only supports one parent interface
	if len(tInterface.Extends) > 0 {
		interfaceExtends = r.hostUtils.ParseHostName(tInterface.Extends[0])
	}

	normalizedInterfaceId := r.hostUtils.ParseHostName(string(tInterface.Id))

	twinInterface := apiv0.TwinInterface{
		TypeMeta: v1.TypeMeta{
			Kind:       "TwinInterface",
			APIVersion: "dtd.ktwin/v0",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      normalizedInterfaceId,
			Namespace: "ktwin",
		},
		Spec: apiv0.TwinInterfaceSpec{
			Id:               normalizedInterfaceId,
			DisplayName:      string(tInterface.DisplayName),
			Description:      string(tInterface.Description),
			Comment:          string(tInterface.Comment),
			Properties:       properties,
			Relationships:    relationships,
			Commands:         commands,
			Telemetries:      telemetries,
			ExtendsInterface: interfaceExtends,
			Service: &apiv0.TwinInterfaceService{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{Containers: []corev1.Container{
						{
							Name:            normalizedInterfaceId,
							Image:           naming.GetContainerRegistry("ktwin-edge-service:0.1"),
							ImagePullPolicy: corev1.PullIfNotPresent,
						},
					}},
				},
			},
		},
	}

	return twinInterface
}

func (r *resourceBuilder) CreateTwinInstance(twinInterface apiv0.TwinInterface, parentTwinInterfaces []apiv0.TwinInterface) apiv0.TwinInstance {
	normalizeTwinInterfacedId := r.hostUtils.ParseHostName(string(twinInterface.Spec.Id))
	normalizeTwinInstanceId := normalizeTwinInterfacedId + INSTANCE_SUFFIX

	twinInstance := apiv0.TwinInstance{
		TypeMeta: v1.TypeMeta{
			Kind:       "TwinInstance",
			APIVersion: "dtd.ktwin/v0",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      normalizeTwinInstanceId,
			Namespace: "ktwin",
		},
		Spec: apiv0.TwinInstanceSpec{
			Interface:                 normalizeTwinInterfacedId,
			TwinInstanceRelationships: r.getTwinInstanceRelationships(parentTwinInterfaces),
			Data:                      r.getTwinData(parentTwinInterfaces),
		},
	}

	return twinInstance
}

func (r *resourceBuilder) getTwinInstanceRelationships(parentTwinInterfaces []apiv0.TwinInterface) []apiv0.TwinInstanceRelationship {
	var twinInstanceRelationship []apiv0.TwinInstanceRelationship

	for _, twinInterface := range parentTwinInterfaces {
		if len(twinInterface.Spec.Relationships) == 0 {
			continue
		}

		for _, twinRelationship := range twinInterface.Spec.Relationships {
			twinInstanceRelationship = append(twinInstanceRelationship, apiv0.TwinInstanceRelationship{
				Name:      twinRelationship.Name + INSTANCE_SUFFIX,
				Interface: twinRelationship.Interface,
				Instance:  twinRelationship.Interface + INSTANCE_SUFFIX,
			})
		}
	}

	return twinInstanceRelationship
}

func (r *resourceBuilder) getTwinData(twinInterfaces []apiv0.TwinInterface) *apiv0.TwinInstanceDataSpec {
	var twinInstanceData *apiv0.TwinInstanceDataSpec

	for _, twinInterface := range twinInterfaces {
		if len(twinInterface.Spec.Properties) == 0 && len(twinInterface.Spec.Telemetries) == 0 {
			continue
		}

		if twinInstanceData == nil {
			twinInstanceData = &apiv0.TwinInstanceDataSpec{}
		}

		for _, twinProperty := range twinInterface.Spec.Properties {
			twinInstanceData.Properties = append(twinInstanceData.Properties, apiv0.TwinInstancePropertyData{
				Id:   twinProperty.Id,
				Name: twinProperty.Name,
			})
		}

		for _, twinTelemetry := range twinInterface.Spec.Telemetries {
			twinInstanceData.Telemetries = append(twinInstanceData.Telemetries, apiv0.TwinInstanceTelemetryData{
				Id:   twinTelemetry.Id,
				Name: twinTelemetry.Name,
			})
		}
	}

	return twinInstanceData
}

func (r *resourceBuilder) processCommand(command dtdl.Command, commands []apiv0.TwinCommand) []apiv0.TwinCommand {
	newCommand := apiv0.TwinCommand{
		Id:          string(command.Id),
		Comment:     command.Comment,
		Description: string(command.Description),
		DisplayName: string(command.DisplayName),
		Name:        command.Name,
		CommandType: command.CommandType,
		Request: apiv0.CommandRequest{
			Name:        command.Request.Name,
			DisplayName: string(command.Request.DisplayName),
			Description: string(command.Request.Comment),
		},
		Response: apiv0.CommandResponse{
			Name:        command.Response.Name,
			DisplayName: string(command.Response.DisplayName),
			Description: string(command.Response.Comment),
			//Schema:      command.Response.Schema,
		},
	}
	commands = append(commands, newCommand)
	return commands
}

func (r *resourceBuilder) processTelemetry(telemetry dtdl.Telemetry, telemetries []apiv0.TwinTelemetry) []apiv0.TwinTelemetry {
	twinSchema := r.createTwinSchema(telemetry.Schema)
	newTelemetry := apiv0.TwinTelemetry{
		Id:          string(telemetry.Id),
		Comment:     telemetry.Comment,
		Description: string(telemetry.Description),
		DisplayName: string(telemetry.DisplayName),
		Name:        telemetry.Name,
		Schema:      twinSchema,
	}
	telemetries = append(telemetries, newTelemetry)
	return telemetries
}

func (r *resourceBuilder) processRelationship(relationship dtdl.Relationship, relationships []apiv0.TwinRelationship) []apiv0.TwinRelationship {

	var relationshipProperties []apiv0.TwinProperty

	if relationship.Properties != nil {
		for _, property := range relationship.Properties {
			relationshipProperties = r.processProperty(property, relationshipProperties)
		}
	}

	twinSchema := r.createTwinSchema(relationship.Schema)
	newRelationship := apiv0.TwinRelationship{
		Id:              string(relationship.Id),
		Comment:         relationship.Comment,
		Description:     string(relationship.Description),
		DisplayName:     string(relationship.DisplayName),
		Name:            relationship.Name,
		Writeable:       relationship.Writeable,
		MaxMultiplicity: relationship.MaxMultiplicity,
		MinMultiplicity: relationship.MinMultiplicity,
		Schema:          twinSchema,
		Properties:      relationshipProperties,
		Interface:       r.hostUtils.ParseHostName(string(relationship.Target)),
	}
	relationships = append(relationships, newRelationship)
	return relationships
}

func (r *resourceBuilder) processProperty(property dtdl.Property, properties []apiv0.TwinProperty) []apiv0.TwinProperty {
	twinSchema := r.createTwinSchema(property.Schema)
	newProperty := apiv0.TwinProperty{
		Id:          string(property.Id),
		Comment:     property.Comment,
		Description: string(property.Description),
		DisplayName: string(property.DisplayName),
		Name:        property.Name,
		Writeable:   property.Writeable,
		Schema:      twinSchema,
	}
	properties = append(properties, newProperty)
	return properties
}

func (r *resourceBuilder) createTwinSchema(schema dtdl.Schema) *apiv0.TwinSchema {

	if reflect.DeepEqual(schema, dtdl.Schema{}) {
		return nil
	}

	var twinEnumSchemaValues []apiv0.TwinEnumSchemaValues
	var twinEnumSchema *apiv0.TwinEnumSchema

	for _, enumValue := range schema.EnumSchema.EnumValues {
		twinEnumValue := apiv0.TwinEnumSchemaValues{
			Name:        enumValue.Name,
			DisplayName: enumValue.DisplayName,
			EnumValue:   enumValue.EnumValue,
		}
		twinEnumSchemaValues = append(twinEnumSchemaValues, twinEnumValue)
	}

	if len(twinEnumSchemaValues) > 1 || schema.EnumSchema.ValueSchema != "" {
		twinEnumSchema = &apiv0.TwinEnumSchema{
			ValueSchema: apiv0.PrimitiveType(schema.EnumSchema.ValueSchema),
			EnumValues:  twinEnumSchemaValues,
		}
	}

	twinSchema := &apiv0.TwinSchema{
		PrimitiveType: apiv0.PrimitiveType(schema.DefaultSchemaValue),
		EnumType:      twinEnumSchema,
	}

	return twinSchema
}
