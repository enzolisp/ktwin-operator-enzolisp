package event

import (
	"fmt"
	"strconv"
	"strings"

	dtdv0 "github.com/Open-Digital-Twin/ktwin-operator/api/dtd/v0"
	"github.com/Open-Digital-Twin/ktwin-operator/pkg/third-party/rabbitmq"

	rabbitmqv1beta1 "github.com/rabbitmq/messaging-topology-operator/api/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kEventing "knative.dev/eventing/pkg/apis/eventing/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func NewTwinEvent() TwinEvent {
	return &twinEvent{}
}

type TwinEvent interface {
	GetTwinInterfaceTrigger(twinInterface *dtdv0.TwinInterface) *kEventing.Trigger
	GetTwinInterfaceCommandTriggers(twinInterface *dtdv0.TwinInterface) []kEventing.Trigger
	GetVirtualCloudEventBrokerBinding(twinInterface *dtdv0.TwinInterface, brokerExchange rabbitmqv1beta1.Exchange) []rabbitmqv1beta1.Binding
	GetRelationshipBrokerBindings(twinInterface *dtdv0.TwinInterface, brokerExchange rabbitmqv1beta1.Exchange, twinInterfaceQueue rabbitmqv1beta1.Queue) []rabbitmqv1beta1.Binding
	GetMQQTDispatcherBindings(twinInterface *dtdv0.TwinInterface) []rabbitmqv1beta1.Binding
}

type twinEvent struct{}

type TriggerParameters struct {
	InterfaceName  string
	TriggerName    string
	Namespace      string
	BrokerName     string
	EventType      string
	Subscriber     string
	OwnerReference []v1.OwnerReference
	Annotations    map[string]string
}

func (e *twinEvent) getEventTypeRealGenerated(twinInterfaceName string) string {
	return fmt.Sprintf(EVENT_TYPE_REAL_GENERATED, twinInterfaceName)
}

func (e *twinEvent) getEventTypeVirtualGenerated(twinInterfaceName string) string {
	return fmt.Sprintf(EVENT_TYPE_VIRTUAL_GENERATED, twinInterfaceName)
}

func (e *twinEvent) getEventTypeCommandExecuted(twinInterfaceName string, commandName string) string {
	return fmt.Sprintf(EVENT_TYPE_COMMAND_EXECUTED, twinInterfaceName, commandName)
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

func (e *twinEvent) GetMQQTDispatcherBindings(
	twinInterface *dtdv0.TwinInterface,
) []rabbitmqv1beta1.Binding {
	var rabbitMQBindings []rabbitmqv1beta1.Binding

	rabbitMQVirtualBinding, _ := rabbitmq.NewBinding(rabbitmq.BindingArgs{
		Name:      strings.ToLower(twinInterface.Name) + "-real-mqtt-dispatcher",
		Namespace: twinInterface.Namespace,
		Owner: []v1.OwnerReference{
			{
				APIVersion: twinInterface.APIVersion,
				Kind:       twinInterface.Kind,
				Name:       twinInterface.Name,
				UID:        twinInterface.UID,
			},
		},
		RabbitmqClusterReference: &rabbitmqv1beta1.RabbitmqClusterReference{
			Name:      "rabbitmq",
			Namespace: "ktwin",
		},
		RabbitMQVhost: RABBITMQ_VHOST,
		Source:        MQTT_EXCHANGE,
		Destination:   MQTT_DISPATCHER_QUEUE,
		Labels: map[string]string{
			"ktwin/twin-interface":         twinInterface.Name,
			"eventing.knative.dev/trigger": twinInterface.Name,
		},
		RoutingKey: e.getEventTypeRealGenerated(twinInterface.Name + ".#"),
	})

	rabbitMQBindings = append(rabbitMQBindings, rabbitMQVirtualBinding)

	for _, twinInterfaceRelationship := range twinInterface.Spec.Relationships {
		if twinInterfaceRelationship.AggregateData {
			rabbitMQVirtualBinding, _ := rabbitmq.NewBinding(rabbitmq.BindingArgs{
				Name:      strings.ToLower(twinInterface.Name) + "-" + strings.ToLower(twinInterfaceRelationship.Name) + "-real-mqtt-dispatcher",
				Namespace: twinInterface.Namespace,
				Owner: []v1.OwnerReference{
					{
						APIVersion: twinInterface.APIVersion,
						Kind:       twinInterface.Kind,
						Name:       twinInterface.Name,
						UID:        twinInterface.UID,
					},
				},
				RabbitmqClusterReference: &rabbitmqv1beta1.RabbitmqClusterReference{
					Name:      "rabbitmq",
					Namespace: "ktwin",
				},
				RabbitMQVhost: RABBITMQ_VHOST,
				Source:        MQTT_EXCHANGE,
				Destination:   MQTT_DISPATCHER_QUEUE,
				Labels: map[string]string{
					"ktwin/twin-interface":         twinInterface.Name,
					"eventing.knative.dev/trigger": twinInterface.Name,
				},
				RoutingKey: e.getEventTypeRealGenerated(twinInterfaceRelationship.Interface + ".#"),
			})
			rabbitMQBindings = append(rabbitMQBindings, rabbitMQVirtualBinding)
		}
	}

	return rabbitMQBindings
}

func (e *twinEvent) GetVirtualCloudEventBrokerBinding(
	twinInterface *dtdv0.TwinInterface,
	brokerExchange rabbitmqv1beta1.Exchange,
) []rabbitmqv1beta1.Binding {
	rabbitMQBindings := []rabbitmqv1beta1.Binding{}
	virtualEventBinding, _ := rabbitmq.NewBinding(rabbitmq.BindingArgs{
		Name:      strings.ToLower(twinInterface.Name) + "-virtual-cloud-event-dispatcher",
		Namespace: twinInterface.Namespace,
		Labels: map[string]string{
			"ktwin/twin-interface":         twinInterface.Name,
			"eventing.knative.dev/trigger": twinInterface.Name,
		},
		Filters: map[string]string{
			"type":                         e.getEventTypeVirtualGenerated(twinInterface.Name),
			"x-knative-trigger":            twinInterface.Name,
			"x-match":                      "all",
			"ktwin/twin-interface":         twinInterface.Name,
			"eventing.knative.dev/trigger": twinInterface.Name,
		},
		RabbitMQVhost: "/",
		Owner: []v1.OwnerReference{
			{
				APIVersion: twinInterface.APIVersion,
				Kind:       twinInterface.Kind,
				Name:       twinInterface.Name,
				UID:        twinInterface.UID,
			},
		},
		RabbitmqClusterReference: &rabbitmqv1beta1.RabbitmqClusterReference{
			Name:      "rabbitmq",
			Namespace: "ktwin",
		},
		Source:      brokerExchange.Spec.Name,     // broker exchange
		Destination: CLOUD_EVENT_DISPATCHER_QUEUE, // trigger queue
	})

	rabbitMQBindings = append(rabbitMQBindings, virtualEventBinding)

	return rabbitMQBindings
}

func (e *twinEvent) GetRelationshipBrokerBindings(
	twinInterface *dtdv0.TwinInterface,
	brokerExchange rabbitmqv1beta1.Exchange,
	twinInterfaceQueue rabbitmqv1beta1.Queue,
) []rabbitmqv1beta1.Binding {
	rabbitMQBindings := []rabbitmqv1beta1.Binding{}
	for _, twinInterfaceRelationship := range twinInterface.Spec.Relationships {
		if twinInterfaceRelationship.AggregateData {
			realEventBinding, _ := rabbitmq.NewBinding(rabbitmq.BindingArgs{
				Name:      strings.ToLower(twinInterface.Name) + "-" + strings.ToLower(twinInterfaceRelationship.Name) + "-real-dispatcher",
				Namespace: twinInterface.Namespace,
				Labels: map[string]string{
					"ktwin/twin-interface":         twinInterface.Name,
					"eventing.knative.dev/trigger": twinInterface.Name,
				},
				Filters: map[string]string{
					"type":              e.getEventTypeRealGenerated(twinInterfaceRelationship.Interface),
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
				},
				RabbitmqClusterReference: &rabbitmqv1beta1.RabbitmqClusterReference{
					Name:      "rabbitmq",
					Namespace: "ktwin",
				},
				Source:      brokerExchange.Spec.Name,     // broker exchange
				Destination: twinInterfaceQueue.Spec.Name, // trigger queue
			})

			virtualEventBinding, _ := rabbitmq.NewBinding(rabbitmq.BindingArgs{
				Name:      strings.ToLower(twinInterface.Name) + "-" + strings.ToLower(twinInterfaceRelationship.Name) + "-virtual-dispatcher",
				Namespace: twinInterface.Namespace,
				Labels: map[string]string{
					"ktwin/twin-interface":         twinInterface.Name,
					"eventing.knative.dev/trigger": twinInterface.Name,
				},
				Filters: map[string]string{
					"type":              e.getEventTypeVirtualGenerated(twinInterfaceRelationship.Interface),
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
				},
				RabbitmqClusterReference: &rabbitmqv1beta1.RabbitmqClusterReference{
					Name:      "rabbitmq",
					Namespace: "ktwin",
				},
				Source:      brokerExchange.Spec.Name,     // broker exchange
				Destination: CLOUD_EVENT_DISPATCHER_QUEUE, // trigger queue
			})

			rabbitMQBindings = append(rabbitMQBindings, virtualEventBinding)
			rabbitMQBindings = append(rabbitMQBindings, realEventBinding)
		}
	}

	return rabbitMQBindings
}

func (e *twinEvent) GetTwinInterfaceTrigger(twinInterface *dtdv0.TwinInterface) *kEventing.Trigger {
	var twinInterfaceTrigger *kEventing.Trigger

	virtualTwinService := twinInterface.Name

	// If TwinInstance has container associated, create the triggers
	if e.hasContainerInTwinInterface(twinInterface) {
		// Real Twin Event Type
		twinInterfaceEventType := e.getEventTypeRealGenerated(twinInterface.Name)
		var triggerAnnotations = make(map[string]string)

		if twinInterface.Spec.Service != nil && twinInterface.Spec.Service.AutoScaling.Parallelism != nil {
			triggerAnnotations["rabbitmq.eventing.knative.dev/parallelism"] = strconv.Itoa(*twinInterface.Spec.Service.AutoScaling.Parallelism)
		}

		twinInterfaceTrigger = e.createTrigger(TriggerParameters{
			TriggerName:   e.getTwinInterfaceTrigger(twinInterface.Name),
			Namespace:     twinInterface.Namespace,
			BrokerName:    EVENT_BROKER_NAME,
			EventType:     twinInterfaceEventType,
			Subscriber:    virtualTwinService,
			InterfaceName: twinInterface.Name,
			OwnerReference: []v1.OwnerReference{
				{
					APIVersion: twinInterface.APIVersion,
					Kind:       twinInterface.Kind,
					Name:       twinInterface.Name,
					UID:        twinInterface.UID,
				},
			},
			Annotations: triggerAnnotations,
		})

	}

	return twinInterfaceTrigger
}

func (e *twinEvent) GetTwinInterfaceCommandTriggers(twinInterface *dtdv0.TwinInterface) []kEventing.Trigger {
	var twinInterfaceTriggers []kEventing.Trigger

	virtualTwinService := twinInterface.Name

	// If TwinInstance has container associated, create the triggers
	if e.hasContainerInTwinInterface(twinInterface) {
		// Real Twin Event Type
		for _, command := range twinInterface.Spec.Commands {
			twinInterfaceCommandEventType := e.getEventTypeCommandExecuted(twinInterface.Name, command.Name)
			var triggerAnnotations = make(map[string]string)

			if twinInterface.Spec.Service != nil && twinInterface.Spec.Service.AutoScaling.Parallelism != nil {
				triggerAnnotations["rabbitmq.eventing.knative.dev/parallelism"] = strconv.Itoa(*twinInterface.Spec.Service.AutoScaling.Parallelism)
			}

			twinInterfaceTrigger := e.createTrigger(TriggerParameters{
				TriggerName:   e.getTwinInterfaceTrigger(twinInterface.Name + "-" + command.Name),
				Namespace:     twinInterface.Namespace,
				BrokerName:    EVENT_BROKER_NAME,
				EventType:     twinInterfaceCommandEventType,
				Subscriber:    virtualTwinService,
				InterfaceName: twinInterface.Name,
				OwnerReference: []v1.OwnerReference{
					{
						APIVersion: twinInterface.APIVersion,
						Kind:       twinInterface.Kind,
						Name:       twinInterface.Name,
						UID:        twinInterface.UID,
					},
				},
				Annotations: triggerAnnotations,
			})

			twinInterfaceTriggers = append(twinInterfaceTriggers, *twinInterfaceTrigger)
		}

	}

	return twinInterfaceTriggers
}

func (e *twinEvent) createTrigger(triggerParameters TriggerParameters) *kEventing.Trigger {
	return &kEventing.Trigger{
		TypeMeta: v1.TypeMeta{
			Kind:       "Trigger",
			APIVersion: "eventing.knative.dev/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:            triggerParameters.TriggerName,
			Namespace:       triggerParameters.Namespace,
			Labels:          e.getTriggerLabels(triggerParameters.InterfaceName),
			OwnerReferences: triggerParameters.OwnerReference,
			Annotations:     triggerParameters.Annotations,
		},
		Spec: kEventing.TriggerSpec{
			Broker: triggerParameters.BrokerName,
			Filter: &kEventing.TriggerFilter{
				Attributes: map[string]string{
					"type": triggerParameters.EventType,
				},
			},
			Subscriber: duckv1.Destination{
				Ref: &duckv1.KReference{
					Kind:       "Service",
					APIVersion: "serving.knative.dev/v1",
					Name:       triggerParameters.Subscriber,
				},
			},
		},
	}
}

func (*twinEvent) hasContainerInTwinInterface(twinInterface *dtdv0.TwinInterface) bool {
	return twinInterface.Spec.Service != nil
}
