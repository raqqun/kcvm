package cmds

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2" // imports as package "cli"
)

// UseCmd declares use command
var UseCmd = cli.Command{
	Name:      "use",
	Usage:     "Specify kubectl version to use",
	ArgsUsage: "VERSION",
	Action: func(c *cli.Context) error {
		kubectlVersion, err := versionArg(c, "kcvm list local")
		if err != nil {
			return err
		}

		binPath := kubectlBinPath(kubectlVersion)

		_, err = os.Stat(binPath)
		if os.IsNotExist(err) {
			return fmt.Errorf("kubectl %s is not installed, run: kcvm install %s", kubectlVersion, kubectlVersion)
		}
		if err != nil {
			return err
		}

		err = activate(binPath)
		if err != nil {
			return err
		}

		log.Info("Now using kubectl " + kubectlVersion)

		return nil
	},
}
