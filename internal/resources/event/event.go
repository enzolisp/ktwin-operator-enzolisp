package event

import (
	dtdv0 "ktwin/operator/api/dtd/v0"
	broker "ktwin/operator/internal/resources/broker"
	eventStore "ktwin/operator/internal/resources/event-store"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kEventing "knative.dev/eventing/pkg/apis/eventing/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const (
	EVENT_REAL_TO_VIRTUAL        string = "ktwin.real.virtual.generated"
	EVENT_VIRTUAL_TO_REAL        string = "ktwin.virtual.real.generated"
	EVENT_REAL_TO_EVENT_STORE    string = "ktwin.real.store.generated"
	EVENT_VIRTUAL_TO_EVENT_STORE string = "ktwin.real.store.generated"
	EVENT_VIRTUAL_TO_VIRTUAL     string = "ktwin.virtual.virtual.generated"
)

func NewTwinEvent() TwinEvent {
	return &twinEvent{}
}

type TwinEvent interface {
	GetTriggers(twinInstance *dtdv0.TwinInstance) []kEventing.Trigger
}

type twinEvent struct{}

type triggerParameters struct {
	triggerName string
	namespace   string
	brokerName  string
	eventType   string
	eventSource string
	subscriber  string
}

func (e *twinEvent) GetTriggers(twinInstance *dtdv0.TwinInstance) []kEventing.Trigger {

	var twinTriggers []kEventing.Trigger
	var trigger kEventing.Trigger

	// Real to Virtual
	trigger = e.createTrigger(triggerParameters{
		triggerName: twinInstance.Name + "-real-to-virtual",
		namespace:   twinInstance.Namespace,
		brokerName:  broker.EVENT_BROKER_NAME,
		eventType:   EVENT_REAL_TO_VIRTUAL,
		eventSource: twinInstance.Name,
		subscriber:  twinInstance.Name,
	})
	twinTriggers = append(twinTriggers, trigger)

	// Virtual to Real
	trigger = e.createTrigger(triggerParameters{
		triggerName: twinInstance.Name + "-virtual-to-real",
		namespace:   twinInstance.Namespace,
		brokerName:  broker.EVENT_BROKER_NAME,
		eventType:   EVENT_VIRTUAL_TO_REAL,
		eventSource: twinInstance.Name,
		subscriber:  twinInstance.Name + "-mqtt",
	})
	twinTriggers = append(twinTriggers, trigger)

	// Real to Event Store
	trigger = e.createTrigger(triggerParameters{
		triggerName: twinInstance.Name + "-real-to-event-store",
		namespace:   twinInstance.Namespace,
		brokerName:  broker.EVENT_BROKER_NAME,
		eventType:   EVENT_REAL_TO_EVENT_STORE,
		eventSource: twinInstance.Name,
		subscriber:  eventStore.EVENT_STORE_SERVICE,
	})
	twinTriggers = append(twinTriggers, trigger)

	// Virtual to Event Store
	trigger = e.createTrigger(triggerParameters{
		triggerName: twinInstance.Name + "-virtual-to-virtual",
		namespace:   twinInstance.Namespace,
		brokerName:  broker.EVENT_BROKER_NAME,
		eventType:   EVENT_VIRTUAL_TO_EVENT_STORE,
		eventSource: twinInstance.Name,
		subscriber:  eventStore.EVENT_STORE_SERVICE,
	})
	twinTriggers = append(twinTriggers, trigger)

	// Virtual to virtual
	// TODO

	return twinTriggers
}

func (*twinEvent) createTrigger(triggerParameters triggerParameters) kEventing.Trigger {
	return kEventing.Trigger{
		TypeMeta: v1.TypeMeta{
			Kind:       "Trigger",
			APIVersion: "eventing.knative.dev/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      triggerParameters.triggerName,
			Namespace: triggerParameters.namespace,
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
