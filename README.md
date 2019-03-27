
zookeeper-operator implements an operator for Zookeeper 3.4.x based on operator-sdk and container-runtime.
It supports creating a zookeeper cluster given a spec containing number of nodes. Scaling up through rolling update
is also supported. It uses the zookeeper image from 'ronin/zookeeper-k8s' on docker hub.
