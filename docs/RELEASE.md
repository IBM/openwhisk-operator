This document describes the release process.


## Prerequisites

- [operator-sdk](https://github.com/operator-framework/operator-sdk)
- [operator-courier](https://github.com/operator-framework/operator-courier)
- [kind](https://github.com/kubernetes-sigs/kind)

## Pre-release

The `hack/prerelease.sh` script generates the OLM catalog and run the scorecard utility.

```sh
hack/prerelease.sh <version>
```

`version` must start with a `v` and contains 3 digits, eg. `v0.8.0`

After successful completion, commit and push the generated files.

## Release

The `hack/release.sh` script tags the repository, generates release artifacts and push them on github.

```sh
hack/prerelease.sh <version> <github_token>
```

## Post-release

Finally, `hack/postrelease.sh` updates your local `community-operators` git repository with the latest
OLM catalog.

```sh
hack/postrelease.sh <community-operators-local-dir>
```

After completion, you can submit a PR against the community-operators git repository
