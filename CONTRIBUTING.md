# Sign the CLA

If you have not previously done so, please fill out and
submit the [Contributor License Agreement](https://cla.pivotal.io/sign/pivotal).

# Contributing to Indicator Protocol

All kinds of contributions to `monitoring-indicator-protocol` are welcome, whether they be improvements to
documentation, feedback or new ideas, or improving the application directly with
bug fixes, improvements to existing features or adding new features.

## Start with a github issue

In all cases, following this workflow will help all contributors to `monitoring-indicator-protocol` to
participate more equitably:

1. Search existing github issues that may already describe the idea you have.
   If you find one, consider adding a comment that adds additional context about
   your use case, the exact problem you need solved and why, and/or your interest
   in helping to contribute to that effort.
2. If there is no existing issue that covers your idea, open a new issue to
   describe the change you would like to see in `monitoring-indicator-protocol`. Please provide as much
   context as you can about your use case, the exact problem you need solved and why,
   and the reason why you would like to see this change. If you are reporting a bug,
   please include steps to reproduce the issue if possible.
3. Any number of folks from the community may comment on your issue and ask
   additional questions. A maintainer will add the `pr welcome` label to the
   issue when it has been determined that the change will be welcome. Anyone
   from the community may step in to make that change.
4. If you intend to make the changes, comment on the issue to indicate your
   interest in working on it to reduce the likelihood that more than one person
   starts to work on it independently.

# Development Guide

## Getting Started

First things first. Just clone the repo and run the tests to make sure you're
ready to safely start exploring or adding new features.

### Clone the repo

```bash
git clone https://github.com/pivotal/monitoring-indicator-protocol
```

`monitoring-indicator-protocol` should _**NOT**_ be cloned to your `GOPATH`.

### Run the tests

We recommend running using a minimum of [Go 1.13](https://golang.org/dl).
The tests may run fine, but you may experience issues when using `go run` or `go build`

We offer several scripts to make local testing convenient.

To run local unit tests:
```bash
./scripts/test.sh
```

To run BOSH end to end tests:
```bash
./scripts/test.sh bosh_e2e
```

To run k8s end to end tests, some additional setup & dependencies are required.
Visit [the k8s readme](k8s/README.md).

## Vendoring dependencies

Our vendoring tool of choice at present is [go modules](https://github.com/golang/go/wiki/Modules)
which is rapidly becoming the standard.

Adding a dependency is relatively straightforward (first make sure you have the dep binary):

```bash
  go get github.com/some-user/some-dep
```

Check in both the manifest changes and the file additions in the vendor directory.

## Code Generation
Before submitting your PR, ensure that you update any generated code
(which includes schema asset generation and k8s generated code).

This is done with:
```bash
hack/update-codegen.sh
go-bindata -o pkg/asset/schema.go -pkg asset schemas.yml
```

## Contibuting your changes

1. When you have a set of changes to contribute back to `monitoring-indicator-protocol`, create a pull
   request (PR) and reference the issue that the changes in the PR are
   addressing.
   **NOTE:** maintainers of `monitoring-indicator-protocol` with commit access _may_ commit
   directly to `monitoring-indicator-protocol` instead of creating a pull request. Alternatively, they may choose
   to create a pull request for greater visibility around a set of changes.
   There is no black and white rule here for maintainer. Use your judgement.
2. The code in your pull request will be automatically tested in our continuous
   integration pipeline. At this time, we cannot expose all the logs for this
   pipeline, but we may do so in the future if we can determine it is safe and
   unlikely to lead to any exposure of sensitive information.
3. Your pull request will be reviewed by one or more maintainers. You may also
   receive feedback from others in the community. The feedback may come in the
   form of requests for additional changes to meet expectations for code
   quality, consistency or test coverage. Or it could be clarifying questions to
   better understand the decisions you made in your implementation.
4. When a maintainer accepts your changes, they will merge your pull request.
   If there are outstanding requests for changes or other small changes they
   feel can be made to improve the changed code, they may make additional
   changes or merge the changes manually. It's always nice to have changes come
   in just as the team would like to see them, but we'll try not to hold up a pull
   request for a long period of time due to minor changes.

NOTE: With any significant change in behavior to `monitoring-indicator-protocol` that should be noted in
the next release's release notes, you should also add a note to [CHANGELOG.md](./CHANGELOG.md).

## Design Goals

- Define an open source standard for Observability as Code
- Enable platform operators and developers to follow Site Reliability Engineering practices.
- Make default observability easy by include 1 yaml file within the developers code repository so that maintenance
is easy

## Anti-goals

- Indicator Protocol will not manage state or statuses of instances the registry knows about
- The registry will not be a database, but rather a leaky bucket that periodically retrieves indicators

## Technical Design Guidelines

- Drive features through Test Driven Development with acceptance tests and unit tests
    - If no tests are included, the PR will be rejected
- Use the built in Golang testing library with the Gomega matcher
- Keep to the [Kubernetes YAML standard](https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/)

# Becoming a maintainer

At this time, there is no official process for becoming a maintainer to `monitoring-indicator-protocol`.
The project is currently in maintenance mode by the Pivotal CF Observability program
but accepting PRs.

# Prior to a PR

Please summarize your changes in CHANGELOG.yml.
Explain the problem you solved,
what command(s) were affected, and
what issue the PR is addressing.

**Thank you for being an `monitoring-indicator-protocol` contributor!**
