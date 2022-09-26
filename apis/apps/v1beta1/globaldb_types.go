/*
AGPL License
Copyright 2022 ysicing(i@ysicing.me).
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:validation:Enum=new;exist
type State string

// GlobalDBSpec defines the desired state of GlobalDB
type GlobalDBSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Type    string `json:"type" yaml:"type"`
	State   State  `json:"state" yaml:"state"`
	Source  Source `json:"source,omitempty" yaml:"source,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

type Source struct {
	Host string `json:"host,omitempty" yaml:"host,omitempty"`
	Port int    `json:"port,omitempty" yaml:"port,omitempty"`
	User string `json:"user,omitempty" yaml:"user,omitempty"`
	Pass string `json:"pass,omitempty" yaml:"pass,omitempty"`
}

// GlobalDBStatus defines the observed state of GlobalDB
type GlobalDBStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Address  string `json:"address,omitempty" yaml:"address,omitempty"`
	Network  bool   `json:"network" yaml:"network"`
	Auth     bool   `json:"auth,omitempty" yaml:"auth,omitempty"`
	Ready    bool   `json:"ready" yaml:"ready"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}

//+genclient
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Network",type=boolean,JSONPath=`.status.network`
//+kubebuilder:printcolumn:name="Auth",type=boolean,JSONPath=`.status.auth`
//+kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
//+kubebuilder:printcolumn:name="Address",type=string,JSONPath=`.status.address`
//+kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:resource:path=globaldbs,shortName=gdb

// GlobalDB is the Schema for the globaldbs API
type GlobalDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GlobalDBSpec   `json:"spec,omitempty"`
	Status GlobalDBStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GlobalDBList contains a list of GlobalDB
type GlobalDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalDB `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GlobalDB{}, &GlobalDBList{})
}
