package config

import (
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const DEFAULT_APPLICATION_CONFIG_FILE = "./metrics-viewer.yaml"

// Load Application Config file
func LoadAppConfig(filename string) *ApplicationConfig {
	filename, _ = homedir.Expand(DEFAULT_APPLICATION_CONFIG_FILE)
	applicationConfig := ApplicationConfig{}

	logrus.Debugf("Loading configfile: %s", filename)

	yamlConfig, err := os.ReadFile(filename)
	if err != nil {
		logrus.Fatalf("Unable to load configuration file: %v", err)
	}

	err = yaml.Unmarshal(yamlConfig, &applicationConfig)
	if err != nil {
		logrus.Fatalf("Unable to load configuration: %v", err)
	}

	return &applicationConfig
}
