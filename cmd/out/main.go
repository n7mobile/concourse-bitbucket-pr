package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/n7mobile/ci-bitbucket-pr/resource"
	"github.com/n7mobile/ci-bitbucket-pr/resource/models"
)

func main() {
	fmt.Fprintln(os.Stderr, "calling in cmd")

	if len(os.Args) < 2 {
		println("usage: " + os.Args[0] + " <destination>")
		os.Exit(1)
	}

	destination := os.Args[1]

	fmt.Fprintln(os.Stderr, "calling out cmd wioth dest ", destination)

	var request models.OutRequest

	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err.Error())
		os.Exit(1)
	}

	logger := resource.Logger{
		Debug: request.Source.Debug,
	}

	command := resource.OutCommand{
		Logger: logger,
	}

	fmt.Fprintln(os.Stderr, "RUUUUUUN")

	response, err := command.Run(request, destination)
	if err != nil {
		logger.Errorf("running command: %w", err)
		os.Exit(1)
	}

	json.NewEncoder(os.Stdout).Encode(response)
}
