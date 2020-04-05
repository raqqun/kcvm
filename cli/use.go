package cmds

import (
	"github.com/urfave/cli/v2" // imports as package "cli"
)

// UseCmd declares list command
var UseCmd = cli.Command{
	Name:  "use",
	Usage: "Specify kubectl version to use",
	Action: func(c *cli.Context) error {

		return nil
	},
}
