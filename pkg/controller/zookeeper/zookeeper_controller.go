package zookeeper

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	wnohangv1alpha1 "github.com/ronin13/zookeeper-operator/pkg/apis/wnohang/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("Zookeeper Operator")

const (
	reconcileInterval = 10
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Zookeeper Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileZookeeper{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("zookeeper-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Zookeeper
	err = c.Watch(&source.Kind{Type: &wnohangv1alpha1.Zookeeper{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Zookeeper
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &wnohangv1alpha1.Zookeeper{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileZookeeper{}

// ReconcileZookeeper reconciles a Zookeeper object
type ReconcileZookeeper struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	zids   string
}

func (r *ReconcileZookeeper) getClient(request reconcile.Request) (*wnohangv1alpha1.Zookeeper, error) {
	// Fetch the Zookeeper instance
	instance := &wnohangv1alpha1.Zookeeper{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Zookeeper resource not found. Ignoring since object must be deleted.")
			return nil, nil
		}

		log.Error(err, "Failed to get zookeeper instance")
		// Error reading the object - requeue the request.
		return nil, err
	}
	return instance, nil
}

// Reconcile reads that state of the cluster for a Zookeeper object and makes changes based on the state read
// and what is in the Zookeeper.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileZookeeper) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	var zooSet *appsv1.StatefulSet
	var createdSet, createdService bool
	var desPartition int32

	var defaultPartition int32
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Zookeeper")

	instance, err := r.getClient(request)
	if err != nil {
		return reconcile.Result{}, err
	}
	if instance == nil {
		return reconcile.Result{}, nil
	}

	createdService, err = r.createService(instance)
	if err != nil {
		reqLogger.Error(err, "Failed to create Service for statefulset")
		return reconcile.Result{}, err
	}

	if createdService {
		reqLogger.Info("Zookeeper Service created")
		return reconcile.Result{RequeueAfter: time.Second * 10}, nil
	}

	zooSet, createdSet, err = r.createStatefulSet(instance)
	if err != nil {
		reqLogger.Error(err, "Failed to create statefulset")
		return reconcile.Result{}, err
	}

	if createdSet {
		reqLogger.Info(fmt.Sprintf("Zookeeper Cluster created with %d  nodes", int(*zooSet.Spec.Replicas)))
		return reconcile.Result{RequeueAfter: time.Second * 30}, nil
	}

	reqLogger.Info(fmt.Sprintf("Zookeeper Cluster exists with %d  nodes", int(*zooSet.Spec.Replicas)))

	desSize := instance.Spec.Nodes
	curSize := *zooSet.Spec.Replicas
	reqZids := getZooIds(desSize)
	if curSize != desSize {

		reqLogger.Info(fmt.Sprintf("Updating statefulset to %d nodes", desSize))
		// currentZids := getZooIds(*zooSet.Spec.Replicas)
		if r.zids != reqZids {
			reqLogger.Info(fmt.Sprintf("Updating existing pods for new pods from %s to %s", r.zids, reqZids))
			zooSet.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{
				{
					Name:  "ZOO_IDS",
					Value: reqZids,
				},
			}
			if curSize < desSize {
				desPartition = curSize

			} else {
				desPartition = desSize

			}
			zooSet.Spec.Replicas = &desSize
			zooSet.Spec.UpdateStrategy.RollingUpdate.Partition = &desPartition

			err = r.client.Update(context.TODO(), zooSet)
			if err != nil {
				reqLogger.Error(err, "Failed to update existing Statefulset")
				return reconcile.Result{}, err
			}
			r.zids = reqZids
			time.Sleep(30 * time.Second)
			return reconcile.Result{Requeue: true}, nil
		}
	}

	/* 	if r.zids != reqZids {
	   		reqLogger.Info(fmt.Sprintf("Updating existing pods for new pods from %s to %s", r.zids, reqZids))
	   		zooSet.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{
	   			{
	   				Name:  "ZOO_IDS",
	   				Value: reqZids,
	   			},
	   		}

	   		err = r.client.Update(context.TODO(), zooSet)
	   		if err != nil {
	   			reqLogger.Error(err, "Failed to update existing Statefulset")
	   			return reconcile.Result{}, err
	   		}
	   		r.zids = reqZids
	   		time.Sleep(30 * time.Second)
	   		return reconcile.Result{Requeue: true}, nil
	   	}
	*/
	if *zooSet.Spec.UpdateStrategy.RollingUpdate.Partition != 0 {
		reqLogger.Info("Adding new nodes")

		zooSet.Spec.UpdateStrategy.RollingUpdate.Partition = &defaultPartition
		err = r.client.Update(context.TODO(), zooSet)
		if err != nil {
			reqLogger.Error(err, "Failed to update  node count")
			return reconcile.Result{}, err
		}

		time.Sleep(30 * time.Second)
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	// Pod already exists - don't requeue
	return reconcile.Result{RequeueAfter: time.Second * reconcileInterval}, nil
}

func (r *ReconcileZookeeper) createService(instance *wnohangv1alpha1.Zookeeper) (bool, error) {
	creatLogger := log.WithValues("Instance.Namespace", instance.Namespace, "Instance.Name", instance.Name)
	foundService := &corev1.Service{}
	zooService, err := newServiceforCR(instance)
	if err != nil {
		// return err
	}
	// Set Zookeeper instance as the owner and controller
	if err = controllerutil.SetControllerReference(instance, zooService, r.scheme); err != nil {
		return false, err
	}

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: zooService.Name, Namespace: zooService.Namespace}, foundService)
	if err != nil {
		if errors.IsNotFound(err) {
			creatLogger.Info("Creating a new Service", "Service.Namespace", zooService.Namespace, "Service.Name", zooService.Name)
			err = r.client.Create(context.TODO(), zooService)
			if err != nil {
				return false, err
			}
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func (r *ReconcileZookeeper) createStatefulSet(instance *wnohangv1alpha1.Zookeeper) (*appsv1.StatefulSet, bool, error) {
	var err error
	var zooSet *appsv1.StatefulSet
	creatLogger := log.WithValues("Instance.Namespace", instance.Namespace, "Instance.Name", instance.Name)

	found := &appsv1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			creatLogger.Info("Creating a new StatefulSet", "StatefulSet.Namespace", instance.Namespace, "StatefulSet.Name", instance.Name)
			r.zids = getZooIds(instance.Spec.Nodes)
			zooSet, err = newStatefulSetForCR(instance, r.zids)
			if err != nil {
				return nil, false, fmt.Errorf("Failed get stateful set definition")
			}
			err = r.client.Create(context.TODO(), zooSet)
			if err != nil {
				return nil, true, err
			}
			// Set Zookeeper instance as the owner and controller
			if err = controllerutil.SetControllerReference(instance, zooSet, r.scheme); err != nil {
				return nil, false, err
			}
			// instance.Status.Sids = zids
			return zooSet, true, nil
		}
		return nil, false, err
	}

	// creatLogger.Info("Skip reconcile: StatefulSet already exists", "StatefulSet.Namespace", found.Namespace, "StatefulSet.Name", found.Name)
	return found, false, nil

}

func newServiceforCR(cr *wnohangv1alpha1.Zookeeper) (*corev1.Service, error) {
	serviceName := cr.Name + "-serv"
	labels := map[string]string{
		"app": serviceName,
	}
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": cr.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name: "client",
					Port: 2181,
				},
				{
					Name: "leader-election",
					Port: 2888,
				},
				{
					Name: "cluster-comms",
					Port: 3888,
				},
			},
			ClusterIP:                "None",
			PublishNotReadyAddresses: true,
			// Type: corev1.ServiceTypeNodePort,
		},
	}, nil

}

