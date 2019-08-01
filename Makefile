
# Image URL to use all building/pushing image targets
IMG ?= ibmcom/openwhisk-operator:latest
LOG_LEVEL ?= 0

all: test manager

# Run tests
test: generate fmt vet manifests
	ginkgo -r --trace --compilers=1 -cover -coverprofile cover.out -outputdir=.

# Run e2e test
e2e:
	@test/e2e/test.sh

# Run travis tests
travistest: docker-build e2e
	@tools/travis/docker.sh

# pretty print cover
cover:
	go tool cover -html=cover.out

# Run function tests
testf:
	go test -p 1  github.com/ibm/cloud-functions-operator/pkg/controller/function -v

# Run invocation tests
testi:
	go test -p 1 github.com/ibm/cloud-functions-operator/pkg/controller/invocation -v

# Run auth tests
testa:
	go test -p 1 github.com/ibm/cloud-functions-operator/pkg/controller/auth -v

# Run trigger tests
testt:
	go test -p 1 github.com/ibm/cloud-functions-operator/pkg/controller/trigger -v

# Run rule tests
testr:
	go test -p 1 github.com/ibm/cloud-functions-operator/pkg/controller/rule -v

# Run package tests
testp:
	go test -p 1 github.com/ibm/cloud-functions-operator/pkg/controller/pkg -v -args -logtostderr=true -v=5

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager github.com/ibm/cloud-functions-operator/cmd/manager

# Generate documentation
doc:
	@patch -tub vendor/k8s.io/kube-openapi/pkg/builder/openapi.go hack/kube-openapi.go.patch
	hack/docgen.sh

# Preprocess markdown
syncmd:
	docker run -v $(shell pwd):/doc villardl/markdownx -u doc/README.md

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/manager/main.go -logtostderr=true -v=5

# Install CRDs into a cluster
install: manifests
	kubectl apply -f config/crds

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all
	./hack/crd_fix.sh

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Generate code
generate:
	go generate ./pkg/... ./cmd/...

# Build the docker image
docker-build:
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# Run the operator-sdk scorecard on latest release
scorecard:
	operator-sdk scorecard

# make a release for olm and releases
release: check-tag
	pip install --user PyYAML
	python hack/package.py v${TAG}

.PHONY: lintall
lintall: fmt lint vet

lint:
	golint -set_exit_status=true pkg/

check-tag:
ifndef TAG
	$(error TAG is undefined! Please set TAG to the latest release tag, using the format x.y.z e.g. export TAG=0.1.1 )
endif
