/*
AGPL License
Copyright 2022 ysicing(i@ysicing.me).
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1beta1 "github.com/ysicing/cloudflow/pkg/client/clientset/versioned/typed/apps/v1beta1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeAppsV1beta1 struct {
	*testing.Fake
}

func (c *FakeAppsV1beta1) GlobalDBs(namespace string) v1beta1.GlobalDBInterface {
	return &FakeGlobalDBs{c, namespace}
}

func (c *FakeAppsV1beta1) Webs(namespace string) v1beta1.WebInterface {
	return &FakeWebs{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeAppsV1beta1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
