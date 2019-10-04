# Running Tests

## Prerequisites

- [kubebuilder](https://book.kubebuilder.io/) testing framework
- [ginkgo](https://onsi.github.io/ginkgo/)

## Unit tests

To run the unit tests, do:

```sh
ginkgo -r --trace -cover -coverprofile cover.out -outputdir=. ./pkg/...
```

## Integration Tests

<!-- To run the integration tests you need to have `~/.wskprops` property configured.

When using the IBM Cloud Functions service, configure `~/.wskprops` by creating a
namespace:

```sh
ibmcloud fn namespace create testing-operator
``` -->

To run the integration tests:

```sh
ginkgo -r --trace -cover -coverprofile cover.out -outputdir=. ./test/...
```






