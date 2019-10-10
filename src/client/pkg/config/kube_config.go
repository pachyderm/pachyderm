package config

import (
	"fmt"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// KubeConfig gets the kubernetes config
func KubeConfig(context *Context) clientcmd.ClientConfig {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}
	if context != nil {
		overrides.Context = api.Context{
			LocationOfOrigin: "pachyderm",
			Cluster:          context.ClusterName,
			AuthInfo:         context.AuthInfo,
			Namespace:        context.Namespace,
		}
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
}

// RawKubeConfig gets the "raw" kube config, e.g. without overrides
func RawKubeConfig() (api.Config, error) {
	kubeConfig := KubeConfig(nil)
	kubeRawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return api.Config{}, fmt.Errorf("could not read raw kube config: %v", err)
	}
	return kubeRawConfig, nil
}
