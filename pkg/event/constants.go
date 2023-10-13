package event

const (
	// Cloud Event Types
	EVENT_TYPE_REAL_GENERATED    string = "ktwin.real.%s"
	EVENT_TYPE_VIRTUAL_GENERATED string = "ktwin.virtual.%s"
	EVENT_VIRTUAL_TO_VIRTUAL     string = "ktwin.virtual.virtual" // TODO: what if someone wants to send an event to a relationship (post processing - use command)
)

const (
	// MQTT Dispatchers name
	CLOUD_EVENT_DISPATCHER = "cloud-event-dispatcher"
	MQTT_DISPATCHER        = "mqtt-dispatcher"

	// MQTT Dispatcher Queues
	CLOUD_EVENT_DISPATCHER_QUEUE string = "cloud-event-dispatcher-queue"
	MQTT_DISPATCHER_QUEUE        string = "mqtt-dispatcher-queue"
)

const (
	EVENT_BROKER_NAME               string = "ktwin"
	RABBITMQ_VHOST                  string = "/"
	CLOUD_EVENT_DISPATCHER_EXCHANGE string = "amq.topic"
)
