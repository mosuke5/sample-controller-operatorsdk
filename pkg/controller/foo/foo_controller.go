package foo

import (
	"context"

	samplecontrollerv1alpha1 "github.com/mosuke5/sample-controller-operatorsdk/pkg/apis/samplecontroller/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = logf.Log.WithName("controller_foo")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Foo Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileFoo{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("foo-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Foo
	err = c.Watch(&source.Kind{Type: &samplecontrollerv1alpha1.Foo{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Foo
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &samplecontrollerv1alpha1.Foo{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileFoo implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileFoo{}

// ReconcileFoo reconciles a Foo object
type ReconcileFoo struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Foo object and makes changes based on the state read
// and what is in the Foo.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileFoo) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Foo")
	ctx := context.Background()
	// Fetch the Foo instance
	instance := &samplecontrollerv1alpha1.Foo{}
	err := r.client.Get(ctx, request.NamespacedName, instance)
	//err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Foo not found. Ignore not found")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get foo")
		return reconcile.Result{}, err
	}

	if err := r.cleanupOwnedResources(ctx, instance); err != nil {
		reqLogger.Error(err, "failed to clean up old Deployment resources for this Foo")
		return reconcile.Result{}, err
	}

	// get deploymentName from foo.Spec
	deploymentName := instance.Spec.DeploymentName

	// define deployment template using deploymentName
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: request.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(instance, samplecontrollerv1alpha1.SchemeGroupVersion.WithKind("Foo")),
			},
		},
	}

	// Create or Update deployment object
	reqLogger.Info("create or update deployment")
	if _, err := ctrl.CreateOrUpdate(ctx, r.client, deployment, func() error {

		// set the replicas from foo.Spec
		replicas := int32(1)
		if instance.Spec.Replicas != nil {
			replicas = *instance.Spec.Replicas
		}
		deployment.Spec.Replicas = &replicas

		// set a label for our deployment
		labels := map[string]string{
			"app":        "nginx",
			"controller": instance.Name,
		}

		// set labels to objectmeta.labels
		if deployment.ObjectMeta.Labels == nil {
			deployment.ObjectMeta.Labels = labels
		}

		// set labels to spec.selector for our deployment
		if deployment.Spec.Selector == nil {
			deployment.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
		}

		// set labels to template.objectMeta for our deployment
		if deployment.Spec.Template.ObjectMeta.Labels == nil {
			deployment.Spec.Template.ObjectMeta.Labels = labels
		}

		// set a container for our deployment
		containers := []corev1.Container{
			{
				Name:  "nginx",
				Image: "nginx:latest",
			},
		}

		// set containers to template.spec.containers for our deployment
		if deployment.Spec.Template.Spec.Containers == nil {
			deployment.Spec.Template.Spec.Containers = containers
		}

		// set the owner so that garbage collection can kicks in
		//if err := ctrl.SetControllerReference(&instance, deployment, r.Scheme); err != nil {
		//	log.Error(err, "unable to set ownerReference from Foo to Deployment")
		//	return err
		//}

		// end of ctrl.CreateOrUpdate
		return nil

	}); err != nil {
		// error handling of ctrl.CreateOrUpdate
		log.Error(err, "unable to ensure deployment is correct")
		return ctrl.Result{}, err
	} 
	
	// compare foo.status.availableReplicas and deployment.status.availableReplicas
	if instance.Status.AvailableReplicas != deployment.Status.AvailableReplicas {

		reqLogger.Info("updating Foo status")
		instance.Status.AvailableReplicas = deployment.Status.AvailableReplicas
		// Update foo spec
		if err := r.client.Update(ctx, instance); err != nil {
			reqLogger.Error(err, "failed to update Foo status")
			// Error updating the object - requeue the request.
			return reconcile.Result{}, err
		}

		reqLogger.Info("updated Foo status", "foo.status.availableReplicas", instance.Status.AvailableReplicas)
	}

	return reconcile.Result{}, nil
}

// cleanupOwnedResources will Delete any existing Deployment resources that
// were created for the given Foo that no longer match the
// foo.spec.deploymentName field.
func (r *ReconcileFoo) cleanupOwnedResources(ctx context.Context, foo *samplecontrollerv1alpha1.Foo) error {
	reqLogger := log.WithValues("Request.Namespace", foo.Namespace, "Request.Name", foo.Name)
	reqLogger.Info("finding existing Deployments for Foo resource")

	// List all deployment resources owned by this Foo
	deployments := &appsv1.DeploymentList{}
	labelSelector := labels.SelectorFromSet(labelsForFoo(foo.Name))
	listOps := &client.ListOptions{
		Namespace:     foo.Namespace,
		LabelSelector: labelSelector,
	}
	if err := r.client.List(ctx, deployments, listOps); err != nil{
	//if err := r.client.List(ctx, listOps, deployments); err != nil{
		reqLogger.Error(err,"faild to get list of deployments")
		return err
	}

	// Delete deployment if the deployment name doesn't match foo.spec.deploymentName
	for _, deployment := range deployments.Items {
		if deployment.Name == foo.Spec.DeploymentName {
			// If this deployment's name matches the one on the Foo resource
			// then do not delete it.
			continue
		}

		// Delete old deployment object which doesn't match foo.spec.deploymentName
		if err := r.client.Delete(ctx, &deployment); err != nil {
			reqLogger.Error(err, "failed to delete Deployment resource")
			return err
		}

		reqLogger.Info("deleted old Deployment resource for Foo", "deploymentName", deployment.Name)
	}

	return nil
}

// labelsForFoo returns the labels for selecting the resources
// belonging to the given foo CR name.
func labelsForFoo(name string) map[string]string {
	return map[string]string{"app": "nginx", "controller": name}
}

// newDeployment creates a new Deployment for a Foo resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Foo resource that 'owns' it.
//func newDeployment(foo *samplecontrollerv1alpha1.Foo) *appsv1.Deployment {
//	labels := map[string]string{
//		"app":        "nginx",
//		"controller": foo.Name,
//	}
//	return &appsv1.Deployment{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      foo.Spec.DeploymentName,
//			Namespace: foo.Namespace,
//			Labels:    labels,
//			OwnerReferences: []metav1.OwnerReference{
//				*metav1.NewControllerRef(foo, samplecontrollerv1alpha1.SchemeGroupVersion.WithKind("Foo")),
//			},
//		},
//		Spec: appsv1.DeploymentSpec{
//			Replicas: foo.Spec.Replicas,
//			Selector: &metav1.LabelSelector{
//				MatchLabels: labels,
//			},
//			Template: corev1.PodTemplateSpec{
//				ObjectMeta: metav1.ObjectMeta{
//					Labels: labels,
//				},
//				Spec: corev1.PodSpec{
//					Containers: []corev1.Container{
//						{
//							Name:  "nginx",
//							Image: "nginx:latest",
//						},
//					},
//				},
//			},
//		},
//	}
//}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *samplecontrollerv1alpha1.Foo) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
