package knative

import (
	"reflect"
	"strconv"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kEventing "knative.dev/eventing/pkg/apis/eventing/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

type TriggerParameters struct {
	TriggerName     string
	Namespace       string
	BrokerName      string
	SubscriberName  string
	OwnerReferences []v1.OwnerReference
	Attributes      map[string]string
	Labels          map[string]string
	URL             TriggerURLParameters
	Parallelism     *int
}

type TriggerURLParameters struct {
	Path string
}

func NewTrigger(triggerParameters TriggerParameters) *kEventing.Trigger {
	var triggerAnnotations = make(map[string]string)

	if triggerParameters.Parallelism != nil {
		triggerAnnotations["rabbitmq.eventing.knative.dev/parallelism"] = strconv.Itoa(*triggerParameters.Parallelism)
	}

	var urlParameters *apis.URL
	if !reflect.DeepEqual(triggerParameters.URL, TriggerURLParameters{}) {
		urlParameters = &apis.URL{
			Path: triggerParameters.URL.Path,
		}
	}

	return &kEventing.Trigger{
		TypeMeta: v1.TypeMeta{
			Kind:       "Trigger",
			APIVersion: "eventing.knative.dev/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:            triggerParameters.TriggerName,
			Namespace:       triggerParameters.Namespace,
			Labels:          triggerParameters.Labels,
			OwnerReferences: triggerParameters.OwnerReferences,
			Annotations:     triggerAnnotations,
		},
		Spec: kEventing.TriggerSpec{
			Broker: triggerParameters.BrokerName,
			Filter: &kEventing.TriggerFilter{
				Attributes: triggerParameters.Attributes,
			},
			Subscriber: duckv1.Destination{
				Ref: &duckv1.KReference{
					Kind:       "Service",
					APIVersion: "serving.knative.dev/v1",
					Name:       triggerParameters.SubscriberName,
				},
				URI: urlParameters,
			},
		},
	}
}
