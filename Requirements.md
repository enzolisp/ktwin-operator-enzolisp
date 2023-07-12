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
    - How to send the response back? It is the user responsability to post the message in the broker with the right event type header.
      - How to convert message response -> Broker - MQTT/AMQP?

- UI to view components created.
