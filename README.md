# Tools used

* kind
* docker
* kubectl
* vi
* golang

# Kind setup

```
brew install kind
kind create cluster --name k8s-kind
kind get clusters
kubectl cluster-info --context kind-k8s-kind
```
# Sample Client setup
```
go get k8s.io/client-go@latest
```
