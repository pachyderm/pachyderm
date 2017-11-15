package server

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"

	"github.com/pachyderm/pachyderm/src/client/deploy"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"
)

type apiServer struct {
	kubeClient    *kube.Clientset
	kubeNamespace string
}

// NewDeployServer creates a deploy server
func NewDeployServer(kubeClient *kube.Clientset, kubeNamespace string) deploy.APIServer {
	return &apiServer{
		kubeClient:    kubeClient,
		kubeNamespace: kubeNamespace,
	}
}

func (s *apiServer) DeployStorageSecret(ctx context.Context, req *deploy.DeployStorageSecretRequest) (*deploy.DeployStorageSecretResponse, error) {
	kubeSecrets := s.kubeClient.CoreV1().Secrets(s.kubeNamespace)

	secret, err := kubeSecrets.Get(client.StorageSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error retrieving secret from kubernetes: %v", err)
	}

	// merge the secrets
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	for key, val := range req.Secrets {
		secret.Data[key] = val
	}

	if _, err := kubeSecrets.Update(secret); err != nil {
		return nil, fmt.Errorf("error updating secret: %v", err)
	}

	return &deploy.DeployStorageSecretResponse{}, nil
}
