package main

import (
	"errors"
	"os"

	"github.com/thedavidweng/flickr-cli/internal/cli"
	"github.com/thedavidweng/flickr-cli/internal/model"
)

func main() {
	if err := cli.Execute(); err != nil {
		if cmdErr, ok := errors.AsType[*model.CommandError](err); ok {
			os.Exit(model.ExitCode(cmdErr.Code))
		}
		os.Exit(1)
	}
}
