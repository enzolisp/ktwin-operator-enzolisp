#!/usr/bin/env bash

SCRIPT_PATH=$(dirname "$0")

KNATIVE_VERSION=v1.10.0

# RabbitMQ
RABBITMQ_VERSION=v2.2.0
RABBITMQ_CERT_MANAGER_VERSION=v1.11.1
RABBITMQ_MESSAGING_TOPOLOGY_OPERATOR_VERSION=v1.10.3
KNATIVE_RABBITMQ_BROKER_VERSION=v1.9.1

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

# MQTT Deployment
# kubectl create namespace default
# kubectl apply -f ${SCRIPT_PATH}/mosquitto --namespace default
# kubectl wait --for=condition=available --timeout=200s --all deployments --namespace default

# RabbitMQ Cluster
# kubectl apply -f https://github.com/rabbitmq/cluster-operator/releases/download/${RABBITMQ_VERSION}/cluster-operator.yml
# kubectl wait --for=condition=available --timeout=200s --all deployments --namespace rabbitmq-system

# kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/${RABBITMQ_CERT_MANAGER_VERSION}/cert-manager.yaml
# kubectl wait --for=condition=available --timeout=200s --all deployments --namespace cert-manager

# kubectl apply -f https://github.com/rabbitmq/messaging-topology-operator/releases/download/${RABBITMQ_MESSAGING_TOPOLOGY_OPERATOR_VERSION}/messaging-topology-operator-with-certmanager.yaml
# kubectl wait --for=condition=available --timeout=200s --all deployments --namespace rabbitmq-system

# # RabbitMQ Eventing
# kubectl apply -f https://github.com/knative-sandbox/eventing-rabbitmq/releases/download/knative-${KNATIVE_RABBITMQ_BROKER_VERSION}/rabbitmq-broker.yaml
# kubectl wait --for=condition=available --timeout=200s --all deployments --namespace knative-eventing

# # RabbitMQ Cluster
# kubectl apply -f ${SCRIPT_PATH}/rabbitmq-cluster -n default
# kubectl wait --for=condition=available --timeout=200s --all deployments --namespace knative-eventing
# kubectl wait --for=condition=available --timeout=200s --all deployments --namespace default
# kubectl wait --for=condition=Ready --timeout=200s --all pods --namespace default

# # RabbitMQ Broker
# kubectl apply -f ${SCRIPT_PATH}/rabbitmq-broker -n default
# kubectl wait --for=condition=available --timeout=200s --all deployments --namespace knative-eventing
# kubectl wait --for=condition=available --timeout=200s --all deployments --namespace default
# kubectl wait --for=condition=Ready --timeout=200s --all pods --namespace default

echo "Local setup script has finished"