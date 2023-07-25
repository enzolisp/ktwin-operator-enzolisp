#!/usr/bin/env bash

SCRIPT_PATH=$(dirname "$0")

KNATIVE_VERSION=v1.10.0

### Execute Installation scripts

# Install Knative Serving
kubectl apply -f https://github.com/knative/serving/releases/download/knative-${KNATIVE_VERSION}/serving-crds.yaml
kubectl apply -f https://github.com/knative/serving/releases/download/knative-${KNATIVE_VERSION}/serving-core.yaml
kubectl get pods --namespace knative-serving

kubectl wait --for=condition=available --timeout=200s --all deployments --namespace knative-serving

# Instal Istio
kubectl apply -l knative.dev/crd-install=true -f https://github.com/knative/net-istio/releases/download/knative-${KNATIVE_VERSION}/istio.yaml

kubectl wait --for=condition=available --timeout=200s --all deployments --namespace istio-system

kubectl apply -f https://github.com/knative/net-istio/releases/download/knative-${KNATIVE_VERSION}/istio.yaml
kubectl apply -f https://github.com/knative/net-istio/releases/download/knative-${KNATIVE_VERSION}/net-istio.yaml
kubectl --namespace istio-system get service istio-ingressgateway
kubectl get pods --namespace knative-serving
kubectl get pods --namespace istio-system

kubectl wait --for=condition=available --timeout=200s --all deployments --namespace knative-serving
kubectl wait --for=condition=available --timeout=200s --all deployments --namespace istio-system

# Install Eventing Components
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-${KNATIVE_VERSION}/eventing-crds.yaml
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-${KNATIVE_VERSION}/eventing-core.yaml
kubectl get pods --namespace knative-eventing

kubectl wait --for=condition=available --timeout=200s --all deployments --namespace knative-eventing

echo "Knative setup script has finished"