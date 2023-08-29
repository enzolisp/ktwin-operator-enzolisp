package service

import (
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	keventing "knative.dev/eventing/pkg/apis/eventing/v1"
	kserving "knative.dev/serving/pkg/apis/serving/v1"

	dtdv0 "ktwin/operator/api/dtd/v0"
)

type TwinServiceParameters struct {
	TwinInterface *dtdv0.TwinInterface
	Broker        keventing.Broker
	Service       kserving.Service
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
	eventStoreUrl := twinServiceParameters.Service.Status.URL.URL()

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
			},
		})
	}

	return containers
}

func (t *twinService) GetService(twinServiceParameters TwinServiceParameters) *kserving.Service {
	twinInterface := twinServiceParameters.TwinInterface
	twinInterfaceName := twinInterface.ObjectMeta.Name
	containers := t.getTwinInterfaceContainers(twinServiceParameters)

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
						Name: twinInterfaceName + "-v1",
						Annotations: map[string]string{
							"autoscaling.knative.dev/target":   "1",
							"autoscaling.knative.dev/maxScale": "1",
						},
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
