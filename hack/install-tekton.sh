TEKTON_RELEASE="previous/v0.49.0"
NAMESPACE="default"

install_tekton() {
  kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/$TEKTON_RELEASE/release.yaml
  kubectl create clusterrolebinding $NAMESPACE:knative-serving-namespaced-admin --clusterrole=knative-serving-namespaced-admin  --serviceaccount=$NAMESPACE:default
}

install_tekton

echo "Installed successfully"


func deploy --remote --registry docker.io/agwermann --git-url=https://github.com/agwermann/hello-python-knative-func --image hello-python
