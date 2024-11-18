package scraper

import (
	"net/http"

	"github.com/bvankampen/metrics-viewer/internal/config"
	"github.com/urfave/cli"
	"k8s.io/client-go/tools/clientcmd/api"
)

type Scraper struct {
	config      config.ApplicationConfig
	kubeConfig  api.Config
	ctx         cli.Context
	httpClient  http.Client
	httpRequest http.Request
}
