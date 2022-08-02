/*
Copyright 2022 ysicing(i@ysicing.me).
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InitNasSpec defines the desired state of InitNas
type InitNasSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of InitNas. Edit initnas_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// InitNasStatus defines the observed state of InitNas
type InitNasStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// InitNas is the Schema for the initnas API
type InitNas struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InitNasSpec   `json:"spec,omitempty"`
	Status InitNasStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InitNasList contains a list of InitNas
type InitNasList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InitNas `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InitNas{}, &InitNasList{})
}
