## Install

```sh
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install ktwin-graph-redis bitnami/redis -f values.yaml -n ktwin
```

### Upgrade

```sh
helm upgrade ktwin-graph-redis bitnami/redis -f values.yaml -n ktwin
```

### Uninstall

```sh
helm uninstall ktwin-graph-redis
```
