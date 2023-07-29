package service

import (
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kserving "knative.dev/serving/pkg/apis/serving/v1"

	dtdv0 "ktwin/operator/api/dtd/v0"
)

func NewTwinService() TwinService {
	return &twinService{}
}

type TwinService interface {
	GetService(twinInterface *dtdv0.TwinInterface) *kserving.Service
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

func (t *twinService) GetService(twinInterface *dtdv0.TwinInterface) *kserving.Service {
	twinInterfaceName := twinInterface.ObjectMeta.Name
	podSpec := twinInterface.Spec.Template.Spec

	service := &kserving.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      twinInterface.ObjectMeta.Name,
			Namespace: twinInterface.ObjectMeta.Namespace,
			Labels:    t.getServiceLabels(twinInterfaceName),
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
							Containers: podSpec.Containers,
						},
					},
				},
			},
		},
	}
	return service
}
