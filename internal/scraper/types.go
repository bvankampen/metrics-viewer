package scraper

import (
	"net/http"

	"github.com/bvankampen/metrics-viewer/internal/config"
	"github.com/bvankampen/metrics-viewer/internal/realtimedata"
	"github.com/urfave/cli"
	"k8s.io/client-go/rest"
)

type Scraper struct {
	config      config.ApplicationConfig
	restConfig  rest.Config
	ctx         cli.Context
	httpClient  http.Client
	httpRequest http.Request
	data        realtimedata.RealTimeData
}
