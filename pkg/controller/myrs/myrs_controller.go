package myrs

import (
	"context"
	"reflect"

	myv1alpha1 "myrs-operator/pkg/apis/my/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_myrs")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new MyRS Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMyRS{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("myrs-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource MyRS
	err = c.Watch(&source.Kind{Type: &myv1alpha1.MyRS{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner MyRS
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &myv1alpha1.MyRS{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileMyRS implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMyRS{}

// ReconcileMyRS reconciles a MyRS object
type ReconcileMyRS struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a MyRS object and makes changes based on the state read
// and what is in the MyRS.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMyRS) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MyRS")

	// Fetch the MyRS instance
	MyRS := &myv1alpha1.MyRS{}
	err := r.client.Get(context.TODO(), request.NamespacedName, MyRS)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	lbls := labels.Set{
		"app":     MyRS.Name,
	}
	existingPods := &corev1.PodList{}
	err = r.client.List(context.TODO(),
		existingPods,
		&client.ListOptions{
			Namespace:     request.Namespace,
			LabelSelector: labels.SelectorFromSet(lbls),
		})
	if err != nil {
		reqLogger.Error(err, "failed to list existing pods in the myRS")
		return reconcile.Result{}, err
	}

	existingPodNames := []string{}

	for _, pod := range existingPods.Items {
		if pod.GetObjectMeta().GetDeletionTimestamp() != nil {
			continue
		}
		if pod.Status.Phase == corev1.PodPending || pod.Status.Phase == corev1.PodRunning {
			existingPodNames = append(existingPodNames, pod.GetObjectMeta().GetName())
		}
	}

	reqLogger.Info("Checking myRS", "expected replicas", MyRS.Spec.Replicas, "Pod.Names", existingPodNames)
	status := myv1alpha1.MyRSStatus{
		Replicas: int32(len(existingPodNames)),
		PodNames: existingPodNames,
	}
	if !reflect.DeepEqual(MyRS.Status, status) {
		MyRS.Status = status
		err := r.client.Status().Update(context.TODO(), MyRS)
		if err != nil {
			reqLogger.Error(err, "failed to update the myRS")
			return reconcile.Result{}, err
		}
	}
	if int32(len(existingPodNames)) > *MyRS.Spec.Replicas {
		reqLogger.Info("Deleting a pod in the myRS", "expected replicas", MyRS.Spec.Replicas, "Pod.Names", existingPodNames)
		pod := existingPods.Items[0]
		err = r.client.Delete(context.TODO(), &pod)
		if err != nil {
			reqLogger.Error(err, "failed to delete a pod")
			return reconcile.Result{}, err
		}
	}
	if int32(len(existingPodNames)) < *MyRS.Spec.Replicas {
		reqLogger.Info("Adding a pod in the podset", "expected replicas", MyRS.Spec.Replicas, "Pod.Names", existingPodNames)
		pod := newPodForCR(MyRS)
		if err := controllerutil.SetControllerReference(MyRS, pod, r.scheme); err != nil {
			reqLogger.Error(err, "unable to set owner reference on new pod")
			return reconcile.Result{}, err
		}
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			reqLogger.Error(err, "failed to create a pod")
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{Requeue: true}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *myv1alpha1.MyRS) *corev1.Pod {
	labels := map[string]string{
		"app": 	   cr.Name,
	}
	for SelectorKey, SelectorValue := range cr.Spec.Selector.MatchLabels {
		labels[SelectorKey] = SelectorValue
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: cr.Name + "-pod-",
			Namespace: 	  cr.Namespace,
			Labels:    	  labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    cr.Spec.Template.Spec.Containers[0].Name,
					Image:   cr.Spec.Template.Spec.Containers[0].Image,
				},
			},
		},
	}
}
