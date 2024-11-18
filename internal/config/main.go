package config

import (
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Load Application Config file
func LoadAppConfig(filename string) *ApplicationConfig {
	filename, _ = homedir.Expand(filename)
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
