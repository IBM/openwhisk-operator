
# Image URL to use all building/pushing image targets
IMG ?= controller:latest
LOG_LEVEL ?= 0

all: test manager

# Run tests
test: generate fmt vet manifests
	ginkgo -r --trace --compilers=1 -cover -coverprofile cover.out -outputdir=. -- -v=${LOG_LEVEL} -logtostderr=true 
	
# pretty print cover
cover: 
	go tool cover -html=cover.out

# Run function tests
testf:
	go test  -p 1  github.com/ibm/openwhisk-operator/pkg/controller/function -v -args -logtostderr=true -v=5

# Run invocation tests
testi:
	go test -p 1 github.com/ibm/openwhisk-operator/pkg/controller/invocation -v -args -logtostderr=true -v=5

# Run auth tests
testa:
	go test -p 1 github.com/ibm/openwhisk-operator/pkg/controller/auth -v -args -logtostderr=true -v=5

# Run trigger tests
testt:
	go test -p 1 github.com/ibm/openwhisk-operator/pkg/controller/trigger -v -args -logtostderr=true -v=5

# Run rule tests
testr:
	go test -p 1 github.com/ibm/openwhisk-operator/pkg/controller/rule -v -args -logtostderr=true -v=5

# Run composition tests
testc:
	go test -p 1 github.com/ibm/openwhisk-operator/pkg/controller/composition -v -args -logtostderr=true -v=5

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager github.com/ibm/openwhisk-operator/cmd/manager

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

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f config/crds
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all
	patch -p0 -i config/patches/crd.patch

# Generate patched
patches:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go crd
	diff -u config/crds/ config/expected/ > config/patches/crd.patch

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
docker-build: test
	docker build . -t ${IMG}
	@echo "updating kustomize image patch file for manager resource"
	sed -i'' -e 's@image: .*@image: '"${IMG}"'@' ./config/default/manager_image_patch.yaml

# Push the docker image
docker-push:
	docker push ${IMG}
