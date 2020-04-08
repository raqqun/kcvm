package cmds

import (
	"errors"
	"fmt"
	"github.com/urfave/cli/v2" // imports as package "cli"
	"os"
	"path"
)

// UseCmd declares list command
var UseCmd = cli.Command{
	Name:  "use",
	Usage: "Specify kubectl version to use",
	Action: func(c *cli.Context) error {
		if c.NArg() == 0 {
			return errors.New("You must provide a version from kcvm list local")
		}

		kubectlVersion := c.Args().Get(0)
		KubectlBinPath := path.Join(KubectlBinPath, fmt.Sprintf("kubectl-%s", kubectlVersion))
		kubectlSymPath := path.Join(KcvmPath, "kubectl")

		_, err := os.Stat(KubectlBinPath)
		if os.IsNotExist(err) {
			if err != nil {
				return errors.New("This kubectl version does not exist")
			}
		}

		_, err = os.Stat(kubectlSymPath)
		if !os.IsNotExist(err) {
			err = os.Remove(kubectlSymPath)
			if err != nil {
				return err
			}
		}

		err = os.Symlink(KubectlBinPath, kubectlSymPath)
		if err != nil {
			return err
		}

		return nil
	},
}
