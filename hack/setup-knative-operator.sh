#!/usr/bin/env bash

# Steps https://knative.dev/docs/install/operator/knative-with-operators

SCRIPT_PATH=$(dirname "$0")

KNATIVE_VERSION=v1.10.0
KNATIVE_OPERATOR_VERSION=v1.11.3
RABBITMQ_CERT_MANAGER_VERSION=v1.11.1

# Install Cert Manager
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/${RABBITMQ_CERT_MANAGER_VERSION}/cert-manager.yaml
kubectl apply -f ${SCRIPT_PATH}/brokers/cluster-operator/2-cert-manager.yaml
kubectl wait --for=condition=available --timeout=200s --all deployments --namespace cert-manager

# Install Knative Operator
kubectl apply -f https://github.com/knative/operator/releases/download/knative-${KNATIVE_OPERATOR_VERSION}/operator.yaml
kubectl wait --for=condition=available --timeout=200s --all deployments

# # Install Istio
istioctl install -y

# Install Knative Serving
kubectl apply -f hack/knative-operator/knative-serving.yaml
kubectl wait --for jsonpath='{.status.phase}=Active' --timeout=5s namespace/knative-serving
kubectl wait --for=condition=available --timeout=200s --all deployments --namespace knative-serving
kubectl wait --for=condition=available --timeout=200s --all deployments --namespace knative-serving

# Install Knative Istio Plugin
kubectl apply -f https://github.com/knative/net-istio/releases/download/knative-${KNATIVE_VERSION}/net-istio.yaml
kubectl wait --for=condition=available --timeout=200s --all deployments --namespace knative-serving

# # Install Knative Eventing
kubectl apply -f hack/knative-operator/knative-eventing.yaml
kubectl wait --for jsonpath='{.status.phase}=Active' --timeout=5s namespace/knative-eventing
kubectl get pods --namespace knative-eventing

kubectl wait --for=condition=available --timeout=200s --all deployments --namespace knative-eventing

echo "Knative setup script has finished"

# Uninstall
# https://knative.dev/docs/install/uninstall/#uninstall-an-operator-based-knative-installation

# kubectl delete KnativeServing knative-serving -n knative-serving
# kubectl delete KnativeEventing knative-eventing -n knative-eventing
# kubectl delete -f https://github.com/knative/operator/releases/download/knative-${KNATIVE_OPERATOR_VERSION}/operator.yaml
