package main

import (
	"fmt"
	"os"
	"time"

	"github.com/bvankampen/metrics-viewer/internal/scraper"
	"github.com/bvankampen/metrics-viewer/internal/ui"
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
		if ctx.GlobalBool("debug") {
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
			Value:  "metrics-viewer.yaml",
			EnvVar: "METRICS_VIEWER_CONFIG",
		},
	}
	app.Action = run
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(ctx *cli.Context) {
	// initialize scraper
	scraper := scraper.Scraper{}
	scraper.Init(ctx)
	scrapeTicker := time.NewTicker(time.Second * time.Duration(scraper.ScrapeInterval()))

	// initialize UI
	ui := ui.UI{}
	ui.Init(ctx)

	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-scrapeTicker.C:
				data, err := scraper.Scrape()
				if err != nil {
					logrus.Errorf("Error scraping metrics: %v", err)
				}
				ui.UpdateScreen(data)
			}
		}
	}()

	fmt.Println("Wait for enter key...")
	fmt.Scanln()
	scrapeTicker.Stop()
	done <- true
}
