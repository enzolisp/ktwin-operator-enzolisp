package service

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kserving "knative.dev/serving/pkg/apis/serving/v1"

	dtdv0 "ktwin/operator/api/dtd/v0"
)

const (
	EVENT_REAL_TO_VIRTUAL        string = "ktwin.real.virtual.generated"
	EVENT_VIRTUAL_TO_REAL        string = "ktwin.virtual.real.generated"
	EVENT_REAL_TO_EVENT_STORE    string = "ktwin.real.store.generated"
	EVENT_VIRTUAL_TO_EVENT_STORE string = "ktwin.real.store.generated"
	EVENT_VIRTUAL_TO_VIRTUAL     string = "ktwin.virtual.virtual.generated"
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
	serviceId := twinInstance.Spec.Id
	podSpec := twinInstance.Spec.Template.Spec
	objectData := twinInstance.ObjectMeta

	service := &kserving.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: objectData,
		Spec: kserving.ServiceSpec{
			ConfigurationSpec: kserving.ConfigurationSpec{
				Template: kserving.RevisionTemplateSpec{
					ObjectMeta: v1.ObjectMeta{
						Name: serviceId + "-v1",
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
