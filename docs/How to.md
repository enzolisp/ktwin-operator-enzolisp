# How to tips

Get auto generated user and password of RabbitMQ admin area.

```sh
kubectl describe secret rabbitmq-default-user -n ktwin
kubectl get secrets/rabbitmq-default-user -n ktwin --template={{.data.host}} | base64 -D
kubectl get secrets/rabbitmq-default-user -n ktwin --template={{.data.port}} | base64 -D
kubectl get secrets/rabbitmq-default-user -n ktwin --template={{.data.username}} | base64 -D
kubectl get secrets/rabbitmq-default-user -n ktwin --template={{.data.password}} | base64 -D
```

Run the following command to expose RabbitMQ cluster Admin area:

```sh
kubectl port-forward -n ktwin --address 0.0.0.0 svc/rabbitmq 15672:15672
```

Access the following URL and login with the credentials previously generated: http://localhost:15672

Run the following command to expose MQTT port:

```sh
kubectl port-forward -n ktwin --address 0.0.0.0 svc/rabbitmq 1883:1883
```

Run the following command to expose AMQP port:

```sh
kubectl port-forward -n ktwin --address 0.0.0.0 svc/rabbitmq 5672:5672
```

## Enable Node Selector Feature Flag

Knative Services do not have support for Node Selector by default. You can enable the [kubernetes.podspec-nodeselector](https://knative.dev/docs/serving/configuration/feature-flags/#kubernetes-node-selector) feature flag.

```sh
kubectl edit configmap config-features -n knative-serving
```

## Label nodes for KTWIN workloads

```sh
kubectl label node mac-porvir-01 ktwin/core-node=true
kubectl label node mac-porvir-01 ktwin/service-node=true
kubectl label node mac-porvir-01 ktwin/device-node=true
```

## Delete stuck resources

```sh
kubectl get binding.rabbitmq.com <resource name> -o=json | \
jq '.metadata.finalizers = null' | kubectl apply -f -
```

```sh
kubectl get binding.rabbitmq.com -o=json | \
jq '.metadata.finalizers = null' | kubectl apply -f -
```
