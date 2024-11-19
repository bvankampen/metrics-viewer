package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bvankampen/metrics-viewer/internal/scraper"
	"github.com/bvankampen/metrics-viewer/internal/ui"
	"github.com/reactivex/rxgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	Version  = "0.0"
	CommitId = "dev"
)

// Transform the pipeline data
// transformed := pipeline.
// 	Map(ui.SafeTransform(func(value interface{}, root map[string]interface{}) (interface{}, error) {
// 		// Transform the car name
// 		str, ok := value.(string)
// 		if !ok {
// 			return nil, fmt.Errorf("value is not a string: %v", value)
// 		}
// 		return str + " Jr.", nil
// 	}, "data.cars[0].details.name")).
// 	Map(ui.SafeTransform(func(value interface{}, root map[string]interface{}) (interface{}, error) {
// 		// Convert timestamp to human-readable format
// 		currentTime, ok := root["currentTime"].(int64)
// 		if !ok {
// 			return nil, fmt.Errorf("currentTime is missing or invalid in root structure")
// 		}
// 		timestamp, ok := value.(int64)
// 		if !ok {
// 			return nil, fmt.Errorf("value is not a valid timestamp: %v", value)
// 		}
// 		return ui.ToHumanAgo(timestamp, currentTime)
// 	}, "data.cars[0].details.timestamp")).
// 	Map(ui.SafeTransform(func(value interface{}, root map[string]interface{}) (interface{}, error) {
// 		// Format the design date
// 		designedOn, ok := value.(int64)
// 		if !ok {
// 			return nil, fmt.Errorf("value is not a valid timestamp: %v", value)
// 		}
// 		return ui.ToISO8601(designedOn), nil
// 	}, "data.cars[0].details.designed_on"))

// transformed.DoOnNext(func(i interface{}) {
// 	fmt.Printf("Pipeline Output: %+v\n", i)
// })

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

	// Initialize scraper
	scraper := scraper.Scraper{}
	scraper.Init(ctx)
	// Timer observable
	timer := rxgo.Interval(rxgo.WithDuration(time.Duration(scraper.ScrapeInterval()) * time.Second)).
		Map(func(ctx context.Context, _ interface{}) (interface{}, error) {
			return time.Now().Unix(), nil
		})

	// Initialize UI TableView
	tv := ui.NewTableView(nil)

	// Create a scraper observable
	dataSource := rxgo.Create([]rxgo.Producer{
		func(ctx context.Context, ch chan<- rxgo.Item) {
			for {
				data, err := scraper.Scrape()
				if err != nil {
					ch <- rxgo.Error(err)
					logrus.Errorf("Error scraping metrics: %v", err)
					continue
				}
				ch <- rxgo.Of(data)
				time.Sleep(1 * time.Second) // Adjust scrape interval as needed
			}
		},
	})

	// Combine dataSource and timer into a pipeline
	pipeline := rxgo.CombineLatest(
		func(i ...interface{}) interface{} {
			return map[string]interface{}{
				"data":        i[0],
				"currentTime": i[1],
			}
		},
		[]rxgo.Observable{dataSource, timer},
	)

	// Transform the RxGo channel for TableView
	observeChan := make(chan interface{})
	go func() {
		for item := range pipeline.Observe() {
			if item.E != nil {
				fmt.Println("Error in pipeline.Observe():", item.E)
				continue
			}
			vMap, ok := item.V.(map[string]interface{})
			if !ok {
				fmt.Println("Error: unexpected data format in pipeline:", item.V)
				continue
			}

			// Extract "data" from combined map and send to observeChan
			data := vMap["data"]
			observeChan <- data
		}
		fmt.Println("pipeline.Observe() channel closed")
		close(observeChan)
	}()

	// Run the TableView with dynamic updates
	tv.Run(observeChan)
}
