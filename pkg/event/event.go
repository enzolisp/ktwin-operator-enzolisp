package event

import (
	"fmt"
	dtdv0 "ktwin/operator/api/dtd/v0"
	broker "ktwin/operator/pkg/broker"
	eventStore "ktwin/operator/pkg/event-store"

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
	GetTriggers(twinInterface *dtdv0.TwinInterface) []kEventing.Trigger
	GetRelationshipBindings(twinInterface *dtdv0.TwinInterface) (*rabbitmqv1beta1.Binding, error)
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

// Knative Triggers do not allow to route multiple events `types` to the same dispatcher service
// The New Trigger filters feature should provide this feature, but it is not yet developed in RabbitMQ Broker resources
// https://knative.dev/docs/eventing/experimental-features/new-trigger-filters
// In order to reduce the number of pods created in KTWIN cluster, we will be creating only one dispatcher per TwinInterface.
// In case multiple event types needs to be routed to the same dispatcher/TwinInterface KService (twin relationship scenario),
// KTWIN creates Binding resources using rabbitmq/messaging-topology-operator and redirect those messages to the same Knative dispatcher.
// This can be revisit when Knative has this feature available out-of-the-box in RabbitMQ Broker
func (e *twinEvent) GetRelationshipBindings(twinInterface *dtdv0.TwinInterface) (*rabbitmqv1beta1.Binding, error) {
	eventTypes := []string{}
	// Relationship event
	for _, twinInterfaceRelationship := range twinInterface.Spec.Relationships {
		if twinInterfaceRelationship.AggregateData {
			eventTypes = append(eventTypes, e.getEventTypeRealGenerated(twinInterfaceRelationship.Target))
		}
	}
	return nil, nil
}

func (e *twinEvent) GetTriggers(twinInterface *dtdv0.TwinInterface) []kEventing.Trigger {
	var twinTriggers []kEventing.Trigger
	var trigger kEventing.Trigger

	// realEventSource := twinInstance.Name
	virtualTwinService := twinInterface.Name

	// If TwinInstance has container associated, create the triggers
	if e.hasContainerInTwinInterface(twinInterface) {
		// Real Twin Event Type
		twinInterfaceEventType := e.getEventTypeRealGenerated(twinInterface.Name)

		trigger = e.createTrigger(triggerParameters{
			triggerName:   e.getTwinInterfaceTrigger(twinInterface.Name),
			namespace:     twinInterface.Namespace,
			brokerName:    broker.EVENT_BROKER_NAME,
			eventType:     twinInterfaceEventType,
			subscriber:    virtualTwinService,
			interfaceName: twinInterface.Name,
		})
		twinTriggers = append(twinTriggers, trigger)

	}

	// Real to Event Store
	trigger = e.createTrigger(triggerParameters{
		triggerName:   e.getRealToEventStoreTriggerName(twinInterface.Name),
		namespace:     twinInterface.Namespace,
		brokerName:    broker.EVENT_BROKER_NAME,
		eventType:     e.getEventTypeRealGenerated(twinInterface.Name),
		subscriber:    eventStore.EVENT_STORE_SERVICE,
		interfaceName: twinInterface.Name,
	})
	twinTriggers = append(twinTriggers, trigger)

	// Virtual to Event Store
	trigger = e.createTrigger(triggerParameters{
		triggerName:   e.getVirtualToEventStoreTriggerName(twinInterface.Name),
		namespace:     twinInterface.Namespace,
		brokerName:    broker.EVENT_BROKER_NAME,
		eventType:     e.getEventTypeVirtualGenerated(twinInterface.Name),
		subscriber:    eventStore.EVENT_STORE_SERVICE,
		interfaceName: twinInterface.Name,
	})
	twinTriggers = append(twinTriggers, trigger)

	//e.populateTwinInstanceEventStructure(twinInstance, twinTriggers)

	return twinTriggers
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
