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

ROOT=$(realpath $(dirname ${BASH_SOURCE})/..)
cd $ROOT

source hack/latest_tag

DEST=$1
if [[ $DEST == "" ]]; then
  echo "usage: postrelease.sh <community-operators-path>"
  exit 1
fi

cd $DEST
if [[ -n "$(git status --porcelain)" ]]; then
  echo "error: community-operators git working directory not clean"
  git status --porcelain
  exit 1
fi

git checkout master
git fetch --all
git rebase upstream/master

git checkout -b cloud-functions-operator-${TAG}
git rebase master

mkdir -p cloud-functions-operator
cd cloud-functions-operator
cp $ROOT/deploy/olm-catalog/v${TAG}/* .

