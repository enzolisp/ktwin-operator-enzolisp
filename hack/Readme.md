# Scripts for development

This directory contains scripts useful in local development.

## How to run

```sh
sh hack/create-kind-cluster.sh && \
    sh hack/pre-setup-ktwin.sh && \
    sh hack/setup-knative.sh && \
    sh hack/setup-brokers.sh && \
    sh hack/setup-scylla-db.sh
```
