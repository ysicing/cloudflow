/*
AGPL License
Copyright 2022 ysicing(i@ysicing.me).
*/

package clusterendpoint

import (
	"context"

	utilclient "github.com/ysicing/cloudflow/pkg/util/client"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	networkv1beta1 "github.com/ysicing/cloudflow/apis/network/v1beta1"
)

var (
	concurrentReconciles = 1
)

const (
	controllerName = "clusterendpoint-controller"
)

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	recorder := mgr.GetEventRecorderFor(controllerName)
	return &ClusterEndpointReconciler{
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

	// Watch for changes to ClusterEndpoint
	err = c.Watch(&source.Kind{Type: &networkv1beta1.ClusterEndpoint{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ClusterEndpointReconciler{}

// ClusterEndpointReconciler reconciles a ClusterEndpoint object
type ClusterEndpointReconciler struct {
	client.Client
	scheme        *runtime.Scheme
	clock         clock.Clock
	eventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=network.ysicing.me,resources=clusterendpoints,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=network.ysicing.me,resources=clusterendpoints/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=network.ysicing.me,resources=clusterendpoints/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClusterEndpoint object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *ClusterEndpointReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterEndpointReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkv1beta1.ClusterEndpoint{}).
		Complete(r)
}
