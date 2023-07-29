package event

import (
	"fmt"
	dtdv0 "ktwin/operator/api/dtd/v0"
	broker "ktwin/operator/internal/resources/broker"
	eventStore "ktwin/operator/internal/resources/event-store"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kEventing "knative.dev/eventing/pkg/apis/eventing/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const (
	EVENT_REAL_TO_VIRTUAL    string = "ktwin.real.%s.generated"
	EVENT_VIRTUAL_TO_REAL    string = "ktwin.virtual.%s.generated"
	EVENT_TO_EVENT_STORE     string = "ktwin.event.store.generated"
	EVENT_VIRTUAL_TO_VIRTUAL string = "ktwin.virtual.virtual.generated" // TODO: what if someone wants to send an event to a relationship (post processing - use command)
)

func NewTwinEvent() TwinEvent {
	return &twinEvent{}
}

type TwinEvent interface {
	GetTriggers(twinInstance *dtdv0.TwinInstance, twinInterface *dtdv0.TwinInterface) []kEventing.Trigger
	GetTriggersDeletionFilterCriteria(namespacedName types.NamespacedName) map[string]string
}

type twinEvent struct{}

type triggerParameters struct {
	instanceName string
	triggerName  string
	namespace    string
	brokerName   string
	eventType    string
	eventSource  string
	subscriber   string
}

func (e *twinEvent) getEventTypeRealToVirtual(twinInstanceName string) string {
	return fmt.Sprintf(EVENT_REAL_TO_VIRTUAL, twinInstanceName)
}

func (e *twinEvent) getEventTypeVirtualToReal(twinInstanceName string) string {
	return fmt.Sprintf(EVENT_VIRTUAL_TO_REAL, twinInstanceName)
}

func (e *twinEvent) getVirtualToVirtualTriggerName(sourceTwinInstanceName string, targetTwinInstanceName string) string {
	return sourceTwinInstanceName + "-to-" + targetTwinInstanceName
}

func (e *twinEvent) getRealToVirtualTriggerName(twinInstanceName string) string {
	return twinInstanceName + "-real-to-virtual"
}

func (e *twinEvent) getVirtualToRealTriggerName(twinInstanceName string) string {
	return twinInstanceName + "-virtual-to-real"
}

func (e *twinEvent) getRealToEventStoreTriggerName(twinInstanceName string) string {
	return twinInstanceName + "-real-to-event-store"
}

func (e *twinEvent) getVirtualToEventStoreTriggerName(twinInstanceName string) string {
	return twinInstanceName + "-virtual-to-event-store"
}

func (e *twinEvent) getTriggerLabels(twinInstanceName string) map[string]string {
	return map[string]string{
		"ktwin/twininstance": twinInstanceName,
	}
}

func (e *twinEvent) GetTriggersDeletionFilterCriteria(namespacedName types.NamespacedName) map[string]string {
	return e.getTriggerLabels(namespacedName.Name)
}

func (e *twinEvent) GetTriggers(twinInstance *dtdv0.TwinInstance, twinInterface *dtdv0.TwinInterface) []kEventing.Trigger {
	var twinTriggers []kEventing.Trigger
	var trigger kEventing.Trigger

	realEventSource := twinInstance.Name + "-mqtt"
	virtualTwinService := twinInterface.Name

	// If twin instance has container associated, create the triggers
	if e.hasContainerInTwinInterface(twinInterface) {
		// Real to Virtual
		trigger = e.createTrigger(triggerParameters{
			triggerName:  e.getRealToVirtualTriggerName(twinInstance.Name),
			namespace:    twinInstance.Namespace,
			brokerName:   broker.EVENT_BROKER_NAME,
			eventType:    e.getEventTypeRealToVirtual(twinInstance.Name),
			eventSource:  realEventSource,
			subscriber:   virtualTwinService,
			instanceName: twinInstance.Name,
		})
		twinTriggers = append(twinTriggers, trigger)

		// Virtual to Real
		trigger = e.createTrigger(triggerParameters{
			triggerName:  e.getVirtualToRealTriggerName(twinInstance.Name),
			namespace:    twinInstance.Namespace,
			brokerName:   broker.EVENT_BROKER_NAME,
			eventType:    e.getEventTypeVirtualToReal(twinInstance.Name),
			eventSource:  virtualTwinService,
			subscriber:   realEventSource,
			instanceName: twinInstance.Name,
		})
		twinTriggers = append(twinTriggers, trigger)

		// Virtual to virtual
		for _, twinInstanceRelationship := range twinInstance.Spec.TwinInstanceRelationships {
			if twinInstanceRelationship.AggregateData {
				realEventSource := twinInstanceRelationship.Target + "-mqtt"
				trigger = e.createTrigger(triggerParameters{
					triggerName:  e.getVirtualToVirtualTriggerName(twinInstanceRelationship.Target, twinInstance.Name),
					namespace:    twinInstance.Namespace,
					brokerName:   broker.EVENT_BROKER_NAME,
					eventType:    e.getEventTypeRealToVirtual(twinInstanceRelationship.Target),
					eventSource:  realEventSource,
					subscriber:   virtualTwinService,
					instanceName: twinInstance.Name,
				})
				twinTriggers = append(twinTriggers, trigger)
			}
		}

	}

	// Real to Event Store
	trigger = e.createTrigger(triggerParameters{
		triggerName:  e.getRealToEventStoreTriggerName(twinInstance.Name),
		namespace:    twinInstance.Namespace,
		brokerName:   broker.EVENT_BROKER_NAME,
		eventType:    e.getEventTypeRealToVirtual(twinInstance.Name),
		eventSource:  realEventSource,
		subscriber:   eventStore.EVENT_STORE_SERVICE,
		instanceName: twinInstance.Name,
	})
	twinTriggers = append(twinTriggers, trigger)

	// Virtual to Event Store
	trigger = e.createTrigger(triggerParameters{
		triggerName:  e.getVirtualToEventStoreTriggerName(twinInstance.Name),
		namespace:    twinInstance.Namespace,
		brokerName:   broker.EVENT_BROKER_NAME,
		eventType:    e.getEventTypeVirtualToReal(twinInstance.Name),
		eventSource:  virtualTwinService,
		subscriber:   eventStore.EVENT_STORE_SERVICE,
		instanceName: twinInstance.Name,
	})
	twinTriggers = append(twinTriggers, trigger)

	e.populateTwinInstanceEventStructure(twinInstance, twinTriggers)

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
			Labels:    e.getTriggerLabels(triggerParameters.instanceName),
		},
		Spec: kEventing.TriggerSpec{
			Broker: triggerParameters.brokerName,
			Filter: &kEventing.TriggerFilter{
				Attributes: map[string]string{
					"type":   triggerParameters.eventType,
					"source": triggerParameters.eventSource,
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
	return !reflect.DeepEqual(twinInterface.Spec.Template, corev1.PodTemplateSpec{})
}
