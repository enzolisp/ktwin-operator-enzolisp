# How to tips

Get auto generated user and password of RabbitMQ admin area.

```sh
kubectl describe secret rabbitmq-default-user
kubectl get secrets/rabbitmq-default-user --template={{.data.username}} | base64 -D
kubectl get secrets/rabbitmq-default-user --template={{.data.password}} | base64 -D
```
