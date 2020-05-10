package cmds

import (
	"github.com/urfave/cli/v2" // imports as package "cli"
	"os"
	"path"
	"runtime"
)

var (
	UserHome, _        = os.UserHomeDir()
	KcvmPath           = path.Join(UserHome, ".kcvm")
	KubectlBinPath     = path.Join(UserHome, ".kcvm", "bin")
	KubectlBinTempPath = path.Join(os.TempDir(), "kcvm")
)

const (
	// GOARCH os architecture
	GOARCH = runtime.GOARCH
	// GOOS os type
	GOOS = runtime.GOOS
)

// Commands declare each command
var Commands = []*cli.Command{
	&InitCmd,
	&InstallCmd,
	&ListCmd,
	&UseCmd,
}
