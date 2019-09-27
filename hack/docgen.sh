#!/bin/bash
#
# Copyright 2017-2018 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
set -e

ROOT=$(dirname ${BASH_SOURCE})/..
source $ROOT/hack/lib/library.sh
ROOT=$(bash::realpath $ROOT)

CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ${GOPATH}/src/k8s.io/code-generator)}

go run ${CODEGEN_PKG}/cmd/openapi-gen/main.go \
  --input-dirs "github.com/ibm/cloud-functions-operator/pkg/apis/ibmcloud/v1alpha1" \
  --input-dirs "github.com/ibm/cloud-functions-operator/vendor/github.com/ibm/cloud-operators/pkg/lib/keyvalue/v1" \
  --output-package "github.com/ibm/cloud-functions-operator/pkg/openapi" \
  --output-base "${ROOT}/../../.."

mkdir -p api/swagger
go run $ROOT/hack/openapi/builder.go api/swagger/swagger.json

docker run -v ${ROOT}/docs:/root/docs -v ${ROOT}/api:/root/api villardl/spectacle -t /root/docs /root/api/swagger/swagger-doc.json

# Override stylesheet until we figure out how to change layout in spectacle.
cp -f $ROOT/docs/stylesheets/spectacle.min.css.patched $ROOT/docs/stylesheets/spectacle.min.css