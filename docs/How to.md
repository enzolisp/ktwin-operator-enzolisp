# How to tips

Get auto generated user and password of RabbitMQ admin area.

```sh
kubectl describe secret rabbitmq-default-user
kubectl get secrets/rabbitmq-default-user --template={{.data.username}} | base64 -D
kubectl get secrets/rabbitmq-default-user --template={{.data.password}} | base64 -D
```

Run the following command to expose RabbitMQ cluster Admin area:

```sh
kubectl port-forward --address 0.0.0.0 svc/rabbitmq 15672:15672
```

Access the following URL and login with the credentials previously generated: http://localhost:15672

Run the following command to expose MQTT port:

```sh
kubectl port-forward --address 0.0.0.0 svc/rabbitmq 1883:1883
```

Run the following command to expose AMQP port:

```sh
kubectl port-forward --address 0.0.0.0 svc/rabbitmq 5672:5672
```
