package main

import (
	"fmt"
	"os"

	"github.com/bvankampen/metrics-viewer/internal/rxgo"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	Version  = "0.0"
	CommitId = "dev"
)

func main() {
	app := cli.NewApp()
	app.Name = "metrics-viewer"
	app.Usage = "a Kubernetes Metrics Viewer"
	app.Authors = []cli.Author{
		{Name: "Volodymyr Katkalov", Email: "volodymyr.katkalov@suse.com"},
		{Name: "Bas van Kampen", Email: "bas.vankampen@suse.com"},
	}

	app.Version = fmt.Sprintf("%s (%s)", Version, CommitId)
	app.Before = func(ctx *cli.Context) error {
		if ctx.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
			logrus.Debugf("Loglevel set to [%v]", logrus.DebugLevel)
		}
		return nil
	}
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "Enable debug",
		},
		&cli.StringFlag{
			Name:   "kubeconfig",
			Usage:  "Kubeconfig file",
			Value:  "~/.kube/config",
			EnvVar: "KUBECONFIG",
		},
		&cli.StringFlag{
			Name:   "config",
			Usage:  "Config file",
			Value:  "~/.config/metrics-viewer.yaml",
			EnvVar: "METRICS_VIEWER_CONFIG",
		},
	}
	app.Action = rxgo.Run
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
