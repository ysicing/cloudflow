// AGPL License
// Copyright 2022 ysicing(i@ysicing.me).

package web

import (
	"time"

	appsv1beta1 "github.com/ysicing/cloudflow/apis/apps/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	// initialingRateLimiter calculates the delay duration for existing Pods
	// triggered Create event when the Informer cache has just synced.
	initialingRateLimiter = workqueue.NewItemExponentialFailureRateLimiter(3*time.Second, 30*time.Second)
)

type deploymentHandler struct {
	client.Reader
}

var _ handler.EventHandler = &deploymentHandler{}

func (e *deploymentHandler) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	deploy := evt.Object.(*appsv1.Deployment)
	if deploy.DeletionTimestamp != nil {
		return
	}
	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(deploy); controllerRef != nil {
		req := resolveControllerRef(deploy.Namespace, controllerRef)
		if req == nil {
			return
		}
		klog.Infof("Deployment %s/%s created, owner: %s", deploy.Namespace, deploy.Name, req.Name)
		q.AddAfter(*req, initialingRateLimiter.When(req))
		return
	}
}

func (e *deploymentHandler) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	// oldDeploy := evt.ObjectOld.(*appsv1.Deployment)
	// curDeploy := evt.ObjectNew.(*appsv1.Deployment)
	// if curDeploy.ResourceVersion == oldDeploy.ResourceVersion {
	// 	return
	// }
	// if curDeploy.DeletionTimestamp != nil {
	// 	e.Delete(event.DeleteEvent{Object: evt.ObjectNew}, q)
	// 	return
	// }
}

func (e *deploymentHandler) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {}

func (e *deploymentHandler) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {

}

type svcHandler struct {
	client.Reader
}

var _ handler.EventHandler = &svcHandler{}

func (e *svcHandler) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {}

func (e *svcHandler) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {}

func (e *svcHandler) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {}

func (e *svcHandler) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {

}

type ingressHandler struct {
	client.Reader
}

var _ handler.EventHandler = &ingressHandler{}

func (e *ingressHandler) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {}

func (e *ingressHandler) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {}

func (e *ingressHandler) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {}

func (e *ingressHandler) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {

}

func resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *reconcile.Request {
	// Parse the Group out of the OwnerReference to compare it to what was parsed out of the requested OwnerType
	refGV, err := schema.ParseGroupVersion(controllerRef.APIVersion)
	if err != nil {
		klog.Errorf("Could not parse OwnerReference %v APIVersion: %v", controllerRef, err)
		return nil
	}

	// Compare the OwnerReference Group and Kind against the OwnerType Group and Kind specified by the user.
	// If the two match, create a Request for the objected referred to by
	// the OwnerReference.  Use the Name from the OwnerReference and the Namespace from the
	// object in the event.
	if controllerRef.Kind == appsv1beta1.SchemeGroupVersion.WithKind("Web").Kind && refGV.Group == appsv1beta1.SchemeGroupVersion.WithKind("Web").Group {
		// Match found - add a Request for the object referred to in the OwnerReference
		req := reconcile.Request{NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      controllerRef.Name,
		}}
		return &req
	}
	return nil
}
