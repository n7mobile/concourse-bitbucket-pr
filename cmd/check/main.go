package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/n7mobile/ci-bitbucket-pr/resource"
	"github.com/n7mobile/ci-bitbucket-pr/resource/models"
)

func main() {
	fmt.Fprintln(os.Stderr, "calling check cmd")

	var request models.CheckRequest

	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err.Error())
		os.Exit(1)
	}

	logger := resource.Logger{
		Debug: request.Source.Debug,
	}

	command := resource.CheckCommand{
		Logger: &logger,
	}

	versions, err := command.Run(request)
	if err != nil {
		fmt.Fprintln(os.Stderr, "running command:", err.Error())
		os.Exit(1)
	}

	json.NewEncoder(os.Stdout).Encode(versions)
}
