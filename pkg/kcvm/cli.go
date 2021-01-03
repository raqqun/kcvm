package cmds

import (
	"github.com/urfave/cli/v2" // imports as package "cli"
	"os"
	"path"
	"runtime"
)

var (
	UserHome string
	KcvmPath string
	// KubectlBinPath     string
	// KubectlBinTempPath string
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

func InitCLI() *cli.App {

	UserHome, _ = os.UserHomeDir()
	KcvmPath = path.Join(UserHome, ".kcvm")
	// KubectlBinPath = path.Join(KcvmPath, "bin")
	// KubectlBinTempPath = path.Join(os.TempDir(), "kcvm")

	app := cli.NewApp()
	app.EnableBashCompletion = true

	app.Name = "kcvm"
	app.HelpName = app.Name
	app.Usage = "A kubectl version manager!"
	app.Commands = Commands
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "dir",
			Value:       KcvmPath,
			Usage:       "home directory for kcvm",
			Destination: &KcvmPath,
		},
	}

	return app
}
