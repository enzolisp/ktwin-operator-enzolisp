package v0

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AutoScalerType string

const (
	CONCURRENCY AutoScalerType = "concurrency"
	RPS         AutoScalerType = "rps"
	CPU         AutoScalerType = "cpu"
	MEMORY      AutoScalerType = "memory"
)

// EventStoreSpec defines the desired state of EventStore
type EventStoreSpec struct {
	AutoScaling EventStoreAutoScaling `json:"autoScaling,omitempty"`
}

type EventStoreAutoScaling struct {
	MinScale                    *int `json:"minScale,omitempty"`
	MaxScale                    *int `json:"maxScale,omitempty"`
	Target                      *int `json:"target,omitempty"`
	TargetUtilizationPercentage *int `json:"targetUtilizationPercentage,omitempty"`
	Parallelism                 *int `json:"parallelism,omitempty"`
	// KNative Metric values (default, if not informed: concurrency)
	// concurrency: the number of simultaneous requests that can be processed by each replica of an application at any given time
	// rps: requests per seconds
	// cpu: cpu usage
	// memory: memory usage
	Metric AutoScalerType `json:"metric,omitempty"`
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
