package cmds

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/go-resty/resty/v2"
	"github.com/urfave/cli/v2" // imports as package "cli"
)

// InstallCmd declares install command
var InstallCmd = cli.Command{
	Name:  "install",
	Usage: "Install kubectl version",
	Action: func(c *cli.Context) error {
		if c.NArg() == 0 {
			return errors.New("You must provide a version from kcvm list")
		}

		kubectlVersion := c.Args().Get(0)
		kubectlBinPath := path.Join(path.Join(KcvmPath, "bin"), fmt.Sprintf("kubectl-%s", kubectlVersion))
		kubectlTempPath := path.Join(path.Join(os.TempDir(), "kcvm"), fmt.Sprintf("kubectl-%s", kubectlVersion))
		kubectlSymPath := path.Join(KcvmPath, "kubectl")

		client := resty.New()

		resp, err := client.R().
			SetOutput(kubectlTempPath).
			Get(fmt.Sprintf("https://dl.k8s.io/release/%s/bin/%s/%s/kubectl", kubectlVersion, GOOS, GOARCH))

		if err != nil {
			return err
		}

		if resp.RawResponse.StatusCode != 200 {
			return errors.New("This kubectl version does not exist")
		}

		err = os.Chmod(kubectlTempPath, 0775)
		if err != nil {
			return err
		}

		err = os.Rename(kubectlTempPath, kubectlBinPath)
		if err != nil {
			return err
		}

		_, err = os.Stat(kubectlSymPath)
		if !os.IsNotExist(err) {
			err = os.Remove(kubectlSymPath)
			if err != nil {
				return err
			}
		}

		err = os.Symlink(kubectlBinPath, kubectlSymPath)
		if err != nil {
			return err
		}

		return nil
	},
}
