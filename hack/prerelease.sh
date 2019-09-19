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
  local errorcode="$1"
  set +e

  if [[ -n "$TEST_CLUSTER" ]]; then
    kind delete cluster --name ${TEST_CLUSTER}
    unset TEST_CLUSTER
  fi

  exit $errorcode
}

function traperr() {
  echo "ERROR: ${BASH_SOURCE[1]} at about ${BASH_LINENO[0]}"
  cleanup 1
}

set -o errtrace
trap traperr ERR
trap traperr INT

ROOT=$(dirname ${BASH_SOURCE})/..
cd $ROOT

if [[ $QUAY_TOKEN == "" ]]; then
  echo "missing QUAY_TOKEN. More information: https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#quay-login"
  exit 1
fi

TAG=$1
if [[ $TAG == "" ]]; then
  echo "usage: prerelease.sh <tag> <quay_namespace>"
  exit 1
fi

if [[ ${TAG:0:1} != "v" ]]; then
  echo "invalid tag format: must start with v"
  exit 1
fi

QUAY_NAMESPACE=$2
if [[ $QUAY_NAMESPACE == "" ]]; then
  echo "missing quay_namespace. usage: prerelease.sh <tag> <quay_namespace>. More information: https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#push-to-quayio"
  exit 1
fi

echo "Creating OLM catalog"
./hack/package.py $TAG

source hack/latest_tag # trim 'v'

IMG=cloudoperators/cloud-functions-operator:${TAG}

echo "Building and pushing candidate docker image"
docker build . -t ${IMG}
docker push ${IMG}

OPERATOR_DIR=deploy/olm-catalog/v${TAG}/
PACKAGE_NAME=cloud-functions-operator
PACKAGE_VERSION=$TAG

echo "Linting OLM catalog"
operator-courier verify --ui_validate_io $OPERATOR_DIR

echo "Pushing operator to quay.io"
operator-courier push "$OPERATOR_DIR" "$QUAY_NAMESPACE" "$PACKAGE_NAME" "$PACKAGE_VERSION" "$QUAY_TOKEN"

echo "Starting k8s cluster"
kind create cluster --name olm
export KUBECONFIG="$(kind get kubeconfig-path --name="olm")"
TEST_CLUSTER=olm

echo "Installing OLM"
kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.10.0/crds.yaml
kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.10.0/olm.yaml

echo "Installing the Operator Marketplace"
tmp_dir=$(mktemp -d)
pushd $tmp_dir
git clone --depth 1 https://github.com/operator-framework/operator-marketplace.git
kubectl apply -f operator-marketplace/deploy/upstream/
popd

echo "Creating the OperatorSource"

cat <<EOF | kubectl create -f -
apiVersion: operators.coreos.com/v1
kind: OperatorSource
metadata:
  name: ${QUAY_NAMESPACE}-operators
  namespace: marketplace
spec:
  type: appregistry
  endpoint: https://quay.io/cnr
  registryNamespace: ${QUAY_NAMESPACE}
EOF

kubectl get operatorsource ${QUAY_NAMESPACE}-operators -n marketplace

echo "Creating an OperatorGroup"
cat <<EOF | kubectl create -f -
apiVersion: operators.coreos.com/v1alpha2
kind: OperatorGroup
metadata:
  name: ${PACKAGE_NAME}group
  namespace: marketplace
spec:
  targetNamespaces: []
EOF

echo "Creating a Subscription"

cat <<EOF | kubectl create -f -
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: ${PACKAGE_NAME}-subscription
  namespace: marketplace
spec:
  channel: alpha
  name: ${PACKAGE_NAME}
  source: ${QUAY_NAMESPACE}-operators
  sourceNamespace: marketplace
EOF

echo "Waiting for CSV to be healthy"

for i in {0..20}; do
  if [[ $(kubectl get clusterserviceversion -n marketplace cloud-functions-operator.v${TAG} -o=jsonpath='{.status.phase}') == "Succeeded" ]]; then
    break
  fi
  sleep 3
done

echo "Inserting the proxy scorecard into the deployment"
kubectl patch -n marketplace deployments.app cloud-functions-operator -p '
{
    "spec": {
        "template": {
            "spec": {
                "containers": [
                    {
                        "name": "scorecard-proxy",
                        "image": "quay.io/operator-framework/scorecard-proxy",
                        "command": [
                            "scorecard-proxy"
                        ],
                        "env": [
                            {
                                "name": "WATCH_NAMESPACE",
                                "valueFrom": {
                                    "fieldRef": {
                                        "apiVersion": "v1",
                                        "fieldPath": "metadata.namespace"
                                    }
                                }
                            }
                        ],
                        "imagePullPolicy": "Always",
                        "ports": [
                            {
                                "name": "proxy",
                                "containerPort": 8889
                            }
                        ]
                    }
                ]
            }
        }
    }
}'

sleep 5 # wait a bit

echo "Running scorecard"
for cr_file in $(find "deploy/crds" -name "*_cr.yaml" -print); do
  operator-sdk scorecard --namespace marketplace --olm-deployed --cr-manifest $cr_file  --csv-path "${OPERATOR_DIR}cloud-functions-operator.v${TAG}.clusterserviceversion.yaml"
done

# cleanup 0