# Installation Steps

## Install in a Kubernetes Cluster


## Local Development

1. Configure your Kubernetes cluster. You can run the platform in [Kind](https://kind.sigs.k8s.io/) in your local computer.

```sh
kind create cluster
```

2. Load Docker image into cluster.

```sh
sh hack/load-local-dependencies.sh
```

3. Deploy ScillaDB for the Event Store.

```sh
sh hack/setup-scylla-db.sh
```

4. Create Namespace and Pre-Dependencies

```sh
sh hack/pre-setup.ktwin.sh
```

5. Install Knative and Istio dependencies.

```sh
sh hack/setup-knative.sh
```

> Note: KTWIN uses Knative node selector features to deploy workloads to specific nodes based on node labels. The node selector feature is disabled by default in KTWIN and it can be enabled by [feature flag](https://knative.dev/docs/serving/configuration/feature-flags). You can apply with the following command in `knative-serving` namespace: `kubectl apply -f hack/knative-operator/config-features.yaml`

5. Install Message Brokers.

```sh
sh hack/setup-brokers.sh
```

5. Install CDR in Cluster and Run the Operator locally.

```
make install
make run-local
```
