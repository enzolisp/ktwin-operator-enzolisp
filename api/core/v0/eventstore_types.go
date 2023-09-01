package v0

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EventStoreSpec defines the desired state of EventStore
type EventStoreSpec struct {
}

// EventStoreStatus defines the observed state of EventStore
type EventStoreStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// EventStore is the Schema for the eventstores API
type EventStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EventStoreSpec   `json:"spec,omitempty"`
	Status EventStoreStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EventStoreList contains a list of EventStore
type EventStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EventStore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EventStore{}, &EventStoreList{})
}
