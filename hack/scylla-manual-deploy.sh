#https://github.com/scylladb/scylla-operator/blob/master/deploy/README.md

#cert-manager dependency installation
kubectl apply -f https://raw.githubusercontent.com/scylladb/scylla-operator/refs/heads/master/examples/third-party/cert-manager.yaml
#kubectl apply -f hack/scylla-operator/cert-manager.yaml

kubectl wait --for condition=established crd/certificates.cert-manager.io crd/issuers.cert-manager.io
kubectl -n cert-manager rollout status deployment.apps/cert-manager-webhook

# Production environment
#kubectl apply -f https://github.com/scylladb/scylla-operator/tree/5df529c5f227596723c04016f29ae456836bab67/deploy/operator
kubectl apply -f https://github.com/scylladb/scylla-operator/tree/5df529c5f227596723c04016f29ae456836bab67/deploy/operator
kubectl wait --for condition=established crd/scyllaclusters.scylla.scylladb.com
kubectl wait --for condition=established crd/nodeconfigs.scylla.scylladb.com
kubectl wait --for condition=established crd/scyllaoperatorconfigs.scylla.scylladb.com
kubectl -n scylla-operator rollout status deployment.apps/scylla-operator
kubectl -n scylla-operator rollout status deployment.apps/webhook-server