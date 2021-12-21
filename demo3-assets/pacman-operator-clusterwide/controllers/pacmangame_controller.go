/*


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
	"reflect"

	"github.com/go-logr/logr"
	appsv1beta1 "github.com/mvazquezc/pacman-operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// PacmanGameReconciler reconciles a PacmanGame object
type PacmanGameReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Finalizer for our objects
const PacmanGameFinalizer = "finalizer.pacmangame.apps.rha.lab"

// +kubebuilder:rbac:groups=apps.rha.lab,resources=pacmangames,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.rha.lab,resources=pacmangames/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.rha.lab,resources=pacmangames/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete

func (r *PacmanGameReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	// Fetch the PacmanGame instance
	instance := &appsv1beta1.PacmanGame{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("PacmanGame resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get PacmanGame")
		return ctrl.Result{}, err
	}

	// Check if the CR is marked to be deleted
	isInstanceMarkedToBeDeleted := instance.GetDeletionTimestamp() != nil
	if isInstanceMarkedToBeDeleted {
		log.Info("Instance marked for deletion, running finalizers")
		if contains(instance.GetFinalizers(), PacmanGameFinalizer) {
			// Run the finalizer logic
			err := r.finalizePacmanGame(log, instance)
			if err != nil {
				// Don't remove the finalizer if we failed to finalize the object
				return ctrl.Result{}, err
			}
			log.Info("Instance finalizers completed")
			// Remove finalizer once the finalizer logic has run
			controllerutil.RemoveFinalizer(instance, PacmanGameFinalizer)
			err = r.Update(ctx, instance)
			if err != nil {
				// If the object update fails, requeue
				return ctrl.Result{}, err
			}
		}
		log.Info("Instance can be deleted now")
		return ctrl.Result{}, nil
	}

	// Add Finalizers to the CR
	if !contains(instance.GetFinalizers(), PacmanGameFinalizer) {
		if err := r.addFinalizer(log, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Reconcile Mongo Deployment object
	result, err := r.reconcileMongoDeployment(instance, log)
	if err != nil {
		return result, err
	}
	// Reconcile Mongo Service object
	result, err = r.reconcileMongoService(instance, log)
	if err != nil {
		return result, err
	}

	// Reconcile Pacman Deployment object
	result, err = r.reconcilePacmanDeployment(instance, log)
	if err != nil {
		return result, err
	}
	// Reconcile Pacman Service object
	result, err = r.reconcilePacmanService(instance, log)
	if err != nil {
		return result, err
	}

	// Reconcile Pacman ServiceAccount object
	result, err = r.reconcilePacmanServiceAccount(instance, log)
	if err != nil {
		return result, err
	}
	// Reconcile Pacman ClusterRole object
	result, err = r.reconcilePacmanClusterRole(instance, log)
	if err != nil {
		return result, err
	}

	// Reconcile Pacman ClusterRoleBinding object
	result, err = r.reconcilePacmanClusterRoleBinding(instance, log)
	if err != nil {
		return result, err
	}

	// The CR status is updated in the Deployment reconcile method

	return ctrl.Result{}, nil
}

func (r *PacmanGameReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1beta1.PacmanGame{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.ClusterRole{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Complete(r)
}

func (r *PacmanGameReconciler) reconcilePacmanDeployment(cr *appsv1beta1.PacmanGame, log logr.Logger) (ctrl.Result, error) {
	// Define a new Deployment object
	deployment := newPacmanDeploymentForCR(cr)

	// Set PacmanGame instance as the owner and controller of the Deployment
	if err := ctrl.SetControllerReference(cr, deployment, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Check if this Deployment already exists
	deploymentFound := &appsv1.Deployment{}
	err := r.Get(context.Background(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, deploymentFound)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.Create(context.Background(), deployment)
		if err != nil {
			return ctrl.Result{}, err
		}
		// Requeue the object to update its status
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	} else {
		// Deployment already exists
		log.Info("Deployment already exists", "Deployment.Namespace", deploymentFound.Namespace, "Deployment.Name", deploymentFound.Name)
	}

	// Ensure deployment replicas match the desired state
	if !reflect.DeepEqual(deploymentFound.Spec.Replicas, deployment.Spec.Replicas) {
		log.Info("Current deployment replicas do not match PacmanGame configured Replicas")
		// Update the replicas
		err = r.Update(context.Background(), deployment)
		if err != nil {
			log.Error(err, "Failed to update Deployment.", "Deployment.Namespace", deploymentFound.Namespace, "Deployment.Name", deploymentFound.Name)
			return ctrl.Result{}, err
		}
	}
	// Ensure deployment container image match the desired state, returns true if deployment needs to be updated
	if checkDeploymentImage(deploymentFound, deployment) {
		log.Info("Current deployment image version do not match PacmanGame configured version")
		// Update the image
		err = r.Update(context.Background(), deployment)
		if err != nil {
			log.Error(err, "Failed to update Deployment.", "Deployment.Namespace", deploymentFound.Namespace, "Deployment.Name", deploymentFound.Name)
			return ctrl.Result{}, err
		}
	}

	// Check if the deployment is ready
	deploymentReady := isDeploymentReady(deploymentFound)

	// Create list options for listing deployment pods
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(deploymentFound.Namespace),
		client.MatchingLabels(deploymentFound.Labels),
	}
	// List the pods for this PacmanGame deployment
	err = r.List(context.Background(), podList, listOpts...)
	if err != nil {
		log.Error(err, "Failed to list Pods.", "Deployment.Namespace", deploymentFound.Namespace, "Deployment.Name", deploymentFound.Name)
		return ctrl.Result{}, err
	}
	// Get running Pods from listing above (if any)
	podNames := getRunningPodNames(podList.Items)
	if deploymentReady {
		// Update the status to ready
		cr.Status.AppPods = podNames
		meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{Type: appsv1beta1.ConditionTypePacmanGameDeploymentNotReady, Status: metav1.ConditionFalse, Reason: appsv1beta1.ConditionTypePacmanGameDeploymentNotReady})
		meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{Type: appsv1beta1.ConditionTypeReady, Status: metav1.ConditionTrue, Reason: appsv1beta1.ConditionTypeReady})
	} else {
		// Update the status to not ready
		cr.Status.AppPods = podNames
		meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{Type: appsv1beta1.ConditionTypePacmanGameDeploymentNotReady, Status: metav1.ConditionTrue, Reason: appsv1beta1.ConditionTypePacmanGameDeploymentNotReady})
		meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{Type: appsv1beta1.ConditionTypeReady, Status: metav1.ConditionFalse, Reason: appsv1beta1.ConditionTypeReady})
	}
	// Reconcile the new status for the instance
	cr, err = r.updatePacmanGameStatus(cr, log)
	if err != nil {
		log.Error(err, "Failed to update PacmanGame Status.")
		return ctrl.Result{}, err
	}
	// Deployment reconcile finished
	return ctrl.Result{}, nil
}

func (r *PacmanGameReconciler) reconcilePacmanServiceAccount(cr *appsv1beta1.PacmanGame, log logr.Logger) (ctrl.Result, error) {
	// Define a new ServiceAccount object
	serviceAccount := newPacmanServiceAccountForCR(cr)

	// Set PacmanGame instance as the owner and controller of the ServiceAccount
	if err := controllerutil.SetControllerReference(cr, serviceAccount, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Check if this ServiceAccount already exists
	serviceAccountFound := &corev1.ServiceAccount{}
	err := r.Get(context.Background(), types.NamespacedName{Name: serviceAccount.Name, Namespace: serviceAccount.Namespace}, serviceAccountFound)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new ServiceAccount", "ServiceAccount.Namespace", serviceAccount.Namespace, "ServiceAccount.Name", serviceAccount.Name)
		err = r.Create(context.Background(), serviceAccount)
		if err != nil {
			return ctrl.Result{}, err
		}
		// ServiceAccount created successfully - don't requeue
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	} else {
		// ServiceAccount already exists
		log.Info("ServiceAccount already exists", "ServiceAccount.Namespace", serviceAccount.Namespace, "ServiceAccount.Name", serviceAccount.Name)
	}
	// ServiceAccount reconcile finished
	return ctrl.Result{}, nil
}

func (r *PacmanGameReconciler) reconcilePacmanClusterRole(cr *appsv1beta1.PacmanGame, log logr.Logger) (ctrl.Result, error) {
	// Define a new ClusterRole object
	clusterRole := newPacmanClusterRoleForCR(cr)

	// Set PacmanGame instance as the owner and controller of the ClusterRole
	//if err := controllerutil.SetControllerReference(cr, clusterRole, r.Scheme); err != nil {
	//	return ctrl.Result{}, err
	//}

	// Check if this ClusterRole already exists
	clusterRoleFound := &rbacv1.ClusterRole{}
	err := r.Get(context.Background(), types.NamespacedName{Name: clusterRole.Name}, clusterRoleFound)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new clusterRole", "clusterRole.Name", clusterRole.Name)
		err = r.Create(context.Background(), clusterRole)
		if err != nil {
			return ctrl.Result{}, err
		}
		// ClusterRole created successfully - don't requeue
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	} else {
		// ClusterRole already exists
		log.Info("clusterRole already exists", "clusterRole.Name", clusterRole.Name)
	}
	// ClusterRole reconcile finished
	return ctrl.Result{}, nil
}

func (r *PacmanGameReconciler) reconcilePacmanClusterRoleBinding(cr *appsv1beta1.PacmanGame, log logr.Logger) (ctrl.Result, error) {
	// Define a new ClusterRoleBinding object
	clusterRoleBinding := newPacmanClusterRoleBindingForCR(cr)

	// Set PacmanGame instance as the owner and controller of the ClusterRoleBinding
	//if err := controllerutil.SetControllerReference(cr, clusterRoleBinding, r.Scheme); err != nil {
	//	return ctrl.Result{}, err
	//}

	// Check if this ClusterRoleBinding already exists
	clusterRoleBindingFound := &rbacv1.ClusterRoleBinding{}
	err := r.Get(context.Background(), types.NamespacedName{Name: clusterRoleBinding.Name}, clusterRoleBindingFound)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new clusterRole", "clusterRoleBinding.Name", clusterRoleBinding.Name)
		err = r.Create(context.Background(), clusterRoleBinding)
		if err != nil {
			return ctrl.Result{}, err
		}
		// ClusterRoleBinding created successfully - don't requeue
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	} else {
		// ClusterRoleBinding already exists
		log.Info("clusterRoleBinding already exists", "clusterRoleBinding.Name", clusterRoleBinding.Name)
	}
	// ClusterRoleBinding reconcile finished
	return ctrl.Result{}, nil
}

func (r *PacmanGameReconciler) reconcilePacmanService(cr *appsv1beta1.PacmanGame, log logr.Logger) (ctrl.Result, error) {
	// Define a new Service object
	service := newPacmanServiceForCR(cr)

	// Set PacmanGame instance as the owner and controller of the Service
	if err := controllerutil.SetControllerReference(cr, service, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Check if this Service already exists
	serviceFound := &corev1.Service{}
	err := r.Get(context.Background(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, serviceFound)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.Create(context.Background(), service)
		if err != nil {
			return ctrl.Result{}, err
		}
		// Service created successfully - don't requeue
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	} else {
		// Service already exists
		log.Info("Service already exists", "Service.Namespace", serviceFound.Namespace, "Service.Name", serviceFound.Name)
	}
	// Service reconcile finished
	return ctrl.Result{}, nil
}

func (r *PacmanGameReconciler) reconcileMongoDeployment(cr *appsv1beta1.PacmanGame, log logr.Logger) (ctrl.Result, error) {
	// Define a new Deployment object
	deployment := newMongoDeploymentForCR(cr)

	// Set PacmanGame instance as the owner and controller of the Deployment
	if err := ctrl.SetControllerReference(cr, deployment, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Check if this Deployment already exists
	deploymentFound := &appsv1.Deployment{}
	err := r.Get(context.Background(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, deploymentFound)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.Create(context.Background(), deployment)
		if err != nil {
			return ctrl.Result{}, err
		}
		// Requeue the object to update its status
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	} else {
		// Deployment already exists
		log.Info("Deployment already exists", "Deployment.Namespace", deploymentFound.Namespace, "Deployment.Name", deploymentFound.Name)
	}

	// Ensure deployment replicas match the desired state
	if !reflect.DeepEqual(deploymentFound.Spec.Replicas, deployment.Spec.Replicas) {
		log.Info("Current deployment replicas do not match PacmanGame configured Replicas")
		// Update the replicas
		err = r.Update(context.Background(), deployment)
		if err != nil {
			log.Error(err, "Failed to update Deployment.", "Deployment.Namespace", deploymentFound.Namespace, "Deployment.Name", deploymentFound.Name)
			return ctrl.Result{}, err
		}
	}
	// Ensure deployment container image match the desired state, returns true if deployment needs to be updated
	if checkDeploymentImage(deploymentFound, deployment) {
		log.Info("Current deployment image version do not match PacmanGame configured version")
		// Update the image
		err = r.Update(context.Background(), deployment)
		if err != nil {
			log.Error(err, "Failed to update Deployment.", "Deployment.Namespace", deploymentFound.Namespace, "Deployment.Name", deploymentFound.Name)
			return ctrl.Result{}, err
		}
	}

	// Reconcile the new status for the instance
	cr, err = r.updatePacmanGameStatus(cr, log)
	if err != nil {
		log.Error(err, "Failed to update PacmanGame Status.")
		return ctrl.Result{}, err
	}
	// Deployment reconcile finished
	return ctrl.Result{}, nil
}

func (r *PacmanGameReconciler) reconcileMongoService(cr *appsv1beta1.PacmanGame, log logr.Logger) (ctrl.Result, error) {
	// Define a new Service object
	service := newMongoServiceForCR(cr)

	// Set PacmanGame instance as the owner and controller of the Service
	if err := controllerutil.SetControllerReference(cr, service, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Check if this Service already exists
	serviceFound := &corev1.Service{}
	err := r.Get(context.Background(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, serviceFound)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.Create(context.Background(), service)
		if err != nil {
			return ctrl.Result{}, err
		}
		// Service created successfully - don't requeue
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	} else {
		// Service already exists
		log.Info("Service already exists", "Service.Namespace", serviceFound.Namespace, "Service.Name", serviceFound.Name)
	}
	// Service reconcile finished
	return ctrl.Result{}, nil
}

// updatePacmanGameStatus updates the Status of a given CR
func (r *PacmanGameReconciler) updatePacmanGameStatus(cr *appsv1beta1.PacmanGame, log logr.Logger) (*appsv1beta1.PacmanGame, error) {
	pacmanGame := &appsv1beta1.PacmanGame{}
	err := r.Get(context.Background(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, pacmanGame)
	if err != nil {
		return pacmanGame, err
	}

	if !reflect.DeepEqual(cr.Status, pacmanGame.Status) {
		log.Info("Updating PacmanGame Status.")
		// We need to update the status
		err = r.Status().Update(context.Background(), cr)
		if err != nil {
			return cr, err
		}
		updatedPacmanGame := &appsv1beta1.PacmanGame{}
		err = r.Get(context.Background(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, updatedPacmanGame)
		if err != nil {
			return cr, err
		}
		cr = updatedPacmanGame.DeepCopy()
	}
	return cr, nil

}

// addFinalizer adds a given finalizer to a given CR
func (r *PacmanGameReconciler) addFinalizer(log logr.Logger, cr *appsv1beta1.PacmanGame) error {
	log.Info("Adding Finalizer for the PacmanGame")
	controllerutil.AddFinalizer(cr, PacmanGameFinalizer)

	// Update CR
	err := r.Update(context.Background(), cr)
	if err != nil {
		log.Error(err, "Failed to update PacmanGame with finalizer")
		return err
	}
	return nil
}

// finalizePacmanGame runs required tasks before deleting the objects owned by the CR
func (r *PacmanGameReconciler) finalizePacmanGame(log logr.Logger, cr *appsv1beta1.PacmanGame) error {
	// TODO(user): Add the cleanup steps that the operator
	// needs to do before the CR can be deleted. Examples
	// of finalizers include performing backups and deleting
	// resources that are not owned by this CR, like a PVC.
	log.Info("Successfully finalized PacmanGame")
	return nil
}

// Returns a new deployment without replicas configured
// replicas will be configured in the sync loop
func newMongoDeploymentForCR(cr *appsv1beta1.PacmanGame) *appsv1.Deployment {
	labels := map[string]string{
		"app": cr.Name,
	}
	// Replicas will be 1
	var replicas int32 = 1

	containerImage := "docker.io/library/mongo:latest"
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mongo-" + cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "mongodb-storage",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: "Memory",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Image: containerImage,
							Name:  "mongo",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "mongodb-storage",
									MountPath: "/data/db",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "MONGO_INITDB_ROOT_USERNAME",
									Value: "admin",
								},
								{
									Name:  "MONGO_INITDB_ROOT_PASSWORD",
									Value: "admin",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 27017,
									Name:          "mongo",
								},
							},
						},
					},
				},
			},
		},
	}
}

// Returns a new mongo service
func newMongoServiceForCR(cr *appsv1beta1.PacmanGame) *corev1.Service {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mongo-" + cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: 27017,
				},
			},
		},
	}
}

// Returns a new deployment without replicas configured
// replicas will be configured in the sync loop
func newPacmanDeploymentForCR(cr *appsv1beta1.PacmanGame) *appsv1.Deployment {
	labels := map[string]string{
		"app": cr.Name,
	}
	replicas := cr.Spec.Replicas
	// Minimum replicas will be 1
	if replicas == 0 {
		replicas = 1
	}
	appVersion := "latest"
	if cr.Spec.AppVersion != "" {
		appVersion = cr.Spec.AppVersion
	}
	mongoService := "mongo-" + cr.Name + "." + cr.Namespace + ".svc.cluster.local"
	// TODO:Check if application version exists
	containerImage := "quay.io/ifont/pacman-nodejs-app:" + appVersion
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pacman-" + cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "pacman-" + cr.Name,
					Containers: []corev1.Container{
						{
							Image: containerImage,
							Name:  "pacman",
							Env: []corev1.EnvVar{
								{
									Name:  "MONGO_SERVICE_HOST",
									Value: mongoService,
								},
								{
									Name:  "MONGO_AUTH_USER",
									Value: "admin",
								},
								{
									Name:  "MONGO_AUTH_PWD",
									Value: "admin",
								},
								{
									Name:  "MONGO_DATABASE",
									Value: "test",
								},
								{
									Name:  "MY_MONGO_PORT",
									Value: "27017",
								},
								{
									Name:      "MY_NODE_NAME",
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
									Name:          "pacman",
								},
							},
						},
					},
				},
			},
		},
	}
}

// Returns a new pacman service
func newPacmanServiceForCR(cr *appsv1beta1.PacmanGame) *corev1.Service {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pacman-" + cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeLoadBalancer,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: 8080,
				},
			},
		},
	}
}

// Returns a new serviceaccount
func newPacmanServiceAccountForCR(cr *appsv1beta1.PacmanGame) *corev1.ServiceAccount {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pacman-" + cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
	}
}

// Returns a clusterRole
func newPacmanClusterRoleForCR(cr *appsv1beta1.PacmanGame) *rbacv1.ClusterRole {
	rules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"pods",
				"nodes",
			},
			Verbs: []string{
				"get",
				"watch",
				"list",
			},
		},
	}
	labels := map[string]string{
		"app": cr.Name,
	}
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "pacman-" + cr.Name,
			Labels: labels,
		},
		Rules: rules,
	}

}

// Returns a clusterRole
func newPacmanClusterRoleBindingForCR(cr *appsv1beta1.PacmanGame) *rbacv1.ClusterRoleBinding {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "pacman-" + cr.Name,
			Labels: labels,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      "pacman-" + cr.Name,
				Namespace: cr.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "pacman-" + cr.Name,
		},
	}
}

// isDeploymentReady returns a true bool if the deployment has all its pods ready
func isDeploymentReady(deployment *appsv1.Deployment) bool {
	configuredReplicas := deployment.Status.Replicas
	readyReplicas := deployment.Status.ReadyReplicas
	deploymentReady := false
	if configuredReplicas == readyReplicas {
		deploymentReady = true
	}
	return deploymentReady
}

// getRunningPodNames returns the pod names for the pods running in the array of pods passed in
func getRunningPodNames(pods []corev1.Pod) []string {
	// Create an empty []string, so if no podNames are returned, instead of nil we get an empty slice
	var podNames []string = make([]string, 0)
	for _, pod := range pods {
		if pod.GetObjectMeta().GetDeletionTimestamp() != nil {
			continue
		}
		if pod.Status.Phase == corev1.PodPending || pod.Status.Phase == corev1.PodRunning {
			podNames = append(podNames, pod.Name)
		}
	}
	return podNames
}

// checkDeploymentImage returns wether the deployment image is different or not
func checkDeploymentImage(current *appsv1.Deployment, desired *appsv1.Deployment) bool {
	for _, curr := range current.Spec.Template.Spec.Containers {
		for _, des := range desired.Spec.Template.Spec.Containers {
			// Only compare the images of containers with the same name
			if curr.Name == des.Name {
				if curr.Image != des.Image {
					return true
				}
			}
		}
	}
	return false
}

// contains returns true if a string is found on a slice
func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
