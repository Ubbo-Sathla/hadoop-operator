/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"math/rand"
	"strconv"

	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apachev1 "github.com/Ubbo-Sathla/hadoop-operator/api/v1"
)

// HadoopReconciler reconciles a Hadoop object
type HadoopReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apache.learn.org,resources=hadoops,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apache.learn.org,resources=hadoops/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apache.learn.org,resources=hadoops/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Hadoop object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *HadoopReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	reqLogger := log.Log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling HadoopService")

	// Fetch the Hadoop instance
	instance := &apachev1.Hadoop{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("HadoopService resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get HadoopService.")
		return ctrl.Result{}, err
	}

	slaveFound := &appsv1.StatefulSet{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-hadoop-slave", Namespace: instance.Namespace}, slaveFound)
	if err != nil && errors.IsNotFound(err) {
		// 1. Define a new StatefulSet for slaves
		slaveStatefulset := r.statefulSetForSlave(instance)
		reqLogger.Info("Creating a new statefulset and services.", "StatefulSet.Namespace", slaveStatefulset.Namespace, "StatefulSet.Name", slaveStatefulset.Name)
		err = r.Client.Create(context.TODO(), slaveStatefulset)
		if err != nil {
			reqLogger.Error(err, "Failed to create new StatefulSet for slaves.", "StatefulSet.Namespace", slaveStatefulset.Namespace, "StatefulSet.Name", slaveStatefulset.Name)
			return ctrl.Result{}, err
		}

		// 2. Define a new Headless Service for slaves
		slaveService := r.serviceForSlave(instance)
		err = r.Client.Create(context.TODO(), slaveService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service for slaves.", "Service.Namespace", slaveService.Namespace, "Service.Name", slaveService.Name)
			return ctrl.Result{}, err
		}

		// 3. Define a new Secret for master (SSH access password)
		passwordGen := randomString(4)
		reqLogger.Info("Generated password is " + passwordGen)
		masterSecret := r.secretForSlave(instance, passwordGen)
		err = r.Client.Create(context.TODO(), masterSecret)
		if err != nil {
			reqLogger.Error(err, "Failed to create new secret for master.", "Secret.Namespace", masterSecret.Namespace, "Secret.Name", masterSecret.Name)
			return ctrl.Result{}, err
		}

		// 4. Define a new StatefulSet for master
		masterStatefulset := r.statefulSetForMaster(instance)
		err = r.Client.Create(context.TODO(), masterStatefulset)
		if err != nil {
			reqLogger.Error(err, "Failed to create new StatefulSet for master.", "StatefulSet.Namespace", masterStatefulset.Namespace, "StatefulSet.Name", masterStatefulset.Name)
			return ctrl.Result{}, err
		}

		// 5. Define a new Headless Service for master
		masterService := r.serviceForMaster(instance)
		err = r.Client.Create(context.TODO(), masterService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service for master.", "Service.Namespace", masterService.Namespace, "Service.Name", slaveService.Name)
			return ctrl.Result{}, err
		}

		// 6. Define a new external Service for master
		masterExternalService := r.externalServiceForMaster(instance)
		err = r.Client.Create(context.TODO(), masterExternalService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service for master.", "Service.Namespace", masterExternalService.Namespace, "Service.Name", masterExternalService.Name)
			return ctrl.Result{}, err
		}

		// All resources are created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get StatefulSet for slaves.")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *HadoopReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apachev1.Hadoop{}).
		Complete(r)
}

//+kubebuilder:rbac:groups="",resources=pods;services;endpoints;persistentvolumeclaims;events;configmaps;secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;replicasets;statefulsets,verbs=get;list;watch;create;update;patch;delete

// Define by coder
func (r *HadoopReconciler) externalServiceForMaster(h *apachev1.Hadoop) *corev1.Service {
	masterServiceName := h.Name + "-hadoop-master-svc-external"
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      masterServiceName,
			Namespace: h.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{Name: "ssh", Port: 22},
				{Name: "yarn", Port: 8088},
				{Name: "metadata", Port: h.Spec.HealthPort},
				{Name: "dashboard", Port: 9870},
				{Name: "yarnproxy", Port: 80},
			},
			Selector: labelsForHadoopMaster(h.Name),
		},
	}

	// Register HadoopService instance as the owner and controller of slaves service
	controllerutil.SetControllerReference(h, service, r.Scheme)
	return service
}

func (r *HadoopReconciler) serviceForMaster(h *apachev1.Hadoop) *corev1.Service {
	masterServiceName := h.Name + "-hadoop-master-svc"
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      masterServiceName,
			Namespace: h.Namespace,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Selector:  labelsForHadoopMaster(h.Name),
		},
	}

	// Register HadoopService instance as the owner and controller of slaves service
	controllerutil.SetControllerReference(h, service, r.Scheme)
	return service
}

