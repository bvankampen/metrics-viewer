package ui

import (
	"fmt"

	"github.com/bvankampen/metrics-viewer/internal/realtimedata"
	"github.com/urfave/cli"
)

func (u *UI) Init(ctx *cli.Context) {
	// placeholder init method
}

func (u *UI) UpdateScreen(data realtimedata.RealTimeData) { // placeholder function for testing
	for _, metric := range data.Metrics {
		fmt.Println(metric.Name)
		for _, value := range metric.Values {
			for _, label := range value.Labels {
				fmt.Printf("  %s : %s\n", label.Label, label.Value)
			}
			fmt.Printf("  %s\n", value.Value)
		}
	}
}
