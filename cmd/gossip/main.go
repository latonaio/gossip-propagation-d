package main

import (
	"os"

	"bitbucket.org/latonaio/gossip-propagation-d/cmd/gossip/app"
)

func main() {
	command := app.NewGossipPropagationCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
