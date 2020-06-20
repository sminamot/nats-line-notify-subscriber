## deploy
```
$ helm install my-nats nats/nats -f helm/values.yaml
$ sops -d k8s/secret/secret-enc.env > k8s/secret/secret.env
$ kustomize build k8s | kubectl apply -f - --prune -l app=nats-line-notify-subscriber
```
