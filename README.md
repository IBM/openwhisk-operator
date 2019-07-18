# Apache OpenWhisk Operators

[![Build Status](https://travis-ci.org/IBM/cloud-functions-operator.svg?branch=master)](https://travis-ci.org/IBM/cloud-functions-operator)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)

This project provides a Kubernetes operator for managing [IBM Cloud Functions](https://www.ibm.com/cloud/functions) resources: actions, packages, rules and triggers.


<!-- TOC -->

- [Prerequisites](#prerequisites)
- [Installing the operator](#installing-the-operator)
- [Using the operators](#using-the-operators)
    - [Setting up IBM cloud credentials](#setting-up-ibm-cloud-credentials)
    - [Deploying your first function](#deploying-your-first-function)

<!-- /TOC -->

# Quick start

## Prerequisites

- A cluster running Kubernetes 1.11+
- `kubectl` installed and configured.
- [`kustomize`](https://github.com/kubernetes-sigs/kustomize) installed.

## Installing the operator

1. Clone this repository
2. Install the CRDs using `kubectl`:

```sh
$ kubectl apply -f config/crds
```

3. Then install the operator:

```sh
$ kubectl apply -f config/manager -f config/rbac/
```

By default the operator is installed in the `ibmcloud-operators` namespace and is granted [clustor-wide](./config/rbac/rbac_role_binding.yaml) [permissions](./config/rbac/rbac_role.yaml).

## Using the operators

### Setting up IBM cloud credentials

By default, all operators look for the IBM cloud function credentials in the `seed-defaults-owprops` secret:

[//]: #embed-code(test/e2e/wskprops-secrets.sh)
```sh
# Extract properties from .wskprops
AUTH=$(cat ~/.wskprops | grep 'AUTH' | awk -F= '{print $2}')
APIHOST=$(cat ~/.wskprops | grep 'APIHOST' | awk -F= '{print $2}')

# And create secret
kubectl create secret generic seed-defaults-owprops \
    --from-literal=apihost=$APIHOST \
    --from-literal=auth=$AUTH
```

Alternativalely, you can directly create a k8s secret

[//]: #embed-code(samples/credentials-guest.yaml)
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: seed-default-owprops
stringData:
  apihost: localhost
  auth: "23bc46b1-71f6-4ed5-8c54-816aa4f8c502:123zO3xZCLrMN6v2BKK1dXYFpXlPkccOFqm12CdAsMgRU4VrNZ9lyGVCGuMDGIwP"
  insecure: "true"
```

**NOTE**: be aware that all operators update IBM cloud function entities and can potentially override existing entities.

### Deploying your first function

The `Function` resource kind allows the deployment of actions:

[//]: #embed-code(test/e2e/greetings.yaml)
```yaml
apiVersion: ibmcloud.ibm.com/v1alpha1
kind: Function
metadata:
  name: greetings
spec:
  codeURI: https://raw.githubusercontent.com/apache/incubator-openwhisk-catalog/master/packages/utils/echo.js
  runtime: nodejs:6
  parameters:
  - name: message
    value: Bonjour
```

Deploy it:

```sh
$ kubectl apply -f sample.yaml
```

wait a little bit and run:

```sh
$ ibmcloud wsk action invoke greetings -br
```

# Learn more

- [reference documentation](https://ibm.github.io/cloud-functions-operator/)
- [contributions](./CONTRIBUTING.md)
