# Apache OpenWhisk Operators

[![Build Status](https://travis-ci.com/IBM/openwhisk-operator.svg?branch=master)](https://travis-ci.com/IBM/openwhisk-operator)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)

This project provides a collection of Kubernetes operators for managing [Apache OpenWhisk](https://openwhisk.apache.org/) resources namely actions, packages, rules and triggers.


<!-- TOC -->

- [Apache OpenWhisk Operators](#apache-openwhisk-operators)
- [Quick start](#quick-start)
    - [Prerequisites](#prerequisites)
    - [Installing the operators](#installing-the-operators)
    - [Using the operators](#using-the-operators)
        - [Setting up OpenWhisk credentials](#setting-up-openwhisk-credentials)
        - [Deploying your first function](#deploying-your-first-function)
- [Learn more](#learn-more)

<!-- /TOC -->

# Quick start

## Prerequisites

- A cluster running Kubernetes 1.11+ 
- `kubectl` installed and configured.
- [`kustomize`](https://github.com/kubernetes-sigs/kustomize) installed.

## Installing the operators

1. Install the CRDs using `kubectl`:

```sh
$ kustomize build github.com/IBM/openwhisk-operator//config/crds | kubectl apply -f -
```

2. Then install the operators:

```sh
$ kustomize build github.com/IBM/openwhisk-operator//config/default | kubectl apply -f -
```

By default the operators are installed in the `openwhisk-system` namespace and are granted [clustor-wide](./config/rbac/rbac_role_binding.yaml) [permissions](./config/rbac/rbac_role.yaml).

## Using the operators

### Setting up OpenWhisk credentials

By default, all operators look for OpenWhisk credentials in the `seed-default-owprops` secret:

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

**NOTE**: be aware that all operators update OpenWhisk entities and can potentially override existing entitites.

### Deploying your first function 

The `Function` resource kind allows the deployement of actions:

[//]: #embed-code(test/e2e/greetings.yaml)
```yaml
apiVersion: openwhisk.seed.ibm.com/v1beta1
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
$ wsk action invoke greetings -br
```

# Learn more

- [reference documentation](https://ibm.github.io/openwhisk-operator/)
- [contributions](./CONTRIBUTING.md)
