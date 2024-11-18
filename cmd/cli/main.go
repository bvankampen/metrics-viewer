package main

import (
	"fmt"

	"github.com/urfave/cli"
)

var (
	Version  = "0.0"
	CommitId = "dev"
)

func main() {
	app := cli.NewApp()
	app.Name = "metrics-viewer"
	app.Version = fmt.Sprintf("%s (%s)", Version, CommitId)
}
