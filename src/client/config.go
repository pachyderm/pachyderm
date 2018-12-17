package client

import (
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
)

func BuildRESTConfig() (*rest.Config, error) {
    rules := clientcmd.NewDefaultClientConfigLoadingRules()
    overrides := &clientcmd.ConfigOverrides{}
    kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
    return kubeConfig.ClientConfig()
}
