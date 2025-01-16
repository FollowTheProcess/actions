# actions

[![License](https://img.shields.io/github/license/FollowTheProcess/actions)](https://github.com/FollowTheProcess/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/FollowTheProcess/actions.svg)](https://pkg.go.dev/github.com/FollowTheProcess/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/FollowTheProcess/actions)](https://goreportcard.com/report/github.com/FollowTheProcess/actions)
[![GitHub](https://img.shields.io/github/v/release/FollowTheProcess/actions?logo=github&sort=semver)](https://github.com/FollowTheProcess/actions)
[![CI](https://github.com/FollowTheProcess/actions/workflows/CI/badge.svg)](https://github.com/FollowTheProcess/actions/actions?query=workflow%3ACI)
[![codecov](https://codecov.io/gh/FollowTheProcess/actions/branch/main/graph/badge.svg)](https://codecov.io/gh/FollowTheProcess/actions)

A GitHub Actions toolkit for Go!

> [!WARNING]
> **actions is in early development and is not yet ready for use**

![caution](./img/caution.png)

## Project Description

If you want to write non-trivial GitHub Actions you really only have 2 choices:

- Use NodeJS and [actions/toolkit]
- Use Docker and your language of choice

The latter option suffers because you likely have to implement *a lot* of stuff yourself to do things like:

- Retrieve action inputs and handle them
- Write to `$GITHUB_OUTPUT` and/or `$GITHUB_STEP_SUMMARY`
- Writing [Workflow Commands] to provide feedback in the logs
- Platform independent filepath manipulations
- Executing external commands
- Responding to `${{ runner.debug }}` for writing debug logs
- And more...

And the former means you have to write Javascript...

### Go and GitHub Actions

I put it to you that Go could be an excellent language for writing GitHub actions:

- It's fast and memory efficient (saving 💵 on GitHub runner costs)
- Excellent support for concurrency
- A powerful, batteries included standard library with more or less everything an action author needs out of the box
- It plays excellently with Docker, leading to tiny, very efficient images
- It's simple and safe

> [!TIP]
> You don't even need Docker! You could just `GOOS=linux GOARCH=amd64 go build <your action>` and fetch the binary
> (e.g. from a GitHub release) inside a [composite action] and pass the inputs as command line arguments/flags!

The only thing missing is a toolkit providing sensible, common functionality for GitHub Action authors... That's where `actions` comes in!

## Installation

```shell
go get github.com/FollowTheProcess/actions@latest
```

## Quickstart

### Credits

This package was created with [copier] and the [FollowTheProcess/go_copier] project template.

[copier]: https://copier.readthedocs.io/en/stable/
[FollowTheProcess/go_copier]: https://github.com/FollowTheProcess/go_copier
[actions/toolkit]: https://github.com/actions/toolkit
[Workflow Commands]: https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions
[composite action]: https://docs.github.com/en/actions/sharing-automations/creating-actions/creating-a-composite-action