func newStatefulSetForCR(cr *wnohangv1alpha1.Zookeeper, zooIDs string) (*appsv1.StatefulSet, error) {
	labels := map[string]string{
		"app": cr.Name,
	}

	zooContainer, err := getZookeeperContainer(zooIDs, cr.Name)
	if err != nil {
		return nil, fmt.Errorf("Failed to get Zookeeper Container")
	}
	standardStorageClass := "standard"
	zStorage, err := resource.ParseQuantity("50Mi")
	if err != nil {
		return nil, fmt.Errorf("Failed to obtain storage spec")
	}

	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			ServiceName:         cr.Name + "-serv",
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Replicas:            &cr.Spec.Nodes,
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "PersistentVolumeClaim",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      cr.Name,
						Namespace: cr.Namespace,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						// VolumeName: "datadir",
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Selector: &metav1.LabelSelector{
							MatchLabels: labels,
						},
						StorageClassName: &standardStorageClass,
						Resources: corev1.ResourceRequirements{
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceStorage: zStorage,
							},
						},
					},
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{zooContainer},
				},
			},
		},
	}, nil

}

func getZooIds(numNodes int32) string {
	zooIDs := "1"

	for nd := 2; nd <= int(numNodes); nd++ {
		zooIDs += "," + strconv.Itoa(nd)
	}

	return zooIDs
}

func getZookeeperContainer(zooIDs string, volume string) (corev1.Container, error) {

	return corev1.Container{
		Name:            "zookeeper",
		Image:           "ronin/zookeeper-k8s",
		ImagePullPolicy: corev1.PullIfNotPresent,
		Ports: []corev1.ContainerPort{
			{
				Name:          "client",
				ContainerPort: 2181,
			},
			{
				Name:          "leader-election",
				ContainerPort: 2888,
			},
			{
				Name:          "cluster-comms",
				ContainerPort: 3888,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      volume,
				MountPath: "/data",
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "ZOO_IDS",
				Value: zooIDs,
			},
		},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(2181),
				},
			},
		},

		LivenessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/sh", "-c", "echo mntr | nc localhost 2181 | grep -q zk_server_state"},
				},
			},
		},
	}, nil
}
