/*
Copyright 2022 ysicing(i@ysicing.me).
*/

package web

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ergoapi/util/exmap"
	"github.com/ergoapi/util/ptr"
	appsv1beta1 "github.com/ysicing/cloudflow/apis/apps/v1beta1"
	utilclient "github.com/ysicing/cloudflow/pkg/util/client"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
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
	concurrentReconciles = 1
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
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	// Watch Service
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	// Watch Ingress
	err = c.Watch(&source.Kind{Type: &networkingv1.Ingress{}}, &handler.EnqueueRequestForObject{})
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
	ctx = context.WithValue(ctx, "request", req)
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
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		instance = nil
	}
	// Web 已经被删除
	if instance == nil || instance.DeletionTimestamp != nil {
		klog.Infof("Web %s has been deleted.", req)
		return ctrl.Result{}, nil
	}
	klog.Infof("parse web %s", req)
	if err := r.cleanOrCreatePVC(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}
	deployRes := &appsv1.Deployment{}
	if err := r.Client.Get(ctx, req.NamespacedName, deployRes); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		if err = r.createDeploy(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if err = r.updateDeploy(ctx, deployRes, instance); err != nil {
			return ctrl.Result{}, err
		}
	}
	if err := r.syncService(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.updateWebStatus(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *WebReconciler) updateWebStatus(ctx context.Context, web *appsv1beta1.Web) error {
	var ws appsv1beta1.WebStatus
	ws.Ready = true
	web.Status = ws
	return r.Status().Update(ctx, web)
}

func (r *WebReconciler) createDeploy(ctx context.Context, web *appsv1beta1.Web) error {
	klog.Infof("start create deploy %s", web.Name)
	deploy := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            web.Name,
			Namespace:       web.Namespace,
			Labels:          getLabels(web.GetLabels(), web.Name),
			Annotations:     web.GetAnnotations(),
			OwnerReferences: ownerReference(web),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: web.Spec.Replicas,
			Strategy: web.Spec.Schedule.Strategy,
			Selector: &metav1.LabelSelector{
				MatchLabels: getLabels(web.GetLabels(), web.Name),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: getLabels(web.GetLabels(), web.Name),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            web.Name,
							Image:           web.Spec.Image,
							ImagePullPolicy: getPullPolicy(web.Spec.PullPolicy),
							Env:             web.Spec.Envs,
							Resources:       web.Spec.Resources,
							VolumeMounts:    getVolumeMounts(web.Name, web.Spec.Volume),
						},
					},
					Volumes:      getVolume(web.Name, web.Spec.Volume),
					NodeSelector: web.Spec.Schedule.NodeSelector,
					Tolerations:  getTolerations(web.Spec.Schedule),
				},
			},
		},
	}
	if err := r.Client.Create(ctx, deploy); err != nil {
		return err
	}
	r.eventRecorder.Eventf(web, corev1.EventTypeNormal, "create", fmt.Sprintf("create deployment %s", deploy.Name))
	return nil
}

func (r *WebReconciler) updateDeploy(ctx context.Context, deploy *appsv1.Deployment, web *appsv1beta1.Web) error {
	klog.Infof("%s check deploy, will update", web.Name)
	newContainer := deploy.Spec.Template.Spec.Containers[0]
	newContainer.Image = web.Spec.Image
	newContainer.Env = web.Spec.Envs
	newContainer.ImagePullPolicy = getPullPolicy(web.Spec.PullPolicy)
	newContainer.Resources = web.Spec.Resources
	newContainer.VolumeMounts = getVolumeMounts(web.Name, web.Spec.Volume)
	newVolume := getVolume(web.Name, web.Spec.Volume)
	klog.Infof("update deploy %s", web.Name)
	deploy.Spec.Replicas = web.Spec.Replicas
	deploy.Spec.Template.Spec.Volumes = newVolume
	deploy.Spec.Template.Spec.Containers[0] = newContainer
	deploy.Spec.Template.Spec.NodeSelector = web.Spec.Schedule.NodeSelector
	deploy.Spec.Template.Spec.Tolerations = getTolerations(web.Spec.Schedule)
	deploy.Spec.Strategy = web.Spec.Schedule.Strategy
	if err := r.Client.Update(ctx, deploy); err != nil {
		return err
	}
	r.eventRecorder.Eventf(web, corev1.EventTypeNormal, "update", "update deployment done")
	return nil
}

