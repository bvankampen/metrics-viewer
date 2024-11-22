package config

import (
	"errors"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Load Application Config file
func LoadAppConfig(filename string) *ApplicationConfig {
	filename, _ = homedir.Expand(filename)
	applicationConfig := ApplicationConfig{}

	_, err := os.Stat(filename) // create new config is not exists
	if errors.Is(err, os.ErrNotExist) {
		err := os.WriteFile(filename, []byte(DEFAULT_CONFIG), 0600)
		if err != nil {
			logrus.Fatalf("Unable to create new configuration file: %v", err)
		}

	}

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
