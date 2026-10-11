[![Go Report Card](https://goreportcard.com/badge/github.com/raqqun/kcvm)](https://goreportcard.com/report/github.com/raqqun/kcvm)
![Build](https://github.com/raqqun/kcvm/workflows/Build/badge.svg?branch=master)

# kcvm
kcvm is a kubectl version manager.

It installs several kubectl versions side by side and lets you switch between them,
so you can always use a kubectl that matches the cluster you are talking to.

## Installation

- Download the latest linux-amd64 binary from the [releases page](https://github.com/raqqun/kcvm/releases), e.g.

`curl -L -o /tmp/kcvm https://github.com/raqqun/kcvm/releases/download/v0.2.2-beta1/kcvm-v0.2.2-beta1-linux-amd64`

- Make it executable and move to $PATH

`chmod +x /tmp/kcvm && sudo mv /tmp/kcvm /usr/local/bin`

On other platforms, build it with Go 1.16 or later:

`go install github.com/raqqun/kcvm/cmd/kcvm@latest`

## Commands

#### Initialize kcvm at ~/.kcvm

`kcvm init`

Add kcvm folder to your $PATH : `export PATH=$HOME/.kcvm:$PATH`

#### List available kubectl versions

`kcvm list remote`

Stable versions are listed from newest to oldest.

#### List installed kubectl versions

`kcvm list local`

The version in use is marked with `*`.

#### Install an available kubectl version

`kcvm install [VERSION]`

The installed version becomes the one in use. Use `--force` to download a version again.

#### Use a specific installed kubectl version

`kcvm use [VERSION]`

Versions can be written with or without the leading `v`, e.g. `v1.20.0` or `1.20.0`.

## How it works

kcvm keeps every kubectl binary in `~/.kcvm/bin` and points the `~/.kcvm/kubectl`
symlink to the version in use:

```
~/.kcvm/
├── kubectl -> ~/.kcvm/bin/kubectl-v1.20.0
└── bin/
    ├── kubectl-v1.19.4
    └── kubectl-v1.20.0
```

- Binaries are downloaded from `dl.k8s.io` for your OS and architecture and verified
  against their published sha256 checksum (releases older than v1.16 publish none).
- Downloads are written to a temporary file and moved into place only once complete
  and verified, and switching versions replaces the symlink atomically, so an
  interrupted command never leaves you without a working kubectl.
- Available versions are read from the tags of the
  [kubernetes repository](https://github.com/kubernetes/kubernetes) with a single git
  request, so listing them is not subject to GitHub API rate limits.

Use the global `--dir` flag to keep kcvm somewhere else than `~/.kcvm`, e.g.
`kcvm --dir /opt/kcvm install v1.20.0`.

## Development

`go test ./test/kcvm/` runs the tests against local fake servers, no network access is needed.
