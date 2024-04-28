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

	corev0 "github.com/Open-Digital-Twin/ktwin-operator/api/core/v0"
	eventStore "github.com/Open-Digital-Twin/ktwin-operator/pkg/event-store"
	keventing "knative.dev/eventing/pkg/apis/eventing/v1"
	kserving "knative.dev/serving/pkg/apis/serving/v1"
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

	return r.createOrUpdateEventStoreResources(ctx, eventStore)
}

func (r *EventStoreReconciler) createOrUpdateEventStoreResources(ctx context.Context, eventStore corev0.EventStore) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	newKService := r.EventStore.GetEventStoreService(&eventStore)

	err := r.Create(ctx, newKService, &client.CreateOptions{})

	if err != nil && !errors.IsAlreadyExists(err) {
		logger.Error(err, fmt.Sprintf("Error while creating Event Store service %s", eventStore.Name))
		return ctrl.Result{}, err
	} else if err != nil {
		currentKService := &kserving.Service{}
		err := r.Get(ctx, types.NamespacedName{Namespace: eventStore.Namespace, Name: eventStore.Name}, currentKService)
		if err != nil {
			logger.Error(err, fmt.Sprintf("Error while getting current Event Store service %s", eventStore.Name))
			return ctrl.Result{}, err
		}

		currentKService = r.EventStore.MergeEventStoreService(currentKService, newKService)
		err = r.Update(ctx, currentKService, &client.UpdateOptions{})
		if err != nil {
			logger.Error(err, fmt.Sprintf("Error while updating Event Store service %s", eventStore.Name))
			return ctrl.Result{}, err
		}

	}

	newTrigger := r.EventStore.GetEventStoreTrigger(&eventStore)
	err = r.Create(ctx, newTrigger, &client.CreateOptions{})

	if err != nil && !errors.IsAlreadyExists(err) {
		logger.Error(err, fmt.Sprintf("Error while creating trigger for event store %s", eventStore.Name))
		return ctrl.Result{}, err
	} else if err != nil {
		currentTrigger := &keventing.Trigger{}
		err := r.Get(ctx, types.NamespacedName{Namespace: eventStore.Namespace, Name: eventStore.Name + "-trigger"}, currentTrigger)
		if err != nil {
			logger.Error(err, fmt.Sprintf("Error while getting current Event Store trigger %s", eventStore.Name+"-trigger"))
			return ctrl.Result{}, err
		}

		currentTrigger = r.EventStore.MergeEventStoreTrigger(currentTrigger, newTrigger)
		err = r.Update(ctx, currentTrigger, &client.UpdateOptions{})
		if err != nil {
			logger.Error(err, fmt.Sprintf("Error while updating trigger for event store %s", eventStore.Name))
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EventStoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev0.EventStore{}).
		Complete(r)
}
