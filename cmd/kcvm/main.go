package main

import (
	"fmt"
	"os"

	cmds "github.com/raqqun/kcvm/pkg/kcvm"
	log "github.com/sirupsen/logrus"
)

func init() {

	log.SetFormatter(&log.TextFormatter{})

	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
}

var (
	version   string
	commitSHA string
)

func main() {
	version := fmt.Sprintf("%s-%s", version, commitSHA)
	kcvm := cmds.InitCLI()
	kcvm.Version = version

	err := kcvm.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
