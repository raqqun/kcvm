package cmds

import (
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2" // imports as package "cli"
)

// InitCmd creates kcvm main directory under ~/.kcvm
var InitCmd = cli.Command{
	Name:  "init",
	Usage: "Initialize kcvm",
	Action: func(c *cli.Context) error {
		_, err := os.Stat(path.Join(KcvmPath, "bin"))
		if os.IsNotExist(err) {
			err = os.MkdirAll(path.Join(KcvmPath, "bin"), 0755)
			if err != nil {
				return err
			}

			log.Info("Path " + KcvmPath + " Created")

			return nil
		}

		log.Info("Path " + KcvmPath + " Exists")

		return nil
	},
}
