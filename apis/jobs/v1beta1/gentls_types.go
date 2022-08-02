/*
Copyright 2022 ysicing(i@ysicing.me).
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GenTLSSpec defines the desired state of GenTLS
type GenTLSSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of GenTLS. Edit gentls_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// GenTLSStatus defines the observed state of GenTLS
type GenTLSStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GenTLS is the Schema for the gentls API
type GenTLS struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GenTLSSpec   `json:"spec,omitempty"`
	Status GenTLSStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GenTLSList contains a list of GenTLS
type GenTLSList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GenTLS `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GenTLS{}, &GenTLSList{})
}
