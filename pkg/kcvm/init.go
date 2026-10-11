package cmds

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2" // imports as package "cli"
)

// InitCmd creates kcvm main directory under ~/.kcvm
var InitCmd = cli.Command{
	Name:  "init",
	Usage: "Initialize kcvm",
	Action: func(c *cli.Context) error {
		_, err := os.Stat(binDir())
		if os.IsNotExist(err) {
			err = os.MkdirAll(binDir(), 0755)
			if err != nil {
				return err
			}

			log.Info("Path " + KcvmPath + " Created")
		} else if err != nil {
			return err
		} else {
			log.Info("Path " + KcvmPath + " Exists")
		}

		if !inPath(KcvmPath) {
			log.Warn("Add kcvm to your $PATH: export PATH=" + KcvmPath + ":$PATH")
		}

		return nil
	},
}

func inPath(dir string) bool {
	for _, entry := range filepath.SplitList(os.Getenv("PATH")) {
		if filepath.Clean(entry) == dir {
			return true
		}
	}
	return false
}
