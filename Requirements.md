# Requirements

- Support HTTP, AMQP and MQTT as protocol for external services.
  - For every twin instance created, three endpoints must be available:
    - HTTP: Expose Broker URL (Pub/Sub) + Use Cloud Event Spec. How HTTP endpoint reply works (webhook?).
    - AMQP: URL + Request and Response topics. How to use existing RabbitMQ? How to build underlying layer?
    - MQTT: URL + Request and Response topics. Currently uses Mosquitto + Camel. Try with RabbitMQ or build a bridge between Mosquitto and RabbitMQ to remove Camel-k.

- Example:
  - MQTT Endpoint:
    - url: mosquitto:1883
    - publisherTopic: twin-topic
    - subscriberTopic: twin2-topic
  - AMQP Endpoint:
    - url: rabbitmq url
    - publisherTopic: twin-topic
    - subscriberTopic: twin2-topic
  - HTTP Endpoint (one way only for now):
    - url: default broker URL
    - Cloud Event (message type header)

- Event Routing between Real->Virtual, Virtual->Real, Real->Storage, Virtual->Storage, Virtual->Virtual.
  - How user defines if the event must be stored or not? Flag in yaml?

- When creating TwinInstance, create table in Store Service for that TwinInstance and store the messages to query and process.
  - How messages are routed to this URL? User is defined a flag in the YAML. By default it is set to true, if not specified. Create two triggers with the same filter.

- User container scope: How to improve the development experience?
  - The request received will always be a HTTP request.
  - User can define a container or a lambda in knative lambda.
  - Service does not need to know if the request came from AMQP, MQTT or HTTP endpoint. If message came from AMQP, how to reply only to AMQP endpoint?
  - Use cases:
    - Application receives a synchronous or asynchronous request
      - External Entity (asynchronous only)
      - Virtual Entity (Command): synchronous or asynchronous request from another TwinInstance that has a relationship with.
        - Why would a Virtual Entity call another Virtual Entity? To get its last states? It can call event-store for this.
        - URLs of relationships can be added to the environment variables of the container (this can be done by the container).
        - Configure initial state of instances
    - How to send the response back? It is the user responsability to post the message in the broker with the right event type header.
      - How to convert message response -> Broker - MQTT/AMQP?

- Functions:
func deploy --remote --git-url=https://github.com/agwermann/pole-function

Must install telko

- UI to view components created.

Poste tem containers. Agrega o dado dos sensores.
Qualidade do ar da região.

Menor quantidade de containers.
- Function (permite escalar melhor)

Definir uma cidade com ruas, postes e estacionamentos.

Any Trigger:
https://github.com/knative/eventing/blob/ffa591593417f3e879a4834ddb87cc13e6cd3e05/pkg/apis/eventing/v1/trigger_types.go#L135

- Sensores que variam com o tempo (bursts). Sensores de estacionamentos. Eficiência (criar menos pods).
- Como diminuir número triggers.
- Reduzir número de pods containers e triggers.
  - Receber mensagens em containers mais alto nível.
  - 1-1 Containers: 1 container (N postes)

- Mover imagem para TwinInterface
- Image precisa identificar qual é o poste que a mensagem
