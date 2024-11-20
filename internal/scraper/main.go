package scraper

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/bvankampen/metrics-viewer/internal/config"
	"github.com/bvankampen/metrics-viewer/internal/kubeconfig"
	"github.com/bvankampen/metrics-viewer/internal/realtimedata"
	"k8s.io/client-go/rest"

	"github.com/urfave/cli"
)

func (s *Scraper) Init(ctx *cli.Context) {
	s.ctx = *ctx
	s.config = *config.LoadAppConfig(ctx.String("config"))
	s.restConfig = *kubeconfig.LoadKubeConfig(ctx.String("kubeconfig"))

	c, _ := rest.HTTPClientFor(&s.restConfig)

	s.httpClient = *c

	request, _ := http.NewRequest("GET", s.restConfig.Host+"/metrics", nil)
	request.Header.Add("Authorization", "Bearer "+s.restConfig.BearerToken)
	s.httpRequest = *request
}

func (s *Scraper) ScrapeInterval() int {
	return s.config.Settings.ScrapeInterval
}

func (s *Scraper) parse(metrics []byte) {
	scanner := bufio.NewScanner(bytes.NewReader(metrics))
	for scanner.Scan() {
		metricLine := scanner.Text()
		for _, m := range s.config.Metrics {
			if strings.HasPrefix(metricLine, "# HELP "+m) { // Description
				s.data.AddDescription(m, metricLine)
			}
			if strings.HasPrefix(metricLine, "# TYPE "+m) { // Metric Type
				s.data.AddType(m, metricLine)
			}
			if strings.HasPrefix(metricLine, m) { // Value
				s.data.AddValue(m, metricLine)
			}
		}
	}
}

func (s *Scraper) Scrape() (realtimedata.RealTimeData, error) {
	response, err := s.httpClient.Do(&s.httpRequest)
	if err != nil {
		return realtimedata.RealTimeData{}, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusOK {
		metrics, err := io.ReadAll(response.Body)
		if err != nil {
			return realtimedata.RealTimeData{}, err
		}
		s.parse(metrics)
	}
	return s.data, nil
}
