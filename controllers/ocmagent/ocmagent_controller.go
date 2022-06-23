package ocmagent

import (
	"context"

	"github.com/go-logr/logr"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	ocmagentv1alpha1 "github.com/openshift/ocm-agent-operator/api/v1alpha1"
	ctrlconst "github.com/openshift/ocm-agent-operator/pkg/consts/controller"
	"github.com/openshift/ocm-agent-operator/pkg/localmetrics"
	"github.com/openshift/ocm-agent-operator/pkg/ocmagenthandler"
)

// ReconcileOCMAgent reconciles a OCMAgent object
type ReconcileOCMAgent struct {
	Client client.Client
	Scheme *runtime.Scheme
	Ctx    context.Context
	Log    logr.Logger

	OCMAgentHandler ocmagenthandler.OCMAgentHandler
}

var log = logf.Log.WithName("controller_ocmagent")

// blank assignment to verify that ReconcileOCMAgent implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileOCMAgent{}

// Reconcile reads that state of the cluster for a OCMAgent object and makes changes based on the state read
// and what is in the OCMAgent.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileOCMAgent) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	r.Ctx = ctx
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling OCMAgent")

	// Fetch the OCMAgent instance
	instance := ocmagentv1alpha1.OcmAgent{}
	err := r.Client.Get(ctx, request.NamespacedName, &instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			localmetrics.UpdateMetricOcmAgentResourceAbsent()
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to retrieve OCMAgent. Will retry on next reconcile.")
		return reconcile.Result{}, err
	}
	localmetrics.ResetMetricOcmAgentResourceAbsent()

	// Is the OCMAgent being deleted?
	if !instance.DeletionTimestamp.IsZero() {
		log.V(2).Info("Entering EnsureOCMAgentResourcesAbsent")
		err := r.OCMAgentHandler.EnsureOCMAgentResourcesAbsent(instance)
		if err != nil {
			log.Error(err, "Failed to remove OCMAgent. Will retry on next reconcile.")
			return reconcile.Result{}, err
		}
		// The finalizer can now be removed
		if controllerutil.ContainsFinalizer(&instance, ctrlconst.ReconcileOCMAgentFinalizer) {
			controllerutil.RemoveFinalizer(&instance, ctrlconst.ReconcileOCMAgentFinalizer)
			if err := r.Client.Update(ctx, &instance); err != nil {
				log.Error(err, "Failed to remove finalizer from OCMAgent resource. Will retry on next reconcile.")
				return reconcile.Result{}, err
			}
		}
		log.Info("Successfully removed OCMAgent resources.")
	} else {
		// There needs to be an OCM Agent
		log.V(2).Info("Entering EnsureOCMAgentResourcesExist")
		err := r.OCMAgentHandler.EnsureOCMAgentResourcesExist(instance)
		if err != nil {
			log.Error(err, "Failed to create OCMAgent. Will retry on next reconcile.")
			return reconcile.Result{}, err
		}

		// The OCM Agent is deployed, so set a finalizer on the resource
		if !controllerutil.ContainsFinalizer(&instance, ctrlconst.ReconcileOCMAgentFinalizer) {
			controllerutil.AddFinalizer(&instance, ctrlconst.ReconcileOCMAgentFinalizer)
			if err := r.Client.Update(ctx, &instance); err != nil {
				log.Error(err, "Failed to apply finalizer to OCMAgent resource. Will retry on next reconcile.")
				return reconcile.Result{}, err
			}
		}
		log.Info("Successfully setup OCMAgent resources.")
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReconcileOCMAgent) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ocmagentv1alpha1.OcmAgent{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Owns(&netv1.NetworkPolicy{}).
		Owns(&monitoringv1.ServiceMonitor{}).
		Complete(r)
}
