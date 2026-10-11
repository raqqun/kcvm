package main

import (
	"fmt"
	"os"

	cmds "github.com/raqqun/kcvm/pkg/kcvm"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
}

// Set at build time with -ldflags "-X main.version=... -X main.commitSHA=..."
var (
	version   = "dev"
	commitSHA = "unknown"
)

func main() {
	kcvm := cmds.InitCLI()
	kcvm.Version = fmt.Sprintf("%s-%s", version, commitSHA)

	err := kcvm.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
