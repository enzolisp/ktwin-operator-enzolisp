package mqtt

type TwinMqttIntegrator interface {
	CreateIntegrator()
}

type twinMqttIntegrator struct{}

func (*twinMqttIntegrator) CreateIntegrator() {}