func (r *WebReconciler) cleanOrCreatePVC(ctx context.Context, web *appsv1beta1.Web) error {
	needpvc := false
	if web.Spec.Volume.Type == appsv1beta1.PVCVolume {
		needpvc = true
	}
	objectKey := ctx.Value("request").(ctrl.Request).NamespacedName
	pvc := &corev1.PersistentVolumeClaim{}
	if err := r.Get(ctx, objectKey, pvc); err != nil {
		if errors.IsNotFound(err) {
			if needpvc {
				pvc.TypeMeta = metav1.TypeMeta{Kind: "PersistentVolumeClaim", APIVersion: "v1"}
				pvc.ObjectMeta = metav1.ObjectMeta{
					Name:            web.Name,
					Namespace:       web.Namespace,
					Labels:          getLabels(web.GetLabels(), web.Name),
					Annotations:     web.GetAnnotations(),
					OwnerReferences: ownerReference(web),
				}
				pvc.Spec = corev1.PersistentVolumeClaimSpec{
					AccessModes: r.fakePVCMode(),
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"storage": *resource.NewQuantity(1073741824, resource.BinarySI), // 1Gi
						},
					},
				}
				if err := r.Create(ctx, pvc); err != nil {
					r.eventRecorder.Eventf(web, corev1.EventTypeWarning, "create", fmt.Sprintf("create pvc failed, err: %v", err))
					return err
				}
				r.eventRecorder.Eventf(web, corev1.EventTypeNormal, "create", fmt.Sprintf("create pvc %s done", web.Name))
			}
			return nil
		}
		return err
	}
	if !needpvc {
		if err := r.Delete(ctx, pvc); err != nil {
			r.eventRecorder.Eventf(web, corev1.EventTypeWarning, "delete", fmt.Sprintf("delete pvc failed, err: %v", err))
			return err
		}
		r.eventRecorder.Eventf(web, corev1.EventTypeNormal, "delete", fmt.Sprintf("create pvc %s done", web.Name))
	}
	return nil
}

func (r *WebReconciler) fakePVCMode() []corev1.PersistentVolumeAccessMode {
	var pvcmode []corev1.PersistentVolumeAccessMode
	if r.checkLocalPathStorageClass() {
		return append(pvcmode, corev1.ReadWriteOnce)
	}
	return append(pvcmode, corev1.ReadWriteMany)
}

func (r *WebReconciler) checkLocalPathStorageClass() bool {
	sc := &storagev1.StorageClassList{}
	if err := r.Client.List(context.TODO(), sc); err != nil {
		return false
	}
	existlocalpath := false
	for _, s := range sc.Items {
		if exmap.GetLabelValue(s.GetAnnotations(), "storageclass.kubernetes.io/is-default-class") == "true" && strings.Contains(s.Provisioner, "local") {
			existlocalpath = true
		}
	}
	return existlocalpath
}

func (r *WebReconciler) syncService(ctx context.Context, web *appsv1beta1.Web) error {
	klog.Infof("sync Service %s", web.Name)
	clean := false
	if len(web.Spec.Service.Ports) == 0 {
		clean = true
	}
	if clean {
		klog.Info("clean service & ingress")
		// 清理service & ingress
		if err := r.cleanService(ctx, web); err != nil {
			r.eventRecorder.Eventf(web, corev1.EventTypeWarning, "delete", fmt.Sprintf("clean service failed, err: %v", err))
			return err
		}
		if err := r.cleanIngress(ctx, web); err != nil {
			r.eventRecorder.Eventf(web, corev1.EventTypeWarning, "delete", fmt.Sprintf("clean ingress failed, err: %v", err))
			return err
		}
		return nil
	}
	if err := r.createService(ctx, web); err != nil {
		r.eventRecorder.Eventf(web, corev1.EventTypeWarning, "create", fmt.Sprintf("create service failed, err: %v", err))
		return err
	}
	if len(web.Spec.Ingress.Hostname) == 0 {
		klog.Info("clean ingress")
		if err := r.cleanIngress(ctx, web); err != nil {
			r.eventRecorder.Eventf(web, corev1.EventTypeWarning, "delete", fmt.Sprintf("clean ingress failed, err: %v", err))
			return err
		}
		return nil
	}
	if err := r.createIngress(ctx, web); err != nil {
		r.eventRecorder.Eventf(web, corev1.EventTypeWarning, "create", fmt.Sprintf("create ingress failed, err: %v", err))
		return err
	}
	return nil
}

