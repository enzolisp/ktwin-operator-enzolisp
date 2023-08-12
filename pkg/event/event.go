package event

import (
	"fmt"
	dtdv0 "ktwin/operator/api/dtd/v0"
	broker "ktwin/operator/pkg/broker"
	"ktwin/operator/pkg/event/rabbitmq"
	"strings"

	"github.com/google/uuid"
	rabbitmqv1beta1 "github.com/rabbitmq/messaging-topology-operator/api/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kEventing "knative.dev/eventing/pkg/apis/eventing/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const (
	EVENT_REAL_TO_VIRTUAL    string = "ktwin.real.%s"
	EVENT_VIRTUAL_TO_REAL    string = "ktwin.virtual.%s"
	EVENT_TO_EVENT_STORE     string = "ktwin.event.store"
	EVENT_VIRTUAL_TO_VIRTUAL string = "ktwin.virtual.virtual" // TODO: what if someone wants to send an event to a relationship (post processing - use command)
)

func NewTwinEvent() TwinEvent {
	return &twinEvent{}
}

type TwinEvent interface {
	GetTwinInterfaceTrigger(twinInterface *dtdv0.TwinInterface) kEventing.Trigger
	GetRelationshipBrokerBindings(twinInterface *dtdv0.TwinInterface, twinInterfaceTrigger kEventing.Trigger, brokerExchange rabbitmqv1beta1.Exchange, twinInterfaceQueue rabbitmqv1beta1.Queue) []rabbitmqv1beta1.Binding
	GetMQQTDispatcherBindings(twinInstance *dtdv0.TwinInstance) []rabbitmqv1beta1.Binding
	GetEventStoreBrokerBindings(twinInterface *dtdv0.TwinInterface) []rabbitmqv1beta1.Binding
	GetTriggersDeletionFilterCriteria(namespacedName types.NamespacedName) map[string]string
}

type twinEvent struct{}

type triggerParameters struct {
	interfaceName string
	triggerName   string
	namespace     string
	brokerName    string
	eventType     string
	subscriber    string
}

func (e *twinEvent) getEventTypeRealGenerated(twinInterfaceName string) string {
	return fmt.Sprintf(EVENT_REAL_TO_VIRTUAL, twinInterfaceName)
}

func (e *twinEvent) getEventTypeVirtualGenerated(twinInterfaceName string) string {
	return fmt.Sprintf(EVENT_VIRTUAL_TO_REAL, twinInterfaceName)
}

func (e *twinEvent) getVirtualToVirtualTriggerName(sourceTwinInstanceName string, targetTwinInstanceName string) string {
	return sourceTwinInstanceName + "-to-" + targetTwinInstanceName
}

func (e *twinEvent) getTwinInterfaceTrigger(twinInterfaceName string) string {
	return twinInterfaceName
}

func (e *twinEvent) getRealToEventStoreTriggerName(twinInterfaceName string) string {
	return twinInterfaceName + "-real-to-event-store"
}

func (e *twinEvent) getVirtualToEventStoreTriggerName(twinInterfaceName string) string {
	return twinInterfaceName + "-virtual-to-event-store"
}

func (e *twinEvent) getTriggerLabels(twinInterfaceName string) map[string]string {
	return map[string]string{
		"ktwin/twin-interface": twinInterfaceName,
	}
}

func (e *twinEvent) GetTriggersDeletionFilterCriteria(namespacedName types.NamespacedName) map[string]string {
	return e.getTriggerLabels(namespacedName.Name)
}

func (e *twinEvent) GetMQQTDispatcherBindings(
	twinInstance *dtdv0.TwinInstance,
) []rabbitmqv1beta1.Binding {
	var rabbitMQBindings []rabbitmqv1beta1.Binding

	rabbitMQVirtualBinding, _ := rabbitmq.NewBinding(rabbitmq.BindingArgs{
		Name:      strings.ToLower(twinInstance.Name) + "-real-" + uuid.NewString(),
		Namespace: twinInstance.Namespace,
		Owner: []v1.OwnerReference{
			{
				APIVersion: twinInstance.APIVersion,
				Kind:       twinInstance.Kind,
				Name:       twinInstance.Name,
				UID:        twinInstance.UID,
			},
		},
		RabbitmqClusterReference: &rabbitmqv1beta1.RabbitmqClusterReference{
			Name:      "rabbitmq",
			Namespace: "default",
		},
		RabbitMQVhost: "/",
		Source:        "amq.topic",
		Destination:   "mqtt-dispatcher-queue",
		Labels:        map[string]string{},
		RoutingKey:    e.getEventTypeRealGenerated(twinInstance.Spec.Interface + "." + twinInstance.Name),
	})

	rabbitMQBindings = append(rabbitMQBindings, rabbitMQVirtualBinding)

	for _, relationship := range twinInstance.Spec.TwinInstanceRelationships {
		if relationship.AggregateData {
			rabbitMQVirtualBinding, _ := rabbitmq.NewBinding(rabbitmq.BindingArgs{
				Name:      strings.ToLower(relationship.Name) + "-real-" + uuid.NewString(),
				Namespace: twinInstance.Namespace,
				Owner: []v1.OwnerReference{
					{
						APIVersion: twinInstance.APIVersion,
						Kind:       twinInstance.Kind,
						Name:       twinInstance.Name,
						UID:        twinInstance.UID,
					},
				},
				RabbitmqClusterReference: &rabbitmqv1beta1.RabbitmqClusterReference{
					Name:      "rabbitmq",
					Namespace: "default",
				},
				RabbitMQVhost: "/",
				Source:        "amq.topic",
				Destination:   "mqtt-dispatcher-queue",
				Labels:        map[string]string{},
				RoutingKey:    e.getEventTypeRealGenerated(twinInstance.Spec.Interface + "." + twinInstance.Name),
			})
			rabbitMQBindings = append(rabbitMQBindings, rabbitMQVirtualBinding)
		}
	}

	return rabbitMQBindings
}

func (e *twinEvent) GetRelationshipBrokerBindings(
	twinInterface *dtdv0.TwinInterface,
	twinInterfaceTrigger kEventing.Trigger,
	brokerExchange rabbitmqv1beta1.Exchange,
	twinInterfaceQueue rabbitmqv1beta1.Queue,
) []rabbitmqv1beta1.Binding {
	rabbitMQBindings := []rabbitmqv1beta1.Binding{}
	virtualEventBinding, _ := rabbitmq.NewBinding(rabbitmq.BindingArgs{
		Name:      strings.ToLower(twinInterface.Name) + "-virtual-" + uuid.NewString(),
		Namespace: twinInterface.Namespace,
		Labels: map[string]string{
			"ktwin/twin-interface":         twinInterface.Name,
			"eventing.knative.dev/trigger": twinInterface.Name,
		},
		Filters: map[string]string{
			"type":              e.getEventTypeVirtualGenerated(twinInterface.Name),
			"x-knative-trigger": twinInterface.Name,
			"x-match":           "all",
		},
		RabbitMQVhost: "/",
		Owner: []v1.OwnerReference{
			{
				APIVersion: twinInterface.APIVersion,
				Kind:       twinInterface.Kind,
				Name:       twinInterface.Name,
				UID:        twinInterface.UID,
			},
			{
				APIVersion: twinInterfaceTrigger.APIVersion,
				Kind:       twinInterfaceTrigger.Kind,
				Name:       twinInterfaceTrigger.Name,
				UID:        twinInterfaceTrigger.UID,
			},
		},
		RabbitmqClusterReference: &rabbitmqv1beta1.RabbitmqClusterReference{
			Name:      "rabbitmq",
			Namespace: "default",
		},
		Source:      brokerExchange.Spec.Name,       // broker exchange
		Destination: "cloud-event-dispatcher-queue", // trigger queue
	})

	rabbitMQBindings = append(rabbitMQBindings, virtualEventBinding)

	for _, twinInterfaceRelationship := range twinInterface.Spec.Relationships {
		if twinInterfaceRelationship.AggregateData {
			realEventBinding, _ := rabbitmq.NewBinding(rabbitmq.BindingArgs{
				Name:      strings.ToLower(twinInterface.Name) + "-" + strings.ToLower(twinInterfaceRelationship.Name) + "-real-" + uuid.NewString(),
				Namespace: twinInterface.Namespace,
				Labels: map[string]string{
					"ktwin/twin-interface":         twinInterface.Name,
					"eventing.knative.dev/trigger": twinInterface.Name,
				},
				Filters: map[string]string{
					"type":              e.getEventTypeRealGenerated(twinInterfaceRelationship.Target),
					"x-knative-trigger": twinInterface.Name,
					"x-match":           "all",
				},
				RabbitMQVhost: "/",
				Owner: []v1.OwnerReference{
					{
						APIVersion: twinInterface.APIVersion,
						Kind:       twinInterface.Kind,
						Name:       twinInterface.Name,
						UID:        twinInterface.UID,
					},
					{
						APIVersion: twinInterfaceTrigger.APIVersion,
						Kind:       twinInterfaceTrigger.Kind,
						Name:       twinInterfaceTrigger.Name,
						UID:        twinInterfaceTrigger.UID,
					},
				},
				RabbitmqClusterReference: &rabbitmqv1beta1.RabbitmqClusterReference{
					Name:      "rabbitmq",
					Namespace: "default",
				},
				Source:      brokerExchange.Spec.Name,     // broker exchange
				Destination: twinInterfaceQueue.Spec.Name, // trigger queue
			})

			virtualEventBinding, _ := rabbitmq.NewBinding(rabbitmq.BindingArgs{
				Name:      strings.ToLower(twinInterface.Name) + "-" + strings.ToLower(twinInterfaceRelationship.Name) + "-virtual-" + uuid.NewString(),
				Namespace: twinInterface.Namespace,
				Labels: map[string]string{
					"ktwin/twin-interface":         twinInterface.Name,
					"eventing.knative.dev/trigger": twinInterface.Name,
				},
				Filters: map[string]string{
					"type":              e.getEventTypeVirtualGenerated(twinInterfaceRelationship.Target),
					"x-knative-trigger": twinInterface.Name,
					"x-match":           "all",
				},
				RabbitMQVhost: "/",
				// Check who is going to be owner
				Owner: []v1.OwnerReference{
					{
						APIVersion: twinInterface.APIVersion,
						Kind:       twinInterface.Kind,
						Name:       twinInterface.Name,
						UID:        twinInterface.UID,
					},
					// {
					// 	APIVersion: twinInterfaceTrigger.APIVersion,
					// 	Kind:       twinInterfaceTrigger.Kind,
					// 	Name:       twinInterfaceTrigger.Name,
					// 	UID:        twinInterfaceTrigger.UID,
					// },
				},
				RabbitmqClusterReference: &rabbitmqv1beta1.RabbitmqClusterReference{
					Name:      "rabbitmq",
					Namespace: "default",
				},
				Source:      brokerExchange.Spec.Name,       // broker exchange
				Destination: "cloud-event-dispatcher-queue", // trigger queue
			})

			rabbitMQBindings = append(rabbitMQBindings, virtualEventBinding)
			rabbitMQBindings = append(rabbitMQBindings, realEventBinding)
		}
	}

	return rabbitMQBindings
}

func (e *twinEvent) GetEventStoreBrokerBindings(twinInterface *dtdv0.TwinInterface) []rabbitmqv1beta1.Binding {
	rabbitMQBindings := []rabbitmqv1beta1.Binding{}
	return rabbitMQBindings
}

func (e *twinEvent) GetTwinInterfaceTrigger(twinInterface *dtdv0.TwinInterface) kEventing.Trigger {
	var twinInterfaceTrigger kEventing.Trigger

	virtualTwinService := twinInterface.Name

	// If TwinInstance has container associated, create the triggers
	if e.hasContainerInTwinInterface(twinInterface) {
		// Real Twin Event Type
		twinInterfaceEventType := e.getEventTypeRealGenerated(twinInterface.Name)

		twinInterfaceTrigger = e.createTrigger(triggerParameters{
			triggerName:   e.getTwinInterfaceTrigger(twinInterface.Name),
			namespace:     twinInterface.Namespace,
			brokerName:    broker.EVENT_BROKER_NAME,
			eventType:     twinInterfaceEventType,
			subscriber:    virtualTwinService,
			interfaceName: twinInterface.Name,
		})

	}

	// // Real to Event Store
	// trigger = e.createTrigger(triggerParameters{
	// 	triggerName:   e.getRealToEventStoreTriggerName(twinInterface.Name),
	// 	namespace:     twinInterface.Namespace,
	// 	brokerName:    broker.EVENT_BROKER_NAME,
	// 	eventType:     e.getEventTypeRealGenerated(twinInterface.Name),
	// 	subscriber:    eventStore.EVENT_STORE_SERVICE,
	// 	interfaceName: twinInterface.Name,
	// })
	// twinTriggers = append(twinTriggers, trigger)

	// // Virtual to Event Store
	// trigger = e.createTrigger(triggerParameters{
	// 	triggerName:   e.getVirtualToEventStoreTriggerName(twinInterface.Name),
	// 	namespace:     twinInterface.Namespace,
	// 	brokerName:    broker.EVENT_BROKER_NAME,
	// 	eventType:     e.getEventTypeVirtualGenerated(twinInterface.Name),
	// 	subscriber:    eventStore.EVENT_STORE_SERVICE,
	// 	interfaceName: twinInterface.Name,
	// })
	// twinTriggers = append(twinTriggers, trigger)

	//e.populateTwinInstanceEventStructure(twinInstance, twinTriggers)

	return twinInterfaceTrigger
}

func (e *twinEvent) populateTwinInstanceEventStructure(twinInstance *dtdv0.TwinInstance, twinTriggers []kEventing.Trigger) *dtdv0.TwinInstance {
	var twinEvents []dtdv0.TwinInstanceEvents
	for _, twinTrigger := range twinTriggers {
		attributesMap := twinTrigger.Spec.Filter.Attributes
		twinInstanceEvents := dtdv0.TwinInstanceEvents{
			Filters: dtdv0.TwinInstanceEventsFilters{
				Exact: dtdv0.TwinInstanceEventsFiltersAttributes(attributesMap),
			},
			Sink: dtdv0.TwinInterfaceEventsSink{
				InstanceId: twinTrigger.Spec.Subscriber.Ref.Name,
			},
		}
		twinEvents = append(twinEvents, twinInstanceEvents)
	}

	twinInstance.Spec.Events = twinEvents

	return twinInstance
}

func (e *twinEvent) createTrigger(triggerParameters triggerParameters) kEventing.Trigger {
	return kEventing.Trigger{
		TypeMeta: v1.TypeMeta{
			Kind:       "Trigger",
			APIVersion: "eventing.knative.dev/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      triggerParameters.triggerName,
			Namespace: triggerParameters.namespace,
			Labels:    e.getTriggerLabels(triggerParameters.interfaceName),
		},
		Spec: kEventing.TriggerSpec{
			Broker: triggerParameters.brokerName,
			Filter: &kEventing.TriggerFilter{
				Attributes: map[string]string{
					"type": triggerParameters.eventType,
				},
			},
			Subscriber: duckv1.Destination{
				Ref: &duckv1.KReference{
					Kind:       "Service",
					APIVersion: "serving.knative.dev/v1",
					Name:       triggerParameters.subscriber,
				},
			},
		},
	}
}

func (*twinEvent) hasContainerInTwinInterface(twinInterface *dtdv0.TwinInterface) bool {
	return twinInterface.Spec.Service != nil
}
