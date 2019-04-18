build:
	operator-sdk build zookeeper-operator


dep:
	dep ensure -v

unit: lint
	go test -v $(shell go list ./... | grep 'pkg/')


# Make sure minikube is running and is in PATH.
end-to-end: lint
	minikube status || exit 1
	eval $(minikube docker-env)
	operator-sdk test local ./test/e2e

test: unit end-to-end

lint: golangci
	golangci-lint run

golangci:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.16.0

minikube:
	wget -O $(shell go env GOPATH)/bin/minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 && chmod +x $(shell go env GOPATH)/bin/minikube
	# GOPATH/bin has to be on your PATH.
	minikube start
