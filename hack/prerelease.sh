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

function cleanup() {
  set +e

  if [[ -n "$TEST_CLUSTER" ]]; then
    kind delete cluster --name ${TEST_CLUSTER}
    unset TEST_CLUSTER
  fi
}

function traperr() {
  echo "ERROR: ${BASH_SOURCE[1]} at about ${BASH_LINENO[0]}"
  cleanup
}

set -o errtrace
trap traperr ERR
trap traperr INT

ROOT=$(realpath $(dirname ${BASH_SOURCE})/..)
cd $ROOT

TAG=$1
if [[ $TAG == "" ]]; then
  echo "usage: prerelease.sh <tag>"
  exit 1
fi

if [[ ${TAG:0:1} != "v" ]]; then
  echo "invalid tag format: must start with v"
  exit 1
fi

echo "Building and pushing latest docker image"
make docker-build docker-push

echo "Creating OLM catalog"
./hack/package.py $TAG

# TODO: needs quay
# echo "Validating OLM catalog"
source hack/latest_tag
# operator-courier verify deploy/olm-catalog/v${TAG}/cloud-functions-operator.v${TAG}.clusterserviceversion.yaml

echo "Running scorecard"
kind create cluster --name olm
export KUBECONFIG="$(kind get kubeconfig-path --name="olm")"
TEST_CLUSTER=olm
operator-sdk scorecard --csv-path "deploy/olm-catalog/v${TAG}/cloud-functions-operator.v${TAG}.clusterserviceversion.yaml"

cleanup