package v0

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MQTTTriggerSpec defines the desired state of MQTTTrigger
type MQTTTriggerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of MQTTTrigger. Edit mqtttrigger_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// MQTTTriggerStatus defines the observed state of MQTTTrigger
type MQTTTriggerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MQTTTrigger is the Schema for the mqtttriggers API
type MQTTTrigger struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MQTTTriggerSpec   `json:"spec,omitempty"`
	Status MQTTTriggerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MQTTTriggerList contains a list of MQTTTrigger
type MQTTTriggerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MQTTTrigger `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MQTTTrigger{}, &MQTTTriggerList{})
}
