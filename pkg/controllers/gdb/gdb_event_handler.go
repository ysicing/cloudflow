package gdb

import (
	"time"

	appsv1beta1 "github.com/ysicing/cloudflow/apis/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
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

type webHandler struct {
	client.Reader
}

var _ handler.EventHandler = &webHandler{}

func (e *webHandler) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	web := evt.Object.(*appsv1beta1.Web)
	if web.DeletionTimestamp != nil {
		return
	}
	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(web); controllerRef != nil {
		req := resolveControllerRef(web.Namespace, controllerRef)
		if req == nil {
			return
		}
		klog.Infof("Web %s/%s created, owner: %s", web.Namespace, web.Name, req.Name)
		q.AddAfter(*req, initialingRateLimiter.When(req))
		return
	}
}

func (e *webHandler) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	oldWeb := evt.ObjectOld.(*appsv1beta1.Web)
	curWeb := evt.ObjectNew.(*appsv1beta1.Web)
	if curWeb.ResourceVersion == oldWeb.ResourceVersion {
		return
	}
	if curWeb.DeletionTimestamp != nil {
		e.Delete(event.DeleteEvent{Object: evt.ObjectNew}, q)
		return
	}
}

func (e *webHandler) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {}

func (e *webHandler) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {

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
	if controllerRef.Kind == appsv1beta1.SchemeGroupVersion.WithKind("GlobalDB").Kind && refGV.Group == appsv1beta1.SchemeGroupVersion.WithKind("GlobalDB").Group {
		// Match found - add a Request for the object referred to in the OwnerReference
		req := reconcile.Request{NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      controllerRef.Name,
		}}
		return &req
	}
	return nil
}
