/*
AGPL License
Copyright 2022 ysicing(i@ysicing.me).
*/

package gdb

import (
	"context"
	"fmt"
	"time"

	"github.com/ergoapi/util/ptr"
	appsv1beta1 "github.com/ysicing/cloudflow/apis/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/ysicing/cloudflow/pkg/controllers/gdb/util"
)

const (
	controllerName             = "globaldb-controller"
	gdbCreationDelayAfterReady = time.Second * 30
	minRequeueDuration         = time.Second * 5
)

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	recorder := mgr.GetEventRecorderFor(controllerName)
	return &GlobalDBReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		EventRecorder: recorder,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	// Watch for changes to GlobalDB
	err = c.Watch(&source.Kind{Type: &appsv1beta1.GlobalDB{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	return nil
}

var _ reconcile.Reconciler = &GlobalDBReconciler{}

// GlobalDBReconciler reconciles a GlobalDB object
type GlobalDBReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	EventRecorder record.EventRecorder
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GlobalDB object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *GlobalDBReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.Info("start reconcile for gdb")
	// fetch gdb
	gdb := &appsv1beta1.GlobalDB{}
	err := r.Get(ctx, req.NamespacedName, gdb)
	if err != nil {
		if !errors.IsNotFound(err) {
			return reconcile.Result{}, fmt.Errorf("failed to get gdb %s: %v", req.NamespacedName.Name, err)
		}
		gdb = nil
	}
	if gdb == nil || gdb.DeletionTimestamp != nil {
		klog.Info("gdb is deleted")
		return reconcile.Result{}, nil
	}
	// TODO if gdb is not exist, we should create it
	klog.Infof("gdb %s is new will create", gdb.Name)
	if len(gdb.Spec.Source.Pass) == 0 {
		// TODO gen password
		gdb.Spec.Source.Pass = "password"
	}
	if isReady, delay := r.getGDBReadyAndDelaytime(gdb); !isReady {
		klog.Infof("skip for gdb %s has not ready yet.", req.Name)
		return ctrl.Result{}, nil
	} else if delay > 0 {
		klog.Infof("skip for gdb %s waiting for ready %s.", req.Name, delay)
		return ctrl.Result{RequeueAfter: delay}, nil
	}
	if err := r.updateGDBStatus(gdb); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: minRequeueDuration}, nil
}

func (r *GlobalDBReconciler) updateGDB(gdb *appsv1beta1.GlobalDB, host string) error {
	gdb.Spec.Source.Host = host
	gdb.Spec.Source.User = "root"
	gdb.Spec.Source.Port = 3306
	gdb.Spec.State = "exist"
	return r.Update(context.TODO(), gdb)
}

func (r *GlobalDBReconciler) updateGDBStatus(gdb *appsv1beta1.GlobalDB) error {
	// update gdb status
	var gstatus appsv1beta1.GlobalDBStatus

	// check network
	dbtool := util.NewDB(gdb.Spec)
	if dbtool == nil {
		gstatus.Auth = false
		gstatus.Ready = false
		r.EventRecorder.Eventf(gdb, corev1.EventTypeWarning, "NotSupport", "Not NotSupport %v", gdb.Spec.Type)
	} else {
		if err := dbtool.CheckNetWork(); err != nil {
			gstatus.Network = false
			gstatus.Ready = false
			r.EventRecorder.Eventf(gdb, corev1.EventTypeWarning, "NetworkUnreachable", "Failed to conn %s:%v for %v", gdb.Spec.Source.Host, gdb.Spec.Source.Port, err)
		} else {
			gstatus.Network = true
			if err := dbtool.CheckAuth(); err != nil {
				gstatus.Auth = false
				gstatus.Ready = false
				r.EventRecorder.Eventf(gdb, corev1.EventTypeWarning, "AuthFailed", "Failed to auth %s:%v for %v", gdb.Spec.Source.Host, gdb.Spec.Source.Port, err)
			} else {
				gstatus.Auth = true
				gstatus.Ready = true
				r.EventRecorder.Eventf(gdb, corev1.EventTypeNormal, "Success", "Success to check %s:%v network & auth", gdb.Spec.Source.Host, gdb.Spec.Source.Port)
			}
		}
	}
	gstatus.Address = fmt.Sprintf("%s:%d", gdb.Spec.Source.Host, gdb.Spec.Source.Port)
	gdb.Status = gstatus
	return r.Status().Update(context.TODO(), gdb)
}

