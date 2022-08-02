/*
Copyright 2022 ysicing(i@ysicing.me).
*/

// Package v1beta1 contains API Schema definitions for the apps v1beta1 API group
//+kubebuilder:object:generate=true
//+groupName=apps.ysicing.me
package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "apps.ysicing.me", Version: "v1beta1"}

	SchemeGroupVersion = GroupVersion

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
