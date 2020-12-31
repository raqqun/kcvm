package main

import (
	"fmt"
	"os"

	cmds "github.com/raqqun/kcvm/pkg/kcvm"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2" // imports as package "cli"
)

var (
	Version string
	CommitID string
)

func init() {

	log.SetFormatter(&log.TextFormatter{})

	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
}

func main() {
	kcvm := cli.NewApp()
	kcvm.EnableBashCompletion = true

	kcvm.Name = "kcvm"
	kcvm.HelpName = kcvm.Name
	kcvm.Usage = "A kubectl version manager!"
	kcvm.Version = fmt.Sprintf("%s - %s", Version, CommitID)

	kcvm.Commands = cmds.Commands

	err := kcvm.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
