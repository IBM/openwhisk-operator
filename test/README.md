# Running Tests

## Prerequisites

- [kubebuilder](https://book.kubebuilder.io/) testing framework
- [ginkgo](https://onsi.github.io/ginkgo/)

## Unit tests

To run the unit tests, do:

```sh
ginkgo -r --trace -cover -coverprofile cover.out -outputdir=.
```




