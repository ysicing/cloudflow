// AGPL License
// Copyright 2022 ysicing(i@ysicing.me).

package client

import (
	"fmt"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewClientFromManager(mgr manager.Manager, name string) client.Client {
	cfg := rest.CopyConfig(mgr.GetConfig())
	cfg.UserAgent = fmt.Sprintf("cloudflow-manager/%s", name)

	delegatingClient, _ := cluster.DefaultNewClient(mgr.GetCache(), cfg, client.Options{Scheme: mgr.GetScheme(), Mapper: mgr.GetRESTMapper()})
	return delegatingClient
}
