package config

import (
	"k8s.io/client-go/tools/clientcmd"
)

// KubeConfig gets the kubernetes config
func KubeConfig() clientcmd.ClientConfig {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
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
