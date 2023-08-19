#!/usr/bin/env bash

# Steps https://knative.dev/docs/install/operator/knative-with-operators

SCRIPT_PATH=$(dirname "$0")

KNATIVE_VERSION=v1.10.0
KNATIVE_OPERATOR_VERSION=v1.11.3

# Install Knative Operator
kubectl create namespace knative
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