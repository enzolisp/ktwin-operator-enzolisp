package integrator

import (
	"fmt"

	camelv1 "github.com/apache/camel-k/pkg/apis/camel/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	dtdv0 "ktwin/operator/api/dtd/v0"
)

const (
	MQTT_BROKER_URL string = "tcp://mqtt-broker:1883"
)

func NewTwinIntegrator() TwinIntegrator {
	return &twinIntegrator{}
}

// Generate MQTT Camel-k Integrators
// - Type: `ktwin.real.<instance_id>.generated`: Event generated from Real Twin and sent to Virtual Twin.
// - `ktwin.virtual.<instance_id>.generated`: Event generated from Virtual Twin and sent to Real Twin.
type TwinIntegrator interface {
	GetIntegrators(twinInstance *dtdv0.TwinInstance) *[]camelv1.Integration
	GetDeletionIntegrator(namespacedName types.NamespacedName) *[]camelv1.Integration
}

type twinIntegrator struct{}

func (*twinIntegrator) getVirtualToRealIntegratorSource(twinInstance *dtdv0.TwinInstance) string {
	integratorName := twinInstance.Name + "-mqtt-to-real"
	mqttTopic := integratorName + "-to-real"

	content := fmt.Sprintf(`from("knative:endpoint/%s")`, integratorName)
	content += `.to("direct:mqtt-response-handler");`
	content += `from("direct:mqtt-response-handler")`
	content += `.to("log:info?showAll=true&multiline=true")`
	content += fmt.Sprintf(`.to("paho:%s?brokerUrl=%s")`, mqttTopic, MQTT_BROKER_URL)
	return content
}

func (*twinIntegrator) getRealToVirtualIntegratorSource(twinInstance *dtdv0.TwinInstance) string {
	integratorName := twinInstance.Name
	mqttTopic := integratorName + "-to-virtual"

	ceType := "ktwin.real." + integratorName + ".generated"
	// ceSource := "CE-Source"

	content := fmt.Sprintf(`from("paho:%s?brokerUrl=%s")`, mqttTopic, MQTT_BROKER_URL)
	content += `.to("log:info")`
	content += fmt.Sprintf(`.setHeader("ce-type").constant("%s")`, ceType)
	// content += fmt.Sprintf(`.setHeader("ce-source").constant("%s")`, ceSource)
	content += fmt.Sprintf(`.to("knative:event")`)

	return content
}

func (t *twinIntegrator) getRealToVirtualIntegratorName(twinInstanceName string) string {
	return twinInstanceName + "-int-real-virtual"
}

func (t *twinIntegrator) getRealToVirtualIntegrator(twinInstance *dtdv0.TwinInstance) camelv1.Integration {
	// TODO: Truncate 63 characters
	integratorName := t.getRealToVirtualIntegratorName(twinInstance.Name)
	return camelv1.Integration{
		TypeMeta: v1.TypeMeta{
			APIVersion: "camel.apache.org/v1",
			Kind:       "Integration",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      integratorName,
			Namespace: "default",
		},
		Spec: camelv1.IntegrationSpec{
			Sources: []camelv1.SourceSpec{
				camelv1.NewSourceSpec(
					integratorName+".kts",
					t.getRealToVirtualIntegratorSource(twinInstance),
					camelv1.LanguageKotlin,
				),
			},
		},
	}
}

func (t *twinIntegrator) getVirtualToRealIntegratorName(twinInstanceName string) string {
	return twinInstanceName + "-int-virtual-real"
}

func (t *twinIntegrator) getVirtualToRealIntegrator(twinInstance *dtdv0.TwinInstance) camelv1.Integration {
	// TODO: Truncate 63 characters
	integratorName := t.getVirtualToRealIntegratorName(twinInstance.Name)
	return camelv1.Integration{
		TypeMeta: v1.TypeMeta{
			APIVersion: "camel.apache.org/v1",
			Kind:       "Integration",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      integratorName,
			Namespace: "default",
		},
		Spec: camelv1.IntegrationSpec{
			Sources: []camelv1.SourceSpec{
				camelv1.NewSourceSpec(
					integratorName+".kts",
					t.getVirtualToRealIntegratorSource(twinInstance),
					camelv1.LanguageKotlin,
				),
			},
		},
	}
}

func (t *twinIntegrator) GetIntegrators(twinInstance *dtdv0.TwinInstance) *[]camelv1.Integration {
	var mqttIntegrators []camelv1.Integration

	virtualToRealIntegrator := t.getVirtualToRealIntegrator(twinInstance)
	realToVirtualIntegrator := t.getRealToVirtualIntegrator(twinInstance)

	mqttIntegrators = append(mqttIntegrators, virtualToRealIntegrator)
	mqttIntegrators = append(mqttIntegrators, realToVirtualIntegrator)

	return &mqttIntegrators
}

func (t *twinIntegrator) GetDeletionIntegrator(namespacedName types.NamespacedName) *[]camelv1.Integration {
	var integrators []camelv1.Integration

	realToVirtualIntegrator := camelv1.Integration{
		TypeMeta: v1.TypeMeta{
			APIVersion: "camel.apache.org/v1",
			Kind:       "Integration",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      t.getRealToVirtualIntegratorName(namespacedName.Name),
			Namespace: namespacedName.Namespace,
		},
	}
	virtualToRealIntegrator := camelv1.Integration{
		TypeMeta: v1.TypeMeta{
			APIVersion: "camel.apache.org/v1",
			Kind:       "Integration",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      t.getVirtualToRealIntegratorName(namespacedName.Name),
			Namespace: namespacedName.Namespace,
		},
	}

	integrators = append(integrators, realToVirtualIntegrator)
	integrators = append(integrators, virtualToRealIntegrator)

	return &integrators
}
