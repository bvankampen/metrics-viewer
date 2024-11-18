package config

type ApplicationConfig struct {
	Metrics  []string `yaml:"metrics"`
	Settings struct {
		ScrapeTimeout   int `yaml:"scrape_timeout"`
		ScrapeInterval  int `yaml:"scrape_interval"`
		RefreshInterval int `yaml:"refresh_interval"`
	} `yaml:"settings"`
}