func (r *GlobalDBReconciler) getGDBReadyAndDelaytime(gdb *appsv1beta1.GlobalDB) (bool, time.Duration) {
	created, done, err := r.checkOrCreateGDBJob(gdb)
	if err != nil {
		klog.Errorf("check gdb job %s failed for %v", gdb.Name, err)
	}
	if !created {
		return false, 0
	}
	if !done {
		delay := gdbCreationDelayAfterReady - time.Since(gdb.CreationTimestamp.Time)
		if delay > 0 {
			return false, delay
		}
	}
	if err := r.updateGDB(gdb, fmt.Sprintf("%s-mysql.%s.svc", gdb.Name, gdb.Namespace)); err != nil {
		klog.Errorf("update gdb %s status failed for %v", gdb.Name, err)
		r.EventRecorder.Eventf(gdb, corev1.EventTypeWarning, "UpdateStatusFailed", "Failed to update gdb %s status", gdb.Name)
	} else {
		r.EventRecorder.Eventf(gdb, corev1.EventTypeNormal, "Success", "Success to update gdb %s status", gdb.Name)
	}

	return true, 0
}

func (r *GlobalDBReconciler) checkOrCreateGDBJob(gdb *appsv1beta1.GlobalDB) (bool, bool, error) {
	// create gdb job
	dbmeta := parsedbtype(gdb.Spec.Type, gdb.Spec.Source.Pass)
	gdbJob := &appsv1beta1.Web{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Web",
			APIVersion: "apps.ysicing.me/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gdb.Name,
			Namespace: gdb.Namespace,
		},
		Spec: appsv1beta1.WebSpec{
			Image:    dbmeta.Image,
			Replicas: ptr.Int32Ptr(1),
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"cpu":    *resource.NewQuantity(100, resource.DecimalSI),
					"memory": *resource.NewQuantity(536870912, resource.BinarySI),
				},
			},
			Envs: dbmeta.Env,
			Volume: appsv1beta1.Volume{
				Type: appsv1beta1.PVCVolume,
				MountPaths: []appsv1beta1.VolumeMount{
					{
						MountPath: dbmeta.MountPath,
					},
				},
			},
			Service: appsv1beta1.Service{
				Ports: []appsv1beta1.ServicePort{
					{
						Port: dbmeta.Port,
					},
				},
			},
		},
	}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: gdb.Name, Namespace: gdb.Namespace}, gdbJob); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(context.TODO(), gdbJob); err != nil {
				return false, false, err
			}
			klog.Infof("create gdb job %s success", gdbJob.Name)
			r.EventRecorder.Eventf(gdb, corev1.EventTypeNormal, "Success", "Success to create gdb job %s", gdbJob.Name)
			return true, false, nil
		}
		return false, false, err
	}
	if getGDBJobStatus(&gdbJob.Status) {
		return true, true, nil
	}
	return true, false, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GlobalDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1beta1.GlobalDB{}).
		Complete(r)
}

func getGDBJobStatus(status *appsv1beta1.WebStatus) bool {
	if status == nil {
		return false
	}
	return status.Ready
}

type Addon struct {
	Image     string
	Port      int32
	MountPath string
	Env       []corev1.EnvVar
}

func parsedbtype(gdbtype, password string) Addon {
	var db Addon
	switch gdbtype {
	case "redis":
		db.Image = "bitnami/redis"
		db.Port = 6379
		db.MountPath = "/bitnami/redis/data"
		db.Env = []corev1.EnvVar{
			{
				Name:  "REDIS_DISABLE_COMMANDS",
				Value: "FLUSHDB,FLUSHALL,CONFIG",
			},
			{
				Name:  "REDIS_PASSWORD",
				Value: password,
			},
		}
	case "etcd":
		db.Image = "bitnami/etcd"
		db.MountPath = "/bitnami/etcd"
		db.Port = 2379
		db.Env = []corev1.EnvVar{
			{
				Name:  "ALLOW_NONE_AUTHENTICATION",
				Value: "yes",
			},
		}
	case "mariadb":
		db.Image = "bitnami/mariadb"
		db.MountPath = "/bitnami/mariadb"
		db.Port = 3306
		db.Env = []corev1.EnvVar{
			{
				Name:  "MARIADB_ROOT_PASSWORD",
				Value: password,
			},
		}
	case "mongodb":
		db.Image = "bitnami/mongodb"
		db.MountPath = "/bitnami/mongodb"
		db.Port = 27017
		db.Env = []corev1.EnvVar{
			{
				Name:  "MONGODB_ROOT_PASSWORD",
				Value: password,
			},
		}
	case "postgresql":
		db.Image = "bitnami/postgresql"
		db.Port = 5432
		db.MountPath = "/bitnami/postgresql"
		db.Env = []corev1.EnvVar{
			{
				Name:  "POSTGRESQL_PASSWORD",
				Value: password,
			},
		}
	default:
		db.Image = "bitnami/mysql"
		db.Port = 3306
		db.MountPath = "/bitnami/mysql/data"
		db.Env = []corev1.EnvVar{
			{
				Name:  "MYSQL_ROOT_PASSWORD",
				Value: password,
			},
		}
	}
	return db
}
