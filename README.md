# Apache OpenWhisk Operators

[![Build Status](https://travis-ci.org/IBM/openwhisk-operator.svg?branch=master)](https://travis-ci.org/IBM/openwhisk-operator)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)

This project provides a collection of Kubernetes operators for managing [Apache OpenWhisk](https://openwhisk.apache.org/) resources namely actions, packages, rules and triggers.


<!-- TOC -->

- [Apache OpenWhisk Operators](#apache-openwhisk-operators)
- [Quick start](#quick-start)
    - [Prerequisites](#prerequisites)
    - [Installing the operators](#installing-the-operators)
- [Using the operators](#using-the-operators)
    - [Setting up OpenWhisk credentials](#setting-up-openwhisk-credentials)
    - [Deploying function](#deploying-function)
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

2. Then install the operator:

```sh
$ kustomize build github.com/IBM/openwhisk-operator//config/default | kubectl apply -f -
```

# Using the operators

## Setting up OpenWhisk credentials

By default, all operators look for OpenWhisk credentials in the `seed-default-owprops` secret:

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

## Deploying function 

You can use the `Function` operator to deploy actions:

[//]: #embed-code(samples/function.yaml)
```yaml
apiVersion: openwhisk.seed.ibm.com/v1beta1
kind: Function
metadata:
  name: myfunction
  namespace: default
spec:
  codeURI: https://raw.githubusercontent.com/apache/incubator-openwhisk-catalog/master/packages/utils/echo.js
  runtime: nodejs:6
  parameters:
  - name: message
    value: "Hello"
```

# Learn more

- [reference documentation](https://pages.github.com/ibm/openwhisk-operator)
- [contributions](./CONTRIBUTING.md)
