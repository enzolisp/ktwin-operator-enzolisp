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

Resources: https://www.rabbitmq.com/kubernetes/operator/operator-monitoring.html

```sh
kubectl apply -n ktwin -f https://raw.githubusercontent.com/rabbitmq/cluster-operator/main/observability/prometheus/monitors/rabbitmq-servicemonitor.yml
kubectl apply -n ktwin -f https://raw.githubusercontent.com/rabbitmq/cluster-operator/main/observability/prometheus/monitors/rabbitmq-cluster-operator-podmonitor.yml
kubectl label ServiceMonitor rabbitmq serviceMonitor=prometheus -n ktwin
kubectl label PodMonitor rabbitmq-cluster-operator podMonitor=prometheus -n ktwin
```

### Creating Prometheus Cluster Roles

Provide Prometheus access to read resources from `ktwin` namespace.

```sh
kubectl apply -n ktwin -f prometheus-roles-ktwin.yaml # Required for RabbitMQ Cluster
```

### Import dashboards

https://github.com/rabbitmq/rabbitmq-server/tree/main/deps/rabbitmq_prometheus/docker/grafana/dashboards

## Knative

https://github.com/knative-extensions/monitoring

```sh
# Install Knative Service Monitor
kubectl delete -f https://raw.githubusercontent.com/knative-sandbox/monitoring/main/servicemonitor.yaml

kubectl apply -f knative-servicemonitor.yaml

kubectl label ServiceMonitor controller serviceMonitor=prometheus -n knative-serving
kubectl label ServiceMonitor autoscaler serviceMonitor=prometheus -n knative-serving
kubectl label ServiceMonitor activator serviceMonitor=prometheus -n knative-serving
kubectl label ServiceMonitor webhook serviceMonitor=prometheus -n knative-serving

# When you use helm to install kube-prometheus-stack, it adds label release: <prometheus-installed-namespace> to Kubernetes resource.
kubectl label ServiceMonitor controller release=prometheus -n knative-serving
kubectl label ServiceMonitor autoscaler release=prometheus -n knative-serving
kubectl label ServiceMonitor activator release=prometheus -n knative-serving
kubectl label ServiceMonitor webhook release=prometheus -n knative-serving

kubectl label ServiceMonitor broker-filter serviceMonitor=prometheus -n knative-eventing
kubectl label ServiceMonitor broker-ingress serviceMonitor=prometheus -n knative-eventing
kubectl label PodMonitor eventing-controller serviceMonitor=prometheus -n knative-eventing
kubectl label PodMonitor imc-controller serviceMonitor=prometheus -n knative-eventing
kubectl label PodMonitor ping-source serviceMonitor=prometheus -n knative-eventing
kubectl label PodMonitor apiserver-source serviceMonitor=prometheus -n knative-eventing

kubectl label ServiceMonitor broker-filter release=prometheus -n knative-eventing
kubectl label ServiceMonitor broker-ingress release=prometheus -n knative-eventing
kubectl label PodMonitor eventing-controller release=prometheus -n knative-eventing
kubectl label PodMonitor imc-controller release=prometheus -n knative-eventing
kubectl label PodMonitor ping-source release=prometheus -n knative-eventing
kubectl label PodMonitor apiserver-source release=prometheus -n knative-eventing

kubectl apply -n knative-eventing -f prometheus-roles-knative.yaml # Required for RabbitMQ Cluster
kubectl apply -n knative-serving -f prometheus-roles-knative.yaml # Required for RabbitMQ Cluster
```

### Import Grafana Dashboards

https://github.com/knative-extensions/monitoring/tree/main/grafana

## Resources

- https://www.rabbitmq.com/kubernetes/operator/operator-monitoring.html
- https://knative.dev/docs/serving/observability/metrics/collecting-metrics/#setting-up-prometheus

```
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install prometheus prometheus-community/kube-prometheus-stack -n monitoring -f values.yaml
```
