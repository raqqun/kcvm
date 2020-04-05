[![Go Report Card](https://goreportcard.com/badge/github.com/raqqun/kcvm)](https://goreportcard.com/report/github.com/raqqun/kcvm)


# kcvm
kcvm is a kubectl version manager

## Installation

- Download latest release.

`curl -L -o /tmp/kcvm https://github.com/raqqun/kcvm/releases/download/v1.0.0-alpha/kcvm-v1.0.0-alpha-linux-amd64`

- Make it executable and move to $PATH

`chmod +x /tmp/kcvm && sudo mv /tmp/kcvm /usr/local/bin`

## Commands

#### Initialize kcvm at ~/.kcvm

`kcvm init`

Add kcvm folder to your $PATH : `export PATH=$PATH:$HOME/.kcvm`

#### List available kubectl versions

`kcvm list remote`

#### List installed kubectl versions

`kcvm list local`

#### Install an available kubectl version

`kcvm install [VERSION]`

#### Use a specific installed kubectl version

`kcvm use [VERSION]`
