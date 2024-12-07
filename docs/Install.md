# Installation Steps

## Install resources in Kubernetes

1. Create Namespace

```sh
sh hack/pre-setup-ktwin.sh
```

2. Deploy Event Store ScyllaDB instance.

```sh
sh hack/setup-scylla-db.sh
```

3. Install Knative and Istio dependencies.

```sh
sh hack/setup-knative-operator.sh
```

> Note: KTWIN uses Knative node selector features to deploy workloads to specific nodes based on node labels. The node selector feature is disabled by default in KTWIN and it can be enabled by [feature flag](https://knative.dev/docs/serving/configuration/feature-flags). You can apply with the following command in `knative-serving` namespace: `kubectl apply -f hack/knative-operator/config-features.yaml`

4. Install Event Brokers.

```sh
sh hack/setup-brokers.sh
```

5. Install CDRs in Cluster and deploy Operator container.

```sh
make install
make deploy IMG=ghcr.io/open-digital-twin/ktwin-operator@sha256:d17285f3e2852023c0dc0d0389615ea96e81ed594d2de8fa480ca178ca2a7b08
```

6. Install Event Store and MQTT Dispatcher resources.

```sh
kubectl apply -Rf hack/ktwin/resources
```

## Local Development

1. Configure your Kubernetes cluster. You can run the platform in [Kind](https://kind.sigs.k8s.io/) in your local computer.

```sh
kind create cluster
```

2. Load Docker image into cluster.

```sh
sh hack/load-local-dependencies.sh
```

3. Create Namespace and Pre-Dependencies

```sh
sh hack/pre-setup-ktwin.sh
```

4. Deploy Event Store ScyllaDB instance.

```sh
sh hack/setup-scylla-db.sh
```

5. Install Knative and Istio dependencies.

```sh
sh hack/setup-knative-operator.sh
```

> Note: KTWIN uses Knative node selector features to deploy workloads to specific nodes based on node labels. The node selector feature is disabled by default in KTWIN and it can be enabled by [feature flag](https://knative.dev/docs/serving/configuration/feature-flags). You can apply with the following command in `knative-serving` namespace: `kubectl apply -f hack/knative-operator/config-features.yaml`

6. Install Event Brokers.

```sh
sh hack/setup-brokers.sh
```

7. Install CDRs in Cluster and Run the Operator locally.

```sh
make install
make run-local
```

8. Install Event Store and MQTT Dispatcher resources.

```sh
kubectl apply -Rf hack/ktwin/resources
```
