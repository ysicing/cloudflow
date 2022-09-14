/*
Copyright 2022 ysicing(i@ysicing.me).
*/

package v1beta1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WebSpec defines the desired state of Web
type WebSpec struct {
	Image      string                      `json:"image"`
	Schedule   Schedule                    `json:"schedule,omitempty"`
	PullPolicy string                      `json:"pull,omitempty"`
	Replicas   *int32                      `json:"replicas,omitempty"`
	Resources  corev1.ResourceRequirements `json:"resources,omitempty"`
	// +optional
	Envs    []corev1.EnvVar `json:"envs,omitempty"`
	Volume  Volume          `json:"volume,omitempty"`
	Service Service         `json:"service,omitempty"`
	Ingress Ingress         `json:"ingress,omitempty"`
}

type Schedule struct {
	// NodeSelector select specific nodes
	NodeSelector map[string]string         `json:"nodeSelector,omitempty"`
	Strategy     appsv1.DeploymentStrategy `json:"strategy,omitempty"`
}

type Volume struct {
	Type       VolumeType    `json:"type,omitempty"`
	HostPath   string        `json:"hostpath,omitempty"`
	MountPaths []VolumeMount `json:"mountPaths"`
}

type VolumeType string

const (
	HostVolume VolumeType = "hostpath"
	PVCVolume  VolumeType = "pvc"
)

type VolumeMount struct {
	MountPath string `json:"mountPath"`
	ReadOnly  bool   `json:"read,omitempty"`
	SubPath   string `json:"subPath,omitempty"`
}

type Service struct {
	Type  string        `json:"type,omitempty"`
	Ports []ServicePort `json:"ports"`
}

type ServicePort struct {
	Name     string `json:"name,omitempty"`
	Port     int32  `json:"port"`
	NodePort int32  `json:"nodePort,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

type Ingress struct {
	Class       string `json:"class,omitempty"`
	Hostname    string `json:"hostname"`
	Port        int32  `json:"port"`
	ExternalDns bool   `json:"edns,omitempty"`
	TLS         string `json:"tls,omitempty"`
	// Whitelist  CIDRs 10.0.0.0/24,172.10.0.1
	Whitelist string `json:"whitelist,omitempty"`
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
//+kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

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
