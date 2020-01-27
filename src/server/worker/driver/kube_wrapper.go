package driver

import (
	kube "k8s.io/client-go/kubernetes"

	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
)

// KubeWrapper is an interface for predicting the number of workers that the pipeline
// will have available.  It exists in particular to remove the worker driver's dependency
// on kube.Clientset to make testing more straightforward.
type KubeWrapper interface {
	GetExpectedNumWorkers(parallelism *pps.ParallelismSpec) (int, error)
}

type kubeWrapper struct {
	kubeClient *kube.Clientset
}

// NewKubeWrapper creates a KubeWrapper dependency from a kubeClient for use by a Driver
func NewKubeWrapper(kubeClient *kube.Clientset) KubeWrapper {
	return &kubeWrapper{
		kubeClient: kubeClient,
	}
}

func (kw *kubeWrapper) GetExpectedNumWorkers(parallelism *pps.ParallelismSpec) (int, error) {
	return ppsutil.GetExpectedNumWorkers(kw.kubeClient, parallelism)
}
