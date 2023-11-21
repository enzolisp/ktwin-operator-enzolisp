# Pre-requisites

- Install Prometheus and Grafana in the cluster.

## Labeling namespaces

```sh
kubectl label namespaces ktwin serviceMonitor=prometheus
kubectl label namespaces ktwin podMonitor=prometheus
kubectl label namespaces knative-serving serviceMonitor=prometheus
kubectl label namespaces knative-serving podMonitor=prometheus
kubectl label namespaces knative-eventing serviceMonitor=prometheus
kubectl label namespaces knative-eventing podMonitor=prometheus
kubectl label namespaces rabbitmq-system serviceMonitor=prometheus
kubectl label namespaces rabbitmq-system podMonitor=prometheus
```

## RabbitMQ Monitoring

```sh
kubectl apply -n ktwin -f rabbitmq-monitoring.yaml
```

### Creating Prometheus Cluster Roles

Provide Prometheus access to read resources from `ktwin` namespace.

```sh
kubectl apply -n ktwin -f prometheus-roles-ktwin.yaml # Required for RabbitMQ Cluster
```

### Dashboards

Import the dashboards in Grafana: https://github.com/rabbitmq/cluster-operator/tree/main/observability/grafana/dashboards

### Import dashboards

https://github.com/rabbitmq/rabbitmq-server/tree/main/deps/rabbitmq_prometheus/docker/grafana/dashboards

## Knative

https://github.com/knative-extensions/monitoring

```sh
# Install Knative Service Monitor - Source: https://raw.githubusercontent.com/knative-sandbox/monitoring/main/servicemonitor.yaml
kubectl apply -f knative-monitoring.yaml
kubectl apply -f prometheus-roles-knative.yaml -n knative-serving
kubectl apply -f prometheus-roles-knative.yaml -n knative-eventing
kubectl apply -f prometheus-roles-knative.yaml -n ktwin
```

kubectl delete -f prometheus-roles-knative.yaml -n knative-serving
kubectl delete -f prometheus-roles-knative.yaml -n knative-eventing
kubectl delete -f prometheus-roles-knative.yaml -n ktwin
kubectl delete -f knative-monitoring.yaml

### Import Grafana Dashboards

https://github.com/knative-extensions/monitoring/tree/main/grafana

## Expose Prometheus Port

```sh
kubectl port-forward -n default svc/prometheus-operated 9090 -n monitoring
```

## Resources

- https://www.rabbitmq.com/kubernetes/operator/operator-monitoring.html
- https://knative.dev/docs/serving/observability/metrics/collecting-metrics/#setting-up-prometheus

```
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install prometheus prometheus-community/kube-prometheus-stack -n monitoring -f values.yaml
```
