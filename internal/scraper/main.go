package scraper

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bvankampen/metrics-viewer/internal/config"
	"github.com/bvankampen/metrics-viewer/internal/kubeconfig"
	"github.com/sirupsen/logrus"

	"github.com/urfave/cli"
	// "github.com/yassinebenaid/godump"
)

func (s *Scraper) Init(ctx *cli.Context) {
	s.ctx = *ctx
	s.config = *config.LoadAppConfig(ctx.String("config"))
	// godump.Dump(s.Config)
	s.kubeConfig = *kubeconfig.LoadKubeConfig(ctx.String("kubeconfig"))

	s.httpClient = http.Client{}

	url, token := s.getUrlAndToken()

	url = fmt.Sprintf("%s/metrics", url)

	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("Authorization", "Bearer "+token)
	s.httpRequest = *request
}

func (s *Scraper) getUrlAndToken() (string, string) {
	currentContext := s.kubeConfig.Contexts[s.kubeConfig.CurrentContext]
	cluster := s.kubeConfig.Clusters[currentContext.Cluster]
	authInfo := s.kubeConfig.AuthInfos[currentContext.AuthInfo]
	return cluster.Server, authInfo.Token
}

func (s *Scraper) Run() {
	err := s.Scrape()
	if err != nil {
		logrus.Fatalf("Error scraping metrics: %v", err)
	}
}

func (s *Scraper) parse(metrics []byte) {
	scanner := bufio.NewScanner(bytes.NewReader(metrics))
	for scanner.Scan() {
		metricLine := scanner.Text()
		for _, m := range s.config.Metrics {
			if strings.HasPrefix(metricLine, "# HELP "+m) {
				fmt.Println(metricLine)
			}
			if strings.HasPrefix(metricLine, "# TYPE "+m) {
				fmt.Println(metricLine)
			}

			if strings.HasPrefix(metricLine, m) {
				fmt.Println(metricLine)
			}
		}
	}
}

func (s *Scraper) Scrape() error {
	response, err := s.httpClient.Do(&s.httpRequest)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusOK {
		metrics, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		s.parse(metrics)
	}
	return nil
}
