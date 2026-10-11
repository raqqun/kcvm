package cmds

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/urfave/cli/v2" // imports as package "cli"
)

var (
	UserHome string
	KcvmPath string

	// KubectlDownloadURL is the base URL kubectl release binaries are downloaded from
	KubectlDownloadURL = "https://dl.k8s.io/release"
	// KubernetesRepoURL is the git repository kubectl versions are listed from
	KubernetesRepoURL = "https://github.com/kubernetes/kubernetes.git"
)

const (
	// GOARCH os architecture
	GOARCH = runtime.GOARCH
	// GOOS os type
	GOOS = runtime.GOOS
)

// versionPattern matches the kubectl versions kcvm accepts, e.g. v1.20.0 or v1.21.0-rc.0
var versionPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+(-(alpha|beta|rc)\.\d+)?$`)

// Commands declare each command
var Commands = []*cli.Command{
	&InitCmd,
	&InstallCmd,
	&ListCmd,
	&UseCmd,
}

func InitCLI() *cli.App {

	UserHome, _ = os.UserHomeDir()
	KcvmPath = filepath.Join(UserHome, ".kcvm")

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
	// The kubectl symlink must point to an absolute path, whatever --dir looks like
	app.Before = func(c *cli.Context) error {
		absPath, err := filepath.Abs(KcvmPath)
		if err != nil {
			return err
		}
		KcvmPath = absPath
		return nil
	}

	return app
}

func binDir() string {
	return filepath.Join(KcvmPath, "bin")
}

func kubectlBinPath(kubectlVersion string) string {
	return filepath.Join(binDir(), "kubectl-"+kubectlVersion)
}

func kubectlSymPath() string {
	return filepath.Join(KcvmPath, "kubectl")
}

// versionArg returns the version given as first argument, normalized to the vX.Y.Z form
func versionArg(c *cli.Context, hint string) (string, error) {
	if c.NArg() == 0 {
		return "", fmt.Errorf("You must provide a version from %s", hint)
	}

	kubectlVersion := c.Args().Get(0)
	if !strings.HasPrefix(kubectlVersion, "v") {
		kubectlVersion = "v" + kubectlVersion
	}
	if !versionPattern.MatchString(kubectlVersion) {
		return "", fmt.Errorf("Invalid kubectl version %q, expected a version like v1.20.0", c.Args().Get(0))
	}

	return kubectlVersion, nil
}

// activeVersion returns the version the kubectl symlink points to, or "" if there is none
func activeVersion() string {
	target, err := os.Readlink(kubectlSymPath())
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(filepath.Base(target), "kubectl-")
}

// activate points the kubectl symlink to the given kubectl binary. The new symlink is
// renamed over the old one so that kubectl is never missing, even if kcvm is interrupted.
func activate(kubectlBinPath string) error {
	tempSymPath := fmt.Sprintf("%s.%d.tmp", kubectlSymPath(), os.Getpid())
	os.Remove(tempSymPath)

	err := os.Symlink(kubectlBinPath, tempSymPath)
	if err != nil {
		return err
	}

	err = os.Rename(tempSymPath, kubectlSymPath())
	if err != nil {
		os.Remove(tempSymPath)
		return err
	}

	return nil
}
