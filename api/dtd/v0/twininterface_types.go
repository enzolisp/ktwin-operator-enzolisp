/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v0

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TwinInterfacePhase string

const (
	TwinInterfacePhasePending TwinInterfacePhase = "Pending"
	TwinInterfacePhaseUnknown TwinInterfacePhase = "Unknown"
	TwinInterfacePhaseRunning TwinInterfacePhase = "Running"
	TwinInterfacePhaseFailed  TwinInterfacePhase = "Failed"
)

type PrimitiveType string
type Multiplicity string

const (
	Integer PrimitiveType = "integer"
	String  PrimitiveType = "string"
	Boolean PrimitiveType = "boolean"
	Double  PrimitiveType = "double"
)

const (
	ONE  Multiplicity = "one"
	MANY Multiplicity = "many"
)

// TwinInterfaceSpec defines the desired state of TwinInterface
type TwinInterfaceSpec struct {
	Id               string                 `json:"id,omitempty"`
	DisplayName      string                 `json:"displayName,omitempty"`
	Description      string                 `json:"description,omitempty"`
	Comment          string                 `json:"comment,omitempty"`
	Properties       []TwinProperty         `json:"properties,omitempty"`
	Commands         []TwinCommand          `json:"commands,omitempty"`
	Relationships    []TwinRelationship     `json:"relationships,omitempty"`
	Telemetries      []TwinTelemetry        `json:"telemetries,omitempty"`
	Template         corev1.PodTemplateSpec `json:"template,omitempty"`
	ExtendsInterface string                 `json:"extendsInterface,omitempty"`
}

type TwinProperty struct {
	Id          string      `json:"id,omitempty"`
	Comment     string      `json:"comment,omitempty"`
	Description string      `json:"description,omitempty"`
	DisplayName string      `json:"displayName,omitempty"`
	Name        string      `json:"name,omitempty"`
	Schema      *TwinSchema `json:"schema,omitempty"`
	Writeable   bool        `json:"writable,omitempty"`
}

type TwinCommand struct {
	Id          string          `json:"id,omitempty"`
	Comment     string          `json:"comment,omitempty"`
	Description string          `json:"description,omitempty"`
	DisplayName string          `json:"displayName,omitempty"`
	Name        string          `json:"name,omitempty"`
	CommandType string          `json:"commandType,omitempty"` // async, sync
	Request     CommandRequest  `json:"request"`
	Response    CommandResponse `json:"response"`
}

type CommandRequest struct {
	Name        string      `json:"name,omitempty"`
	DisplayName string      `json:"displayName,omitempty"`
	Description string      `json:"description,omitempty"`
	Schema      *TwinSchema `json:"schema,omitempty"`
}

type CommandResponse struct {
	Name        string      `json:"name,omitempty"`
	DisplayName string      `json:"displayName,omitempty"`
	Description string      `json:"description,omitempty"`
	Schema      *TwinSchema `json:"schema,omitempty"`
}

type TwinRelationship struct {
	Id              string         `json:"id,omitempty"`
	Comment         string         `json:"comment,omitempty"`
	Description     string         `json:"description,omitempty"`
	DisplayName     string         `json:"displayName,omitempty"`
	MaxMultiplicity int            `json:"maxMultiplicity,omitempty"`
	MinMultiplicity int            `json:"minMultiplicity,omitempty"`
	Name            string         `json:"name,omitempty"`
	Properties      []TwinProperty `json:"properties,omitempty"`
	Target          string         `json:"target,omitempty"`
	Schema          *TwinSchema    `json:"schema,omitempty"`
	Writeable       bool           `json:"writeable,omitempty"`
}

type TwinTelemetry struct {
	Id          string      `json:"id,omitempty"`
	Comment     string      `json:"comment,omitempty"`
	Description string      `json:"description,omitempty"`
	DisplayName string      `json:"displayName,omitempty"`
	Name        string      `json:"name,omitempty"`
	Schema      *TwinSchema `json:"schema,omitempty"`
}

type TwinSchema struct {
	PrimitiveType PrimitiveType   `json:"primitiveType,omitempty"`
	EnumType      *TwinEnumSchema `json:"enumType,omitempty"`
}

type TwinEnumSchema struct {
	ValueSchema PrimitiveType          `json:"valueSchema,omitempty"`
	EnumValues  []TwinEnumSchemaValues `json:"enumValues,omitempty"`
}

type TwinEnumSchemaValues struct {
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	EnumValue   string `json:"enumValue,omitempty"`
}

// TwinInterfaceStatus defines the observed state of TwinInterface
type TwinInterfaceStatus struct {
	Status TwinInterfacePhase `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TwinInterface is the Schema for the twininterfaces API
type TwinInterface struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TwinInterfaceSpec   `json:"spec,omitempty"`
	Status TwinInterfaceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TwinInterfaceList contains a list of TwinInterface
type TwinInterfaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TwinInterface `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TwinInterface{}, &TwinInterfaceList{})
}
