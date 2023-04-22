# Installation Steps

## Local Development

1. Configure your Kubernetes cluster. You can run the platform in [Kind](https://kind.sigs.k8s.io/) in your local computer.

```sh
kind create cluster
```

2. Load Docker image into cluster.

```sh
kind load docker-image dev.local/edge-service:0.1
```

3. Deploy ScillaDB for the Event Store.

```sh

```

4. Deploy Mosquitto Mqtt Broker.

```sh

```

3. Install Knative and Istio dependencies.

- Knative Serving:

```sh
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.8.0/serving-crds.yaml
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.8.0/serving-core.yaml
kubectl get pods --namespace knative-serving
```

- Knative Eventing:

```sh
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.8.0/eventing-crds.yaml
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.8.0/eventing-core.yaml
kubectl get pods --namespace knative-eventing
```

5. Install CDR in Cluster

```
make install
```

4. Install Camel-k

```sh
kubectl -n default create secret docker-registry external-registry-secret --docker-username <DOCKER_USERNAME> --docker-password <DOCKER_PASSWORD> -n core
kamel install --operator-image=docker.io/apache/camel-k:1.10.3 --olm=false -n core --global --registry docker.io --organization agwermann --registry-secret external-registry-secret --force
```