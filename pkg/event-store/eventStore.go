package eventStore

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev0 "ktwin/operator/api/core/v0"
	dtdv0 "ktwin/operator/api/dtd/v0"
	"ktwin/operator/pkg/naming"
	knative "ktwin/operator/pkg/third-party/knative"
	"ktwin/operator/pkg/third-party/rabbitmq"

	"github.com/google/uuid"
	rabbitmqv1beta1 "github.com/rabbitmq/messaging-topology-operator/api/v1beta1"
	kEventing "knative.dev/eventing/pkg/apis/eventing/v1"
	kserving "knative.dev/serving/pkg/apis/serving/v1"
)

const (
	EVENT_STORE_SERVICE string = "event-store"
)

func NewEventStore() EventStore {
	return &eventStore{}
}

// TODO: create binding for mqtt and cloud event
// TODO: Implement creation of TwinInstance and TwinInterface in event store tables
type EventStore interface {
	GetEventStoreService(eventStore *corev0.EventStore) *kserving.Service
	GetEventStoreTrigger(eventStore *corev0.EventStore) kEventing.Trigger
	GetEventStoreBrokerBindings(twinInterface *dtdv0.TwinInterface, twinInterfaceTrigger kEventing.Trigger, brokerExchange rabbitmqv1beta1.Exchange, eventStoreQueue rabbitmqv1beta1.Queue) []rabbitmqv1beta1.Binding
}

type eventStore struct{}

func (t *eventStore) GetEventStoreService(eventStore *corev0.EventStore) *kserving.Service {
	eventStoreName := eventStore.ObjectMeta.Name

	service := &kserving.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      eventStore.ObjectMeta.Name,
			Namespace: eventStore.ObjectMeta.Namespace,
			Labels: map[string]string{
				"ktwin/event-store": eventStoreName,
			},
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: eventStore.APIVersion,
					Kind:       eventStore.Kind,
					Name:       eventStore.ObjectMeta.Name,
					UID:        eventStore.UID,
				},
			},
		},
		Spec: kserving.ServiceSpec{
			ConfigurationSpec: kserving.ConfigurationSpec{
				Template: kserving.RevisionTemplateSpec{
					ObjectMeta: v1.ObjectMeta{
						Name: eventStoreName + "-v1",
						Annotations: map[string]string{
							"autoscaling.knative.dev/target":   "2",
							"autoscaling.knative.dev/minScale": "1",
							"autoscaling.knative.dev/maxScale": "10",
						},
					},
					Spec: kserving.RevisionSpec{
						PodSpec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:            "ktwin-event-store",
									Image:           "dev.local/ktwin/" + "event-store" + ":0.1",
									ImagePullPolicy: corev1.PullIfNotPresent,
									Env: []corev1.EnvVar{
										{
											Name:  "DB_HOST",
											Value: "scylla-client.scylla.svc.cluster.local",
										},
										{
											Name:  "DB_PASSWORD",
											Value: "",
										},
										{
											Name:  "DB_KEYSPACE",
											Value: "ktwin",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return service
}

func (t *eventStore) GetEventStoreTrigger(eventStore *corev0.EventStore) kEventing.Trigger {
	return knative.NewTrigger(knative.TriggerParameters{
		TriggerName:    eventStore.Name + "-trigger",
		Namespace:      eventStore.Namespace,
		BrokerName:     "ktwin",
		SubscriberName: "event-store",
		OwnerReferences: []v1.OwnerReference{
			{
				APIVersion: eventStore.APIVersion,
				Kind:       eventStore.Kind,
				Name:       eventStore.Name,
				UID:        eventStore.UID,
			},
		},
		Attributes: map[string]string{
			"type": "ktwin.event-store",
		},
		Labels: map[string]string{
			"ktwin/event-store": "event-store",
		},
	})
}

func (t *eventStore) GetEventStoreBrokerBindings(twinInterface *dtdv0.TwinInterface, twinInterfaceTrigger kEventing.Trigger, brokerExchange rabbitmqv1beta1.Exchange, eventStoreQueue rabbitmqv1beta1.Queue) []rabbitmqv1beta1.Binding {
	var eventStoreBindings []rabbitmqv1beta1.Binding

	for _, relationship := range twinInterface.Spec.Relationships {
		realEventBinding, _ := rabbitmq.NewBinding(rabbitmq.BindingArgs{
			Name:      strings.ToLower(relationship.Name) + "-real-event-store-" + uuid.NewString(),
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
			RabbitMQVhost: "/",
			Source:        brokerExchange.Spec.Name,
			Destination:   eventStoreQueue.Spec.Name,
			Filters: map[string]string{
				"type":              naming.GetEventTypeRealGenerated(twinInterface.Name),
				"x-knative-trigger": "event-store-trigger",
				"x-match":           "all",
			},
			Labels: map[string]string{},
		})

		virtualEventBinding, _ := rabbitmq.NewBinding(rabbitmq.BindingArgs{
			Name:      strings.ToLower(relationship.Name) + "-virtual-event-store-" + uuid.NewString(),
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
			RabbitMQVhost: "/",
			Source:        brokerExchange.Spec.Name,
			Destination:   eventStoreQueue.Spec.Name,
			Filters: map[string]string{
				"type":              naming.GetEventTypeVirtualGenerated(twinInterface.Name),
				"x-knative-trigger": "event-store-trigger",
				"x-match":           "all",
			},
			Labels: map[string]string{},
		})

		eventStoreBindings = append(eventStoreBindings, realEventBinding)
		eventStoreBindings = append(eventStoreBindings, virtualEventBinding)
	}

	return eventStoreBindings
}