func (r *HadoopReconciler) secretForSlave(h *apachev1.Hadoop, password string) *corev1.Secret {
	masterSecretName := h.Name + "-hadoop-master-secret"
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      masterSecretName,
			Namespace: h.Namespace,
		},
		Data: map[string][]byte{
			"password": []byte(password),
		},
		Type: "Opaque",
	}

	// Register HadoopService instance as the owner and controller of slaves service
	controllerutil.SetControllerReference(h, secret, r.Scheme)
	return secret
}
func (r *HadoopReconciler) statefulSetForMaster(h *apachev1.Hadoop) *appsv1.StatefulSet {
	ls := labelsForHadoopMaster(h.Name)
	var replicas int32 = 1
	masterSecretName := h.Name + "-hadoop-master-secret"

	slaveServiceName := h.Name + "-hadoop-slave-svc"
	slaveStatefulsetName := h.Name + "-hadoop-slave"

	masterName := h.Name + "-hadoop-master"
	masterServiceName := h.Name + "-hadoop-master-svc"
	masterEndpoint := h.Name + "-hadoop-master-0." + masterServiceName + "." + h.Namespace + ".svc.cluster.local"

	slaveCount := int(h.Spec.ClusterSize - 1)

	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      masterName,
			Namespace: h.Namespace,
			Labels:    ls,
		},
		Spec: appsv1.StatefulSetSpec{
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Replicas:            &replicas,
			ServiceName:         masterServiceName,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{{
						Name: masterSecretName,
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{SecretName: masterSecretName},
						},
					}},
					Containers: []corev1.Container{{
						ImagePullPolicy: corev1.PullAlways,
						Image:           h.Spec.Image,
						Name:            "hadoop-master",
						Command:         []string{"/master-entrypoint.sh"},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(int(h.Spec.HealthPort))}, // hadoop metadata port
							},
							InitialDelaySeconds: 5,
							PeriodSeconds:       2,
						},
						Ports: []corev1.ContainerPort{
							{ContainerPort: 22, Name: "ssh"},
							{ContainerPort: h.Spec.HealthPort, Name: "metadata"},
							{ContainerPort: 8088, Name: "yarn"},
							{ContainerPort: 9870, Name: "dashboard"},
							{ContainerPort: 80, Name: "yarnproxy"},
						},
						Env: []corev1.EnvVar{
							{Name: "MASTER_ENDPOINT", Value: masterEndpoint},
							{Name: "SLAVES_SVC_NAME", Value: slaveServiceName},
							{Name: "SLAVES_SS_NAME", Value: slaveStatefulsetName},
							{Name: "SLAVES_COUNT", Value: strconv.Itoa(slaveCount)},
							{Name: "NAMESPACE", Value: h.Namespace},
						},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      masterSecretName,
							MountPath: "/etc/rootpwd",
							ReadOnly:  true,
						}},
					}},
				},
			},
		},
	}

	// Register HadoopService instance as the owner and controller of slaves StatefulSet
	controllerutil.SetControllerReference(h, statefulset, r.Scheme)
	return statefulset
}
func (r *HadoopReconciler) serviceForSlave(h *apachev1.Hadoop) *corev1.Service {
	slaveServiceName := h.Name + "-hadoop-slave-svc"
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      slaveServiceName,
			Namespace: h.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name: "ssh",
				Port: 22,
			}},
			ClusterIP: "None",
			Selector:  labelsForHadoopSlave(h.Name),
		},
	}

	// Register HadoopService instance as the owner and controller of slaves service
	controllerutil.SetControllerReference(h, service, r.Scheme)
	return service
}

func (r *HadoopReconciler) statefulSetForSlave(h *apachev1.Hadoop) *appsv1.StatefulSet {
	ls := labelsForHadoopSlave(h.Name)
	replicas := h.Spec.ClusterSize

	slaveServiceName := h.Name + "-hadoop-slave-svc"
	slaveStatefulsetName := h.Name + "-hadoop-slave"

	masterServiceName := h.Name + "-hadoop-master-svc"
	masterEndpoint := h.Name + "-hadoop-master-0." + masterServiceName + "." + h.Namespace + ".svc.cluster.local"
	// masterEndpoint example : hadoop-master-0.hadoop-master-svc.default.svc.cluster.local

	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      slaveStatefulsetName,
			Namespace: h.Namespace,
			Labels:    ls,
		},
		Spec: appsv1.StatefulSetSpec{
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Replicas:            &replicas,
			ServiceName:         slaveServiceName,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						ImagePullPolicy: corev1.PullAlways,
						Image:           h.Spec.Image,
						Name:            "hadoop-slave",
						Command:         []string{"/slave-entrypoint.sh"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 22,
							Name:          "ssh",
						}},
						Env: []corev1.EnvVar{{
							Name:  "MASTER_ENDPOINT",
							Value: masterEndpoint,
						}},
					}},
				},
			},
		},
	}

	// Register HadoopService instance as the owner and controller of slaves StatefulSet
	controllerutil.SetControllerReference(h, statefulset, r.Scheme)
	return statefulset
}

// labelsForHadoopService returns the labels for selecting the resources
// belonging to the given HadoopService custom resource.
func labelsForHadoopSlave(name string) map[string]string {
	return map[string]string{"app": "hadoopservice", "hadoop_cr": name, "role": "slave"}
}
func labelsForHadoopMaster(name string) map[string]string {
	return map[string]string{"app": "hadoopservice", "hadoop_cr": name, "role": "master"}
}

func randomString(len int) string {
	bytes := make([]byte, len)
	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}
	return string(bytes)
}
