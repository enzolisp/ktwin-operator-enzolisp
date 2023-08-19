package core

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev0 "ktwin/operator/api/core/v0"
	eventStore "ktwin/operator/pkg/event-store"
)

// EventStoreReconciler reconciles a EventStore object
type EventStoreReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	EventStore eventStore.EventStore
}

//+kubebuilder:rbac:groups=core.ktwin,resources=eventstores,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.ktwin,resources=eventstores/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.ktwin,resources=eventstores/finalizers,verbs=update

func (r *EventStoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	eventStore := corev0.EventStore{}
	err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, &eventStore)

	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, fmt.Sprintf("Unexpected error while deleting TwinInstance %s", req.Name))
		return ctrl.Result{}, err
	}

	return r.createOrUpdateMQTTTrigger(ctx, req, eventStore)
}

func (r *EventStoreReconciler) createOrUpdateMQTTTrigger(ctx context.Context, req ctrl.Request, eventStore corev0.EventStore) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	kservice := r.EventStore.GetEventStoreService(&eventStore)

	err := r.Create(ctx, kservice, &client.CreateOptions{})
	if err != nil {
		logger.Error(err, fmt.Sprintf("Error while creating Event Store service %s", eventStore.Name))
		return ctrl.Result{}, err
	}

	trigger := r.EventStore.GetEventStoreTrigger(&eventStore)
	err = r.Create(ctx, &trigger, &client.CreateOptions{})
	if err != nil {
		logger.Error(err, fmt.Sprintf("Error while creating trigger for event store %s", eventStore.Name))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EventStoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev0.EventStore{}).
		Complete(r)
}
