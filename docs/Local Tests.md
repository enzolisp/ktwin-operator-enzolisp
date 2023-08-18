# Scripts to execute local tests

## Pre-requisites

1. Create cluster.
2. Execute the installation scripts.
3. Load local docker images.
4. Install Ktwin resources.

## Steps

### Expose the MQTT Port

```sh
kubectl port-forward -n ktwin --address 0.0.0.0 svc/rabbitmq 1883:1883
```

### Get Broker credentials

```sh
kubectl get secrets/rabbitmq-default-user -n ktwin --template={{.data.username}} | base64 -D
kubectl get secrets/rabbitmq-default-user -n ktwin --template={{.data.password}} | base64 -D
```

### Connect as a MQTT subscriber

```sh
mosquitto_sub -h localhost -u USER -P PASSWORD -p 1883 -d -t ktwin/virtual/city-pole/city-pole-001
```

### Connect as a MQTT publisher

```sh
mosquitto_pub -h localhost -m "{"temperature":{"min":21,"max":29,"unit":"celsius"},"time":1568881230}" -u USER -P PASSWORD -p 1883 -d -t ktwin/real/city-pole/city-pole-001
```

```sh
mosquitto_pub -h localhost -m "{"temperature":{"min":21,"max":29,"unit":"celsius"},"time":1568881230}" -u USER -P PASSWORD -p 1883 -d -t ktwin/real/ngsi-ld-city-noiselevelobserved/ngsi-ld-city-noiselevelobserved-001
```

### Publish Event within the cluster

```sh
kubectl run curl \
    --image=curlimages/curl --rm=true --restart=Never -ti -- \
    -X POST -v \
    -H "content-type: application/json"  \
    -H "ce-specversion: 1.0"  \
    -H "ce-source: city-pole-001"  \
    -H "ce-type: ktwin.virtual.city-pole"  \
    -H "ce-id: 123-abc"  \
    -d '{"data":"response"}' \
    http://ktwin-broker-ingress.ktwin.svc.cluster.local
```

### Get Event Store data

```sh
kubectl run curl \
    --image=curlimages/curl --rm=true --restart=Never -ti -- \
    -X GET -v \
    http://event-store.ktwin.svc.cluster.local/api/v1/twin-events
```
