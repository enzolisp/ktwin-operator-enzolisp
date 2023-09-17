# Install operator helm dependencies
helm repo add scylla https://scylla-operator-charts.storage.googleapis.com/stable
helm repo update

# Configure cert-manager
kubectl apply -f https://raw.githubusercontent.com/scylladb/scylla-operator/v1.9/examples/common/cert-manager.yaml
kubectl wait -n cert-manager --for=condition=ready pod -l app=cert-manager --timeout=200s

# Configure operator with helm values
helm install scylla-operator scylla/scylla-operator --values hack/scylla-operator/helm/values.operator.yaml --create-namespace --namespace scylla-operator
kubectl wait -n scylla-operator --for=condition=ready pod -l app.kubernetes.io/name=scylla-operator --timeout=200s

# Configure manager with helm values
helm install scylla-manager scylla/scylla-manager --values hack/scylla-operator/helm/values.manager.yaml --create-namespace --namespace scylla-manager
kubectl wait -n scylla-manager --for=condition=ready pod -l app.kubernetes.io/name=scylla-manager --timeout=200s
kubectl wait -n scylla-manager --for=condition=ready pod -l app.kubernetes.io/name=scylla-manager-controller --timeout=200s

# Configure scylla with helm values
helm install scylla scylla/scylla --values hack/scylla-operator/helm/values.cluster.yaml --create-namespace --namespace ktwin
kubectl wait -n ktwin --for=condition=ready pod -l app.kubernetes.io/name=scylla --timeout=200s

# Configure scylla cluster monitoring
kubectl apply -f hack/scylla-operator/monitoring.yaml

# Import Grafana Dashboards
# https://github.com/scylladb/scylla-monitoring/tree/master/grafana

# Uninstall
# helm uninstall scylla -n ktwin
# helm uninstall scylla-manager -n scylla-manager
# helm uninstall scylla-operator -n scylla-operator

# Expose scylla
# kubectl port-forward --address 0.0.0.0 svc/scylla-client 9042:9042 -n ktwin