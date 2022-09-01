/*
Copyright 2022 ysicing(i@ysicing.me).
*/

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WebSpec defines the desired state of Web
type WebSpec struct {
	Type            string                      `json:"type,omitempty"`
	Image           string                      `json:"image"`
	ImagePullPolicy string                      `json:"imagePullPolicy,omitempty"`
	Replicas        *int32                      `json:"replicas,omitempty"`
	RestartPolicy   string                      `json:"restartPolicy,omitempty"`
	Resources       corev1.ResourceRequirements `json:"resources,omitempty"`
	// +optional
	Envs    []corev1.EnvVar `json:"envs,omitempty"`
	Volume  Volume          `json:"volume,omitempty"`
	Service Service         `json:"service,omitempty"`
}

type Volume struct {
	Name       string        `json:"name"`
	Type       string        `json:"type"`
	Path       string        `json:"path,omitempty"`
	MountPaths []VolumeMount `json:"mountPaths"`
}

type VolumeMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
	SubPath   string `json:"subPath,omitempty"`
}

type Service struct {
	Type  string        `json:"type,omitempty"`
	Ports []ServicePort `json:"ports"`
}

type ServicePort struct {
	Name     string         `json:"name,omitempty"`
	Port     int32          `json:"port"`
	Protocol string         `json:"protocol,omitempty"`
	Ingress  ServiceIngress `json:"ingress,omitempty"`
}

type ServiceIngress struct {
	Class    string `json:"class,omitempty"`
	Hostname string `json:"hostname"`
	TLSName  string `json:"tlsName,omitempty"`
}

// WebStatus defines the observed state of Web
type WebStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Ready bool `json:"ready"`
}

//+genclient
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Web is the Schema for the webs API
type Web struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebSpec   `json:"spec,omitempty"`
	Status WebStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WebList contains a list of Web
type WebList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Web `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Web{}, &WebList{})
}
