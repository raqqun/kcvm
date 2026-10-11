package main

import (
	"fmt"
	"os"

	"github.com/raqqun/kcvm"
	cmds "github.com/raqqun/kcvm/pkg/kcvm"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
}

// Set at build time with -ldflags "-X main.commitSHA=..."
var commitSHA = "unknown"

func main() {
	app := cmds.InitCLI()
	app.Version = fmt.Sprintf("%s-%s", kcvm.Version, commitSHA)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
