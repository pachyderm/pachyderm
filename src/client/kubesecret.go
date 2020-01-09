package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pachyderm/pachyderm/src/client/pkg/config"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // enables support for configs with auth
	"k8s.io/client-go/rest"
)

// KubeSecretConfig handles the config for interacting with kubernetes secrets
type KubeSecretConfig struct {
	core      corev1.CoreV1Interface
	client    rest.Interface
	config    *rest.Config
	namespace string
}

// KubeSecretConfig creates a new docker registry
func NewKubeSecretConfig() (*KubeSecretConfig, error) {

	cfg, err := config.Read()
	if err != nil {
		return nil, fmt.Errorf("could not read config: %v", err)
	}
	_, context, err := cfg.ActiveContext()
	if err != nil {
		return nil, fmt.Errorf("could not get active context: %v", err)
	}

	kubeConfig := config.KubeConfig(context)

	kubeClientConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return nil, err
	}

	core := client.CoreV1()

	return &KubeSecretConfig{
		core:      core,
		client:    core.RESTClient(),
		config:    kubeClientConfig,
		namespace: "default",
	}, nil
}

// Create creates a new kubernetes secret
func (c *KubeSecretConfig) Create(file string) (string, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}
	var s v1.Secret
	err = json.Unmarshal(b, &s)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal: %v", err)
	}

	secret, err := c.core.Secrets(c.namespace).Create(&s)
	if err != nil {
		return "", fmt.Errorf("failed to create secret: %v", err)
	}
	return secret.GetObjectMeta().GetName(), nil
}
