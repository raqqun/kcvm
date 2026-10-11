// Package kcvm holds the kcvm version, the commands live in pkg/kcvm
package kcvm

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var version string

// Version is the kcvm version, bump it in the VERSION file for each release
var Version = strings.TrimSpace(version)
