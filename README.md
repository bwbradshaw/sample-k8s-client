# Problem statement
Create a program in golang that interacts with a k8s cluster using the client-go library (GitHub - kubernetes/client-go: Go client for Kubernetes. ).  The program should perform the following:

connect to the k8s cluster

print out the namespaces on the cluster

create a new namespace

create a pod in that namespace that runs a simple hello-world container

print out pod names and the namespace they are in for any pods that have a label of ‘k8s-app=kube-dns’ or a similar label is ok as well

delete the hello-world pod created from above

extra credit - show how an client-go informer works

The example should be loaded into a github repo of the candidate’s choice to assist in reviewing of the code.  The candidate should be able to describe the following:

how they set up their k8s dev host

what tools they used to code up the example

how their code is structured and what it does including how they used features of the client-go library

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

# Sample output
```
Connecting to local cluster ...

** Namespaces Before Creation **
	default
	kube-node-lease
	kube-public
	kube-system
	local-path-storage

Creating namespace ...

** Namespaces After Creation **
	default
	kube-node-lease
	kube-public
	kube-system
	local-path-storage
	milky-way

Creating pod ...

 -> pod: kube-scheduler-k8s-kind-control-plane created in namespace: kube-system <-

 -> pod: coredns-64897985d-gzhpt created in namespace: kube-system <-

 -> pod: coredns-64897985d-cthw7 created in namespace: kube-system <-

 -> pod: local-path-provisioner-5ddd94ff66-bvxwb created in namespace: local-path-storage <-

 -> pod: kube-apiserver-k8s-kind-control-plane created in namespace: kube-system <-

 -> pod: kube-controller-manager-k8s-kind-control-plane created in namespace: kube-system <-

 -> pod: etcd-k8s-kind-control-plane created in namespace: kube-system <-

 -> pod: kindnet-9k6fm created in namespace: kube-system <-

 -> pod: kube-proxy-9c8qj created in namespace: kube-system <-

 -> pod: hello-world-pod created in namespace: milky-way <-
INFO[0015] not implemented                              

** List of Pods **
Pod			Namespace	
 ----			----		
hello-world-pod		milky-way
** Deleting pod ... **

** List of Pods after deletion **
Pod	Namespace	
 ----	----
```
