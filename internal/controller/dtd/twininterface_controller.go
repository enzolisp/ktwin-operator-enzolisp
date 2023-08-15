/*
Copyright 2023.

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

package dtd

import (
	"context"
	"fmt"

	rabbitmqv1beta1 "github.com/rabbitmq/messaging-topology-operator/api/v1beta1"
	kEventing "knative.dev/eventing/pkg/apis/eventing/v1"
	kServing "knative.dev/serving/pkg/apis/serving/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dtdv0 "ktwin/operator/api/dtd/v0"
	twinevent "ktwin/operator/pkg/event"
	eventStore "ktwin/operator/pkg/event-store"
	twinservice "ktwin/operator/pkg/service"
)

// TwinInterfaceReconciler reconciles a TwinInterface object
type TwinInterfaceReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	TwinService twinservice.TwinService
	TwinEvent   twinevent.TwinEvent
	EventStore  eventStore.EventStore
}

//+kubebuilder:rbac:groups=dtd.ktwin,resources=twininterfaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dtd.ktwin,resources=twininterfaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dtd.ktwin,resources=twininterfaces/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TwinInterface object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *TwinInterfaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	twinInterface := &dtdv0.TwinInterface{}
	err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, twinInterface)

	// Delete scenario
	if err != nil {
		if errors.IsNotFound(err) {
			return r.deleteTwinInterface(ctx, req, req.NamespacedName)
		}
		logger.Error(err, fmt.Sprintf("Unexpected error while deleting TwinInstance %s", req.Name))
		return ctrl.Result{}, err
	}

	// Create Mew Schema in Event Store

	return r.createUpdateTwinInterface(ctx, req, twinInterface)
}

func (r *TwinInterfaceReconciler) deleteTwinInterface(ctx context.Context, req ctrl.Request, namespacedName types.NamespacedName) (ctrl.Result, error) {
	var errorsResult []error
	logger := log.FromContext(ctx)

	// Delete Service Instance
	deletionServiceLabels := r.TwinService.GetServiceDeletionCriteria(namespacedName)

	kServiceList := kServing.ServiceList{}
	kServiceListOptions := []client.ListOption{
		client.InNamespace(namespacedName.Namespace),
		client.MatchingLabels(deletionServiceLabels),
	}

	err := r.List(ctx, &kServiceList, kServiceListOptions...)

	if err != nil {
		logger.Error(err, fmt.Sprintf("Error while getting services to be deleted %s", namespacedName.Name))
		return ctrl.Result{}, err
	}

	for _, kService := range kServiceList.Items {
		err := r.Delete(ctx, &kService, &client.DeleteOptions{})
		if err != nil {
			if !errors.IsNotFound(err) {
				logger.Error(err, fmt.Sprintf("Error while deleting Knative Service %s", namespacedName.Name))
				errorsResult = append(errorsResult, err)
			}
		}
	}

	// Delete Triggers
	deletionTriggerLabels := r.TwinEvent.GetTriggersDeletionFilterCriteria(namespacedName)
	triggerList := kEventing.TriggerList{}
	triggerListOptions := []client.ListOption{
		client.InNamespace(namespacedName.Namespace),
		client.MatchingLabels(deletionTriggerLabels),
	}

	err = r.List(ctx, &triggerList, triggerListOptions...)

	if err != nil {
		logger.Error(err, fmt.Sprintf("Error while getting triggers %s", namespacedName.Name))
		return ctrl.Result{}, err
	}

	for _, trigger := range triggerList.Items {
		err := r.Delete(ctx, &trigger, &client.DeleteOptions{})
		if err != nil {
			if !errors.IsNotFound(err) {
				logger.Error(err, fmt.Sprintf("Error while deleting trigger %s", namespacedName.Name))
				errorsResult = append(errorsResult, err)
			}
		}
	}

	if len(errorsResult) > 0 {
		return ctrl.Result{}, errorsResult[0]
	}

	return ctrl.Result{}, nil
}

func (r *TwinInterfaceReconciler) createUpdateTwinInterface(ctx context.Context, req ctrl.Request, twinInterface *dtdv0.TwinInterface) (ctrl.Result, error) {
	twinInterfaceName := twinInterface.ObjectMeta.Name

	var resultErrors []error
	logger := log.FromContext(ctx)

	// Create Service Instance and Trigger, if pod is specified
	if twinInterface.Spec.Service != nil {
		kService := r.TwinService.GetService(twinInterface)
		err := r.Create(ctx, kService, &client.CreateOptions{})

		if err != nil && !errors.IsAlreadyExists(err) {
			logger.Error(err, fmt.Sprintf("Error while creating Knative Service %s", twinInterfaceName))
			resultErrors = append(resultErrors, err)
		}

		// Create Trigger
		trigger := r.TwinEvent.GetTwinInterfaceTrigger(twinInterface)
		err = r.Create(ctx, &trigger, &client.CreateOptions{})
		if err != nil && !errors.IsAlreadyExists(err) {
			logger.Error(err, fmt.Sprintf("Error while creating Twin Events %s", twinInterfaceName))
			resultErrors = append(resultErrors, err)
		}

		// Get Created Trigger
		err = r.Get(ctx, types.NamespacedName{Namespace: trigger.Namespace, Name: trigger.Name}, &trigger, &client.GetOptions{})
		if err != nil && !errors.IsAlreadyExists(err) {
			logger.Error(err, fmt.Sprintf("Error while getting trigger %s", trigger.Name))
			resultErrors = append(resultErrors, err)
		}

		// Create Relationship RabbitMQ bindings to existing Queue and Eventing
		// RabbitMQ exchange (Broker): https://github.com/knative-extensions/eventing-rabbitmq/blob/main/pkg/reconciler/broker/broker.go#L133
		// RabbitMQ Queue (Trigger): https://github.com/knative-extensions/eventing-rabbitmq/blob/main/pkg/reconciler/trigger/trigger.go#L233
		// Deletion: Can use ownerReferences for deletion in cascade

		eventStoreQueuesList := rabbitmqv1beta1.QueueList{}
		queueListOptions := []client.ListOption{
			client.InNamespace(twinInterface.Namespace),
			client.MatchingLabels(client.MatchingFields{
				"eventing.knative.dev/trigger": "event-store-trigger",
			}),
		}

		err = r.List(ctx, &eventStoreQueuesList, queueListOptions...)

		if len(eventStoreQueuesList.Items) == 0 {
			logger.Error(err, fmt.Sprintf("No Queue found for event store %s", twinInterfaceName))
			resultErrors = append(resultErrors, err)
			return ctrl.Result{}, err
		}

		exchangeList := rabbitmqv1beta1.ExchangeList{}
		exchangeListOptions := []client.ListOption{
			client.InNamespace(twinInterface.Namespace),
			client.MatchingLabels(client.MatchingFields{
				"eventing.knative.dev/broker": "default",
			}),
		}

		err = r.List(ctx, &exchangeList, exchangeListOptions...)

		if err != nil {
			logger.Error(err, fmt.Sprintf("Error while getting default broker exchange"))
			resultErrors = append(resultErrors, err)
		}

		if len(exchangeList.Items) == 0 {
			logger.Error(err, fmt.Sprintf("No Broker Exchange found for TwinInterface %s", twinInterfaceName))
			resultErrors = append(resultErrors, err)
		} else {
			queueList := rabbitmqv1beta1.QueueList{}
			queueListOptions := []client.ListOption{
				client.InNamespace(twinInterface.Namespace),
				client.MatchingLabels(client.MatchingFields{
					"eventing.knative.dev/broker":  "default",
					"eventing.knative.dev/trigger": twinInterface.Name,
				}),
			}

			err = r.List(ctx, &queueList, queueListOptions...)

			if len(queueList.Items) == 0 {
				logger.Error(err, fmt.Sprintf("No Broker Queue found for TwinInterface %s", twinInterfaceName))
				resultErrors = append(resultErrors, err)
			} else {
				brokerExchange := exchangeList.Items[0]
				twinInterfaceQueue := queueList.Items[0]
				bindings := r.TwinEvent.GetRelationshipBrokerBindings(twinInterface, trigger, brokerExchange, twinInterfaceQueue)

				for _, binding := range bindings {
					err = r.Create(ctx, &binding, &client.CreateOptions{})
					if err != nil && !errors.IsAlreadyExists(err) {
						logger.Error(err, fmt.Sprintf("Error while creating TwinInterface Binding %s", binding.Name))
						resultErrors = append(resultErrors, err)
					}
				}

				eventStoreQueue := eventStoreQueuesList.Items[0]
				bindings = r.EventStore.GetEventStoreBrokerBindings(twinInterface, trigger, brokerExchange, eventStoreQueue)

				for _, binding := range bindings {
					err = r.Create(ctx, &binding, &client.CreateOptions{})
					if err != nil && !errors.IsAlreadyExists(err) {
						logger.Error(err, fmt.Sprintf("Error while creating EventStore TwinInterface Bindings %s", binding.Name))
						resultErrors = append(resultErrors, err)
					}
				}

			}
		}
	}

	if len(resultErrors) > 0 {
		twinInterface.Status.Status = dtdv0.TwinInterfacePhaseFailed
		return ctrl.Result{}, resultErrors[0]
	} else {
		twinInterface.Status.Status = dtdv0.TwinInterfacePhaseRunning
	}

	twinInterface.Labels = map[string]string{
		"ktwin/twin-interface": twinInterfaceName,
	}

	// Update Status for Running or Failed
	_, err := r.updateTwinInterface(ctx, req, twinInterface)

	if err != nil {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *TwinInterfaceReconciler) updateTwinInterface(ctx context.Context, req ctrl.Request, twinInterface *dtdv0.TwinInterface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	err := r.Update(ctx, twinInterface, &client.UpdateOptions{})

	if err != nil {
		logger.Error(err, fmt.Sprintf("Error while updating TwinInterface %s", twinInterface.ObjectMeta.Name))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TwinInterfaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dtdv0.TwinInterface{}).
		Complete(r)
}
