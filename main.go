package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"dsab.slacker/cli"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		DisableColors:    false,
		DisableTimestamp: true,
		PadLevelText:     true,
	})

	if err := cli.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
