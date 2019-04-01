
## Overview

zookeeper-operator implements an operator for Zookeeper 3.4.x based on [operator-sdk](https://github.com/operator-framework/operator-sdk)
and [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime).
It supports creating a zookeeper cluster given a spec containing number of nodes and size of data directory.
Persistence is achieved through 'standard' storage class and PersistentVolumeClaim.

Scaling up and down through rolling update is also supported. Dockerfile to build zookeeper also resides in
same repo.

## Start


```sh

# Make sure [operator-sdk](https://github.com/operator-framework/operator-sdk) is installed
# Refer to operator-sdk docs for more and dependencies of operator-sdk.

$ mkdir -p $GOPATH/src/github.com/operator-framework
$ cd $GOPATH/src/github.com/operator-framework
$ git clone https://github.com/operator-framework/operator-sdk
$ cd operator-sdk
$ git checkout master
$ make dep
$ make install



$ mkdir -p $GOPATH/src/github.com/ronin13/
$ cd $GOPATH/src/github.com/ronin13/
$ git clone https://github.com/ronin13/zookeeper-operator
$ cd zookeeper-operator

# Make sure kubectl points to a running k8s cluster (or minikube for a start).
$ minikube start

# Build docker build
$ eval $(minikube docker-env)
$ make -C docker-zookeeper build

# Create accounts for RBAC
$ kubectl create -f deploy/service_account.yaml
$ kubectl create -f deploy/role.yaml
$ kubectl create -f deploy/role_binding.yaml

# Deploy zookeeper CRD
$ kubectl create -f deploy/crds/wnohang_v1alpha1_zookeeper_crd.yaml

# Deploy zookeeper custom resource  (default name is zoos with 3 nodes).
$ kubectl create -f deploy/crds/wnohang_v1alpha1_zookeeper_cr.yaml

# Finally deploy the operator.
$ kubectl create -f deploy/operator.yaml

# See if the cluster is running.

$ kubectl get pods -l app=zoos
NAME     READY   STATUS    RESTARTS   AGE
zoos-0   0/1     Running   0          6s
zoos-1   0/1     Running   0          6s
zoos-2   0/1     Running   0          6s


$ kubectl describe zookeeper.wnohang.net/zoos
Name:         zoos
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  wnohang.net/v1alpha1
Kind:         Zookeeper
Metadata:
  Creation Timestamp:  2019-04-01T22:16:41Z
  Generation:          1
  Resource Version:    184847
  Self Link:           /apis/wnohang.net/v1alpha1/namespaces/default/zookeepers/zoos
  UID:                 d39f11f4-54cb-11e9-8160-08002759e00a
Spec:
  Nodes:  3
Events:   <none>


# Verify that it is running.

$ for x in 0 1 2;do kubectl exec zoos-$x -- sh -c  "echo -n $x' ';  echo mntr  | nc localhost 2181 | grep zk_server_state"; done
0 zk_server_state       follower
1 zk_server_state       follower
2 zk_server_state       leader

# Cleanup
$ kubectl delete -R -f deploy/


```
