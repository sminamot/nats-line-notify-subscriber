## deploy
```
$ sops -d k8s/secret/secret-enc.env > k8s/secret/secret.env
$ kustomize build k8s | kubectl apply -f -
```
