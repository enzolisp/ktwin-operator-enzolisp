package service

import (
	"reflect"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	keventing "knative.dev/eventing/pkg/apis/eventing/v1"
	kserving "knative.dev/serving/pkg/apis/serving/v1"

	dtdv0 "github.com/Open-Digital-Twin/ktwin-operator/api/dtd/v0"
)

// Used to inject settings as environment variables
type KtwinEnvironmentSettings struct {
	Relationships []KtwinRelationshipSettings `json:"relationships"`
	Parent        KtwinRelationshipSettings   `json:"parent"`
}

type KtwinRelationshipSettings struct {
	Name      string `json:"name"`
	Interface string `json:"interface"`
	Instance  string `json:"instance"`
}

type TwinServiceParameters struct {
	TwinInterface         *dtdv0.TwinInterface
	Broker                keventing.Broker
	EventStoreService     kserving.Service
	OperatorMetricService kserving.Service
}

func NewTwinService() TwinService {
	return &twinService{}
}

type TwinService interface {
	GetService(twinServiceParameters TwinServiceParameters) *kserving.Service
	GetServiceDeletionCriteria(namespacedName types.NamespacedName) map[string]string
}

type twinService struct{}

func (e *twinService) getServiceLabels(twinInterfaceName string) map[string]string {
	return map[string]string{
		"ktwin/twininterface": twinInterfaceName,
	}
}

func (e *twinService) GetServiceDeletionCriteria(namespacedName types.NamespacedName) map[string]string {
	return e.getServiceLabels(namespacedName.Name)
}

func (e *twinService) getTwinInterfaceContainers(twinServiceParameters TwinServiceParameters) []corev1.Container {
	var containers []corev1.Container

	brokerUrl := twinServiceParameters.Broker.Status.Address.URL.URL()
	eventStoreUrl := twinServiceParameters.EventStoreService.Status.URL.URL()
	operatorMetricUrl := twinServiceParameters.OperatorMetricService.Status.URL.URL()

	for _, container := range twinServiceParameters.TwinInterface.Spec.Service.Template.Spec.Containers {
		containers = append(containers, corev1.Container{
			Name:            container.Name,
			Image:           container.Image,
			ImagePullPolicy: container.ImagePullPolicy,
			Env: []corev1.EnvVar{
				{
					Name:  "KTWIN_BROKER",
					Value: brokerUrl.String(),
				},
				{
					Name:  "KTWIN_EVENT_STORE",
					Value: eventStoreUrl.String(),
				},
				{
					Name:  "KTWIN_GRAPH_URL",
					Value: operatorMetricUrl.String() + ":8443/twin-graph",
				},
			},
		})
	}

	return containers
}

func (t *twinService) GetService(twinServiceParameters TwinServiceParameters) *kserving.Service {
	twinInterface := twinServiceParameters.TwinInterface
	twinInterfaceName := twinInterface.ObjectMeta.Name
	containers := t.getTwinInterfaceContainers(twinServiceParameters)
	var autoScalingAnnotations map[string]string = make(map[string]string)

	if !reflect.DeepEqual(twinInterface.Spec.Service.AutoScaling, dtdv0.TwinInterfaceAutoScaling{}) {
		autoScaling := twinInterface.Spec.Service.AutoScaling
		autoScalingAnnotations = make(map[string]string)
		if autoScaling.MaxScale != nil {
			autoScalingAnnotations["autoscaling.knative.dev/maxScale"] = strconv.Itoa(*autoScaling.MaxScale)
		} else {
			autoScalingAnnotations["autoscaling.knative.dev/maxScale"] = strconv.Itoa(1)
		}

		if autoScaling.MinScale != nil {
			autoScalingAnnotations["autoscaling.knative.dev/minScale"] = strconv.Itoa(*autoScaling.MinScale)
		} else {
			autoScalingAnnotations["autoscaling.knative.dev/minScale"] = strconv.Itoa(0)
		}

		if autoScaling.Target != nil {
			autoScalingAnnotations["autoscaling.knative.dev/target"] = strconv.Itoa(*autoScaling.Target)
		}

		if autoScaling.Metric != "" {
			autoScalingAnnotations["autoscaling.knative.dev/metric"] = string(*&autoScaling.Metric)
		}
	}

	service := &kserving.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      twinInterface.ObjectMeta.Name,
			Namespace: twinInterface.ObjectMeta.Namespace,
			Labels:    t.getServiceLabels(twinInterfaceName),
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: twinInterface.APIVersion,
					Kind:       twinInterface.Kind,
					Name:       twinInterface.Name,
					UID:        twinInterface.UID,
				},
			},
		},
		Spec: kserving.ServiceSpec{
			ConfigurationSpec: kserving.ConfigurationSpec{
				Template: kserving.RevisionTemplateSpec{
					ObjectMeta: v1.ObjectMeta{
						Name:        twinInterfaceName + "-v1",
						Annotations: autoScalingAnnotations,
					},
					Spec: kserving.RevisionSpec{
						PodSpec: corev1.PodSpec{
							Containers: containers,
						},
					},
				},
			},
		},
	}
	return service
}
