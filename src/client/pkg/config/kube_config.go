package config

import (
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
