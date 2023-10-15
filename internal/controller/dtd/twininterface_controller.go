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
	"reflect"

	rabbitmqv1beta1 "github.com/rabbitmq/messaging-topology-operator/api/v1beta1"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dtdv0 "github.com/Open-Digital-Twin/ktwin-operator/api/dtd/v0"
	twinevent "github.com/Open-Digital-Twin/ktwin-operator/pkg/event"
	eventStore "github.com/Open-Digital-Twin/ktwin-operator/pkg/event-store"
	twinservice "github.com/Open-Digital-Twin/ktwin-operator/pkg/service"
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

func (r *TwinInterfaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	twinInterface := &dtdv0.TwinInterface{}
	err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, twinInterface)

	// Delete scenario
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, fmt.Sprintf("Unexpected error while deleting TwinInstance %s", req.Name))
		return ctrl.Result{}, err
	}

	// TODO: Create Entry in Event Store

	return r.createUpdateTwinInterface(ctx, req, twinInterface)
}

func (r *TwinInterfaceReconciler) createUpdateTwinInterface(ctx context.Context, req ctrl.Request, twinInterface *dtdv0.TwinInterface) (ctrl.Result, error) {
	twinInterfaceName := twinInterface.ObjectMeta.Name

	var resultErrors []error
	var twinInterfaceTrigger eventingv1.Trigger
	logger := log.FromContext(ctx)

	// Create Service Instance and Trigger, if pod is specified
	if twinInterface.Spec.Service != nil {
		// Get Broker
		broker := eventingv1.Broker{}
		err := r.Get(ctx, types.NamespacedName{Namespace: "ktwin", Name: "ktwin"}, &broker)

		if err != nil {
			logger.Error(err, fmt.Sprintf("Error while getting Broker"))
			resultErrors = append(resultErrors, err)
		}

		// Get Event Store
		eventStoreService := servingv1.Service{}
		err = r.Get(ctx, types.NamespacedName{Namespace: "ktwin", Name: "event-store"}, &eventStoreService)

		if err != nil {
			logger.Error(err, fmt.Sprintf("Error while getting event store"))
			resultErrors = append(resultErrors, err)
		}

		// Get all TwinInstances
		twinInstanceList := dtdv0.TwinInstanceList{}
		listOption := []client.ListOption{
			client.InNamespace("ktwin"),
			// client.MatchingLabels(client.MatchingFields{
			// 	"ktwin/twin-interface": twinInterfaceName,
			// }),
		}

		err = r.List(ctx, &twinInstanceList, listOption...)

		if err != nil {
			logger.Error(err, fmt.Sprintf("Error while getting TwinInstances"))
			resultErrors = append(resultErrors, err)
		}

		kService := r.TwinService.GetService(twinservice.TwinServiceParameters{
			TwinInterface: twinInterface,
			Broker:        broker,
			Service:       eventStoreService,
			TwinInstances: twinInstanceList.Items,
		})

		oldKService := servingv1.Service{}
		err = r.Get(ctx, types.NamespacedName{Namespace: kService.Namespace, Name: kService.Name}, &oldKService)

		if !reflect.DeepEqual(oldKService, servingv1.Service{}) {
			oldContainers := oldKService.Spec.Template.Spec.Containers
			newContainers := kService.Spec.Template.Spec.Containers

			if !reflect.DeepEqual(oldContainers[0].Env, newContainers[0].Env) {
				_, err = controllerutil.CreateOrUpdate(ctx, r.Client, kService, func() error {
					return r.Patch(ctx, kService, client.MergeFrom(kService.DeepCopy()))
				})
			} else {
				logger.Info("No changes in environment variables, so no change in the KService: " + oldKService.Name)
			}
		} else if err != nil && errors.IsNotFound(err) {
			err = r.Create(ctx, kService, &client.CreateOptions{})
		}

		if err != nil {
			logger.Error(err, fmt.Sprintf("Error while creating Knative Service %s", twinInterfaceName))
			resultErrors = append(resultErrors, err)
		}

		// Create Trigger
		twinInterfaceTrigger = r.TwinEvent.GetTwinInterfaceTrigger(twinInterface)
		err = r.Create(ctx, &twinInterfaceTrigger, &client.CreateOptions{})
		if err != nil && !errors.IsAlreadyExists(err) {
			logger.Error(err, fmt.Sprintf("Error while creating Twin Events %s", twinInterfaceName))
			resultErrors = append(resultErrors, err)
		}

		// Get Created Trigger
		err = r.Get(ctx, types.NamespacedName{Namespace: twinInterfaceTrigger.Namespace, Name: twinInterfaceTrigger.Name}, &twinInterfaceTrigger, &client.GetOptions{})
		if err != nil && !errors.IsAlreadyExists(err) {
			logger.Error(err, fmt.Sprintf("Error while getting trigger %s", twinInterfaceTrigger.Name))
			resultErrors = append(resultErrors, err)
		}

	}

	// Create Relationship RabbitMQ bindings to existing Queue and Eventing
	// RabbitMQ exchange (Broker): https://github.com/knative-extensions/eventing-rabbitmq/blob/main/pkg/reconciler/broker/broker.go#L133
	// RabbitMQ Queue (Trigger): https://github.com/knative-extensions/eventing-rabbitmq/blob/main/pkg/reconciler/trigger/trigger.go#L233
	// Deletion: Can use ownerReferences for deletion in cascade

	eventStoreQueue, err := r.getEventStoreQueue(ctx, req, twinInterface)
	if err != nil {
		logger.Error(err, fmt.Sprintf("No Queue found for event store %s", twinInterfaceName))
		resultErrors = append(resultErrors, err)
		return ctrl.Result{}, err
	}

	brokerExchange, err := r.getBrokerExchange(ctx, req, twinInterface)

	if err != nil {
		logger.Error(err, fmt.Sprintf("No Broker Exchange found for TwinInterface %s", twinInterfaceName))
		resultErrors = append(resultErrors, err)
	} else {

		bindings := r.EventStore.GetEventStoreBrokerBindings(twinInterface, brokerExchange, eventStoreQueue)

		for _, binding := range bindings {
			err = r.Create(ctx, &binding, &client.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				logger.Error(err, fmt.Sprintf("Error while creating EventStore TwinInterface Bindings %s", binding.Name))
				resultErrors = append(resultErrors, err)
			}
		}

		twinInterfaceQueue, err := r.getTwinInterfaceQueue(ctx, req, twinInterface)

		if err != nil {
			if !errors.IsNotFound(err) {
				logger.Error(err, fmt.Sprintf("Error while getting TwinInterface %s Queue", twinInterfaceName))
				resultErrors = append(resultErrors, err)
			}
		} else {
			if !reflect.DeepEqual(eventingv1.Trigger{}, twinInterfaceTrigger) {
				bindings := r.TwinEvent.GetRelationshipBrokerBindings(twinInterface, twinInterfaceTrigger, brokerExchange, twinInterfaceQueue)

				for _, binding := range bindings {
					err = r.Create(ctx, &binding, &client.CreateOptions{})
					if err != nil && !errors.IsAlreadyExists(err) {
						logger.Error(err, fmt.Sprintf("Error while creating TwinInterface Binding %s", binding.Name))
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
	_, err = r.updateTwinInterface(ctx, req, twinInterface)

	if err != nil {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *TwinInterfaceReconciler) getEventStoreQueue(ctx context.Context, req ctrl.Request, twinInterface *dtdv0.TwinInterface) (rabbitmqv1beta1.Queue, error) {
	logger := log.FromContext(ctx)
	eventStoreQueuesList := rabbitmqv1beta1.QueueList{}
	queueListOptions := []client.ListOption{
		client.InNamespace(twinInterface.Namespace),
		client.MatchingLabels(client.MatchingFields{
			"eventing.knative.dev/trigger": "event-store-trigger",
		}),
	}

	err := r.List(ctx, &eventStoreQueuesList, queueListOptions...)

	if len(eventStoreQueuesList.Items) == 0 {
		logger.Error(err, fmt.Sprintf("No Queue found for event store %s", twinInterface.Name))
		return rabbitmqv1beta1.Queue{}, err
	}
	return eventStoreQueuesList.Items[0], nil
}

func (r *TwinInterfaceReconciler) getBrokerExchange(ctx context.Context, req ctrl.Request, twinInterface *dtdv0.TwinInterface) (rabbitmqv1beta1.Exchange, error) {
	logger := log.FromContext(ctx)
	exchangeList := rabbitmqv1beta1.ExchangeList{}
	exchangeListOptions := []client.ListOption{
		client.InNamespace(twinInterface.Namespace),
		client.MatchingLabels(client.MatchingFields{
			"eventing.knative.dev/broker": "ktwin",
		}),
	}

	err := r.List(ctx, &exchangeList, exchangeListOptions...)

	if err != nil {
		logger.Error(err, fmt.Sprintf("Error while getting default broker exchange"))
		return rabbitmqv1beta1.Exchange{}, err
	}

	return exchangeList.Items[0], nil
}

func (r *TwinInterfaceReconciler) getTwinInterfaceQueue(ctx context.Context, req ctrl.Request, twinInterface *dtdv0.TwinInterface) (rabbitmqv1beta1.Queue, error) {
	queueList := rabbitmqv1beta1.QueueList{}
	queueListOptions := []client.ListOption{
		client.InNamespace(twinInterface.Namespace),
		client.MatchingLabels(client.MatchingFields{
			"eventing.knative.dev/broker":  "ktwin",
			"eventing.knative.dev/trigger": twinInterface.Name,
		}),
	}

	err := r.List(ctx, &queueList, queueListOptions...)

	if err != nil {
		return rabbitmqv1beta1.Queue{}, err
	}

	if len(queueList.Items) == 0 {
		return rabbitmqv1beta1.Queue{}, errors.NewNotFound(rabbitmqv1beta1.Resource("rabbitmqv1beta1.Queue"), twinInterface.Name)
	}

	return queueList.Items[0], nil
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
