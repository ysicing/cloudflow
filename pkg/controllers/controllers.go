// AGPL License
// Copyright 2022 ysicing(i@ysicing.me).

package controllers

import (
	coreapp "github.com/ysicing/cloudflow/pkg/controllers/core"
	gdbapp "github.com/ysicing/cloudflow/pkg/controllers/gdb"
	webapp "github.com/ysicing/cloudflow/pkg/controllers/web"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var controllerAddFuncs []func(manager.Manager) error

func init() {
	controllerAddFuncs = append(controllerAddFuncs, webapp.Add)
	controllerAddFuncs = append(controllerAddFuncs, gdbapp.Add)
	controllerAddFuncs = append(controllerAddFuncs, coreapp.Add)
}

func SetupWithManager(m manager.Manager) error {
	for _, f := range controllerAddFuncs {
		if err := f(m); err != nil {
			if kindMatchErr, ok := err.(*meta.NoKindMatchError); ok {
				klog.Infof("CRD %v is not installed, its controller will perform noops!", kindMatchErr.GroupKind)
				continue
			}
			return err
		}
	}
	return nil
}
