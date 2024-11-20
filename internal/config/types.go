package config

type ApplicationConfig struct {
	Metrics  []string `yaml:"metrics"`
	Settings struct {
		ScrapeInterval int `yaml:"scrape_interval"`
	} `yaml:"settings"`
}
