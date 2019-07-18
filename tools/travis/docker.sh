#!/usr/bin/env bash
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

ROOT=$(dirname ${BASH_SOURCE})/../..

TAG=${TRAVIS_TAG:-vlatest}
TAG=${TAG:1}
NAME=ibmcom/cloud-functions-operator

cd $ROOT

echo "make docker image"
docker build -t $NAME:$TAG -f Dockerfile .

echo "push docker image"
docker login -u $DOCKER_USERNAME -p $DOCKER_PASSWORD
docker push $NAME:$TAG