func (r *WebReconciler) getIngressAnnotations(web *appsv1beta1.Web) map[string]string {
	old := web.GetAnnotations()
	if web.Spec.Ingress.ExternalDns {
		old = exmap.MergeLabels(old, map[string]string{"external-dns.alpha.kubernetes.io/hostname": web.Spec.Ingress.Hostname})
	}
	if len(web.Spec.Ingress.Whitelist) > 0 {
		old = exmap.MergeLabels(old, map[string]string{"nginx.ingress.kubernetes.io/whitelist-source-range": web.Spec.Ingress.Hostname})
	}
	return old
}

func (r *WebReconciler) createIngress(ctx context.Context, web *appsv1beta1.Web) error {
	objectKey := ctx.Value("request").(ctrl.Request).NamespacedName
	ingress := &networkingv1.Ingress{}
	newingress := &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{Kind: "Ingress", APIVersion: "networking.k8s.io/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:            web.Name,
			Namespace:       web.Namespace,
			Labels:          getLabels(web.GetLabels(), web.Name),
			Annotations:     r.getIngressAnnotations(web),
			OwnerReferences: ownerReference(web),
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: web.Spec.Ingress.Hostname,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: ptrPathType("Prefix"),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: web.Name,
											Port: networkingv1.ServiceBackendPort{
												Number: web.Spec.Ingress.Port,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if len(web.Spec.Ingress.Class) > 0 {
		newingress.Spec.IngressClassName = ptr.StringPtr(web.Spec.Ingress.Class)
	}

	if len(web.Spec.Ingress.TLS) > 0 {
		newingress.Spec.TLS = append(ingress.Spec.TLS, networkingv1.IngressTLS{
			Hosts:      []string{web.Spec.Ingress.Hostname},
			SecretName: web.Spec.Ingress.TLS,
		})
	}

	if err := r.Client.Get(ctx, objectKey, ingress); err != nil {
		if errors.IsNotFound(err) {
			klog.Infof("create ingress %s/%s", web.Name, web.Namespace)
			if err = r.Client.Create(ctx, newingress); err != nil {
				return err
			}
			r.eventRecorder.Eventf(web, corev1.EventTypeNormal, "create", "create ingress done")
			return nil
		}
		return err
	}
	klog.Infof("update ingress %s/%s", web.Name, web.Namespace)
	if err := r.Client.Update(ctx, newingress); err != nil {
		return err
	}
	r.eventRecorder.Eventf(web, corev1.EventTypeNormal, "update", "update ingress done")
	return nil
}

func getPortName(name string, port int32) string {
	if len(name) > 0 {
		return name
	}
	return fmt.Sprintf("http-%d", port)
}

func (r *WebReconciler) createService(ctx context.Context, web *appsv1beta1.Web) error {
	objectKey := ctx.Value("request").(ctrl.Request).NamespacedName
	var ports []corev1.ServicePort
	for _, i := range web.Spec.Service.Ports {
		ports = append(ports, corev1.ServicePort{
			Name: getPortName(i.Name, i.Port),
			Port: i.Port,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Type(0),
				IntVal: i.Port,
			},
			Protocol: portProtocol(i.Protocol),
		})
	}
	service := &corev1.Service{}
	newservice := &corev1.Service{
		TypeMeta: metav1.TypeMeta{Kind: "Service", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:            web.Name,
			Namespace:       web.Namespace,
			Labels:          getLabels(web.GetLabels(), web.Name),
			Annotations:     web.GetAnnotations(),
			OwnerReferences: ownerReference(web),
		},
		Spec: corev1.ServiceSpec{
			Type:     getServiceType(web.Spec.Service.Type),
			Selector: getLabels(web.GetLabels(), web.Name),
		},
	}
	newservice.Spec.Ports = ports
	if err := r.Client.Get(ctx, objectKey, service); err != nil {
		if errors.IsNotFound(err) {
			klog.Infof("create service %s/%s", web.Name, web.Namespace)
			if err = r.Client.Create(ctx, newservice); err != nil {
				return err
			}
			r.eventRecorder.Eventf(web, corev1.EventTypeNormal, "create", "create service done")
			return nil
		}
		return err
	}
	klog.Infof("update service %s/%s", web.Name, web.Namespace)
	if err := r.Client.Update(ctx, newservice); err != nil {
		return err
	}
	r.eventRecorder.Eventf(web, corev1.EventTypeNormal, "update", "update service done")
	return nil
}

func (r *WebReconciler) cleanService(ctx context.Context, web *appsv1beta1.Web) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		objKey := client.ObjectKey{
			Namespace: web.Namespace,
			Name:      web.Name,
		}
		svc := &corev1.Service{}
		if err := r.Get(ctx, objKey, svc); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return err
		}
		if err := r.Delete(ctx, svc); err != nil {
			return err
		}
		r.eventRecorder.Eventf(web, corev1.EventTypeNormal, "delete", "clean service done")
		return nil
	})
}

func (r *WebReconciler) cleanIngress(ctx context.Context, web *appsv1beta1.Web) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		objKey := client.ObjectKey{
			Namespace: web.Namespace,
			Name:      web.Name,
		}
		ingress := &networkingv1.Ingress{}
		if err := r.Get(ctx, objKey, ingress); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return err
		}
		if err := r.Delete(ctx, ingress); err != nil {
			return err
		}
		r.eventRecorder.Eventf(web, corev1.EventTypeNormal, "delete", "clean ingress done")
		return nil
	})
}

// SetupWithManager sets up the controller with the Manager.
func (r *WebReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1beta1.Web{}).
		Complete(r)
}

func getLabels(old map[string]string, name string) map[string]string {
	return exmap.MergeLabels(old, map[string]string{"app.ysicing.me/name": name})
}

func ownerReference(obj metav1.Object) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(obj,
			schema.GroupVersionKind{
				Group:   appsv1beta1.SchemeGroupVersion.WithKind("Web").Group,
				Version: appsv1beta1.SchemeGroupVersion.WithKind("Web").Version,
				Kind:    appsv1beta1.SchemeGroupVersion.WithKind("Web").Kind,
			}),
	}
}

func portProtocol(p string) corev1.Protocol {
	if strings.ToLower(p) == "udp" {
		return corev1.ProtocolUDP
	}
	return corev1.ProtocolTCP
}

func getServiceType(p string) corev1.ServiceType {
	if strings.ToLower(p) == "lb" || strings.ToLower(p) == "loadbalancer" {
		return corev1.ServiceTypeLoadBalancer
	}
	if strings.ToLower(p) == "node" || strings.ToLower(p) == "nodeport" {
		return corev1.ServiceTypeNodePort
	}
	return corev1.ServiceTypeClusterIP
}

func ptrPathType(p networkingv1.PathType) *networkingv1.PathType {
	return &p
}

func getPullPolicy(t string) corev1.PullPolicy {
	if strings.ToLower(t) == "never" {
		return corev1.PullNever
	}
	if strings.ToLower(t) == "ifnotpresent" {
		return corev1.PullIfNotPresent
	}
	return corev1.PullAlways
}

func getVolumeMounts(name string, volume appsv1beta1.Volume) []corev1.VolumeMount {
	if len(volume.MountPaths) == 0 {
		return nil
	}
	var vms []corev1.VolumeMount
	for _, vm := range volume.MountPaths {
		vms = append(vms, corev1.VolumeMount{
			Name:      name,
			MountPath: vm.MountPath,
			ReadOnly:  vm.ReadOnly,
			SubPath:   vm.SubPath,
		})
	}
	return vms
}

func getVolume(name string, volume appsv1beta1.Volume) []corev1.Volume {
	if len(volume.MountPaths) == 0 {
		return nil
	}
	var vm corev1.VolumeSource
	if volume.Type == "nas" || volume.Type == "nfs" || volume.Type == "hostpath" {
		vm.HostPath = &corev1.HostPathVolumeSource{
			Path: volume.HostPath,
			Type: ptrHostPathType(corev1.HostPathDirectoryOrCreate),
		}
	} else if volume.Type == "pvc" {
		vm.PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{
			ClaimName: name,
		}
	} else {
		vm.EmptyDir = &corev1.EmptyDirVolumeSource{}
	}
	return []corev1.Volume{
		{
			Name:         name,
			VolumeSource: vm,
		},
	}
}

func getTolerations(t appsv1beta1.Schedule) []corev1.Toleration {
	if t.NodeSelector == nil {
		return nil
	}
	return []corev1.Toleration{
		{
			Operator: corev1.TolerationOperator("Exists"),
		},
	}
}
func ptrHostPathType(p corev1.HostPathType) *corev1.HostPathType {
	return &p
}
