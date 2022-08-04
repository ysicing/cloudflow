/*
AGPL License
Copyright 2022 ysicing(i@ysicing.me).
*/

package v1beta1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ServicePort struct {
	// The IP protocol for this port. Supports "TCP", "UDP", and "SCTP".
	// Default is TCP.
	// +optional
	Protocol v1.Protocol `json:"protocol,omitempty"`
	// The port that will be exposed by this service.
	Port int32 `json:"port"`
	// Number of seconds after which the probe times out.
	// Defaults to 1 second. Minimum value is 1.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	// +optional
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty"`
}

// ClusterEndpointSpec defines the desired state of ClusterEndpoint
type ClusterEndpointSpec struct {
	ClusterIP string        `json:"clusterIP,omitempty"`
	Host      string        `json:"host"`
	Ports     []ServicePort `json:"ports"`
}

type Phase string

// These are the valid phases of node.
const (
	// Pending means the node has been created/added by the system.
	Pending Phase = "Pending"
	// Healthy means the cluster service is healthy.
	Healthy Phase = "Healthy"
	// UnHealthy means the cluster service is not healthy.
	UnHealthy Phase = "UnHealthy"
)

type ConditionType string

const (
	SyncServiceReady  ConditionType = "SyncServiceReady"
	SyncEndpointReady ConditionType = "SyncEndpointReady"
	Initialized       ConditionType = "Initialized"
	Ready             ConditionType = "Ready"
)

type Condition struct {
	Type ConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=ConditionType"`
	// Status is the status of the condition. One of True, False, Unknown.
	Status v1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// LastHeartbeatTime is the last time this condition was updated.
	// +optional
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime,omitempty" protobuf:"bytes,3,opt,name=lastHeartbeatTime"`
	// LastTransitionTime is the last time the condition changed from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// Reason is a (brief) reason for the condition's last status change.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Message is a human-readable message indicating details about the last status change.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

// ClusterEndpointStatus defines the observed state of ClusterEndpoint
type ClusterEndpointStatus struct {
	Phase      Phase       `json:"phase,omitempty"`
	Conditions []Condition `json:"conditions"`
}

//+genclient
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ClusterEndpoint is the Schema for the clusterendpoints API
type ClusterEndpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterEndpointSpec   `json:"spec,omitempty"`
	Status ClusterEndpointStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=cep
//+kubebuilder:printcolumn:name="Age",type=date,description="The creation date",JSONPath=`.metadata.creationTimestamp`,priority=0
//+kubebuilder:printcolumn:name="Status",type=string,description="The status",JSONPath=`.status.phase`,priority=0
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterEndpointList contains a list of ClusterEndpoint
type ClusterEndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterEndpoint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterEndpoint{}, &ClusterEndpointList{})
}
