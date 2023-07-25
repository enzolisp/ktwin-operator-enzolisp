package service

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kserving "knative.dev/serving/pkg/apis/serving/v1"

	dtdv0 "ktwin/operator/api/dtd/v0"
)

func NewTwinService() TwinService {
	return &twinService{}
}

type TwinService interface {
	GetService(twinInstance *dtdv0.TwinInstance) *kserving.Service
	GetDeletionService(namespacedName types.NamespacedName) *kserving.Service
}

type twinService struct{}

func (*twinService) GetService(twinInstance *dtdv0.TwinInstance) *kserving.Service {
	twinInstanceName := twinInstance.ObjectMeta.Name
	podSpec := twinInstance.Spec.Template.Spec

	service := &kserving.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      twinInstance.ObjectMeta.Name,
			Namespace: twinInstance.ObjectMeta.Namespace,
		},
		Spec: kserving.ServiceSpec{
			ConfigurationSpec: kserving.ConfigurationSpec{
				Template: kserving.RevisionTemplateSpec{
					ObjectMeta: v1.ObjectMeta{
						Name: twinInstanceName + "-v1",
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

func (*twinService) GetDeletionService(namespacedName types.NamespacedName) *kserving.Service {
	return &kserving.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
	}
}
