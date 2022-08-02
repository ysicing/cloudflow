// AGPL License
// Copyright 2022 ysicing(i@ysicing.me).

package v1beta1

import "k8s.io/apimachinery/pkg/runtime/schema"

// Resource is required by pkg/client/listers/...
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}
