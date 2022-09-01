/*
Copyright 2022 ysicing(i@ysicing.me).
*/

package web

import (
	"context"
	"time"

	appsv1beta1 "github.com/ysicing/cloudflow/apis/apps/v1beta1"
	utilclient "github.com/ysicing/cloudflow/pkg/util/client"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	concurrentReconciles = 3
)

const (
	controllerName = "webapp-controller"
)

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	recorder := mgr.GetEventRecorderFor(controllerName)
	return &WebReconciler{
		Client:        utilclient.NewClientFromManager(mgr, controllerName),
		scheme:        mgr.GetScheme(),
		clock:         clock.RealClock{},
		eventRecorder: recorder,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r, MaxConcurrentReconciles: concurrentReconciles})
	if err != nil {
		return err
	}

	// Watch for changes to Web
	err = c.Watch(&source.Kind{Type: &appsv1beta1.Web{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch Deployment
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &deploymentHandler{Reader: mgr.GetCache()})
	if err != nil {
		return err
	}
	// Watch Service
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &svcHandler{Reader: mgr.GetCache()})
	if err != nil {
		return err
	}
	// Watch Ingress
	err = c.Watch(&source.Kind{Type: &networkingv1.Ingress{}}, &ingressHandler{Reader: mgr.GetCache()})
	if err != nil {
		return err
	}
	return nil
}

var _ reconcile.Reconciler = &WebReconciler{}

// WebReconciler reconciles a Web object
type WebReconciler struct {
	client.Client
	scheme        *runtime.Scheme
	clock         clock.Clock
	eventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=apps.ysicing.me,resources=webs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.ysicing.me,resources=webs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.ysicing.me,resources=webs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Web object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *WebReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, retErr error) {
	startTime := time.Now()
	defer func() {
		if retErr == nil {
			if res.Requeue || res.RequeueAfter > 0 {
				klog.Infof("Finished syncing Web %s, cost %v, result: %v", req, time.Since(startTime), res)
			} else {
				klog.Infof("Finished syncing Web %s, cost %v", req, time.Since(startTime))
			}
		} else {
			klog.Errorf("Failed syncing Web %s: %v", req, retErr)
		}
	}()
	// Fetch the Web instance
	instance := &appsv1beta1.Web{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		instance = nil
	}
	if instance == nil || instance.DeletionTimestamp != nil {
		klog.V(3).Infof("Web %s has been deleted.", req)
		return ctrl.Result{}, nil
	}
	klog.Infof("parse web %s", req)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WebReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1beta1.Web{}).
		Complete(r)
}
