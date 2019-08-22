package config

import (
	"sync"

	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeConfigOnce  sync.Once
	kubeConfigValue clientcmd.ClientConfig
)

// ReadKubeConfig gets the kubernetes config
func ReadKubeConfig() clientcmd.ClientConfig {
	kubeConfig.Do(func() error {
		rules := clientcmd.NewDefaultClientConfigLoadingRules()
		overrides := &clientcmd.ConfigOverrides{}
		kubeConfigValue = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
	})
	return kubeConfigValue
}

// KubeNamespace gets the default namespace for the active context
func KubeNamespace(kubeConfig clientcmd.ClientConfig) (string, error) {
	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		return "", err
	}
	if namespace == "" {
		return "default", nil
	}
	return namespace, nil
}
