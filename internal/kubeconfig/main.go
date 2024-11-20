package kubeconfig

import (
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func LoadKubeConfig(filename string) *rest.Config {
	filename, _ = homedir.Expand(filename)
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", filename)

	logrus.Debugf("Loading kubeconfig file: %s", filename)

	if err != nil {
		logrus.Fatalf("Error loading kubeconfig %s: %v", filename, err)
	}
	return kubeConfig
}
