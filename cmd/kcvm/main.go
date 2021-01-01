package main

import (
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

func main() {
	kcvm := cmds.InitCLI()

	err := kcvm.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
