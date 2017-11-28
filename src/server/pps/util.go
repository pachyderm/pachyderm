package pps

import (
	"fmt"
	"math"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
)

// JobRepo creates a pfs repo for a given job.
func JobRepo(job *ppsclient.Job) *pfs.Repo {
	return &pfs.Repo{Name: fmt.Sprintf("job_%s", job.ID)}
}

// PipelineRepo creates a pfs repo for a given pipeline.
func PipelineRepo(pipeline *ppsclient.Pipeline) *pfs.Repo {
	return &pfs.Repo{Name: pipeline.Name}
}

// PipelineRcName generates the name of the k8s replication controller that
// manages a pipeline's workers
func PipelineRcName(name string, version uint64) string {
	// k8s won't allow RC names that contain upper-case letters
	// or underscores
	// TODO: deal with name collision
	name = strings.Replace(name, "_", "-", -1)
	return fmt.Sprintf("pipeline-%s-v%d", strings.ToLower(name), version)
}

// getNumNodes attempts to retrieve the number of nodes in the current k8s
// cluster
func getNumNodes(kubeClient *kube.Clientset) (int, error) {
	nodeList, err := kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("unable to retrieve node list from k8s to determine parallelism: %v", err)
	}
	if len(nodeList.Items) == 0 {
		return 0, fmt.Errorf("pachyderm.pps.jobserver: no k8s nodes found")
	}
	return len(nodeList.Items), nil
}

// GetExpectedNumWorkers computes the expected number of workers that
// pachyderm will start given the ParallelismSpec 'spec'.
//
// This is only exported for testing
func GetExpectedNumWorkers(kubeClient *kube.Clientset, spec *ppsclient.ParallelismSpec) (int, error) {
	if spec == nil || (spec.Constant == 0 && spec.Coefficient == 0) {
		return 1, nil
	} else if spec.Constant > 0 && spec.Coefficient == 0 {
		return int(spec.Constant), nil
	} else if spec.Constant == 0 && spec.Coefficient > 0 {
		// Start ('coefficient' * 'nodes') workers. Determine number of workers
		numNodes, err := getNumNodes(kubeClient)
		if err != nil {
			return 0, err
		}
		result := math.Floor(spec.Coefficient * float64(numNodes))
		return int(math.Max(result, 1)), nil
	}
	return 0, fmt.Errorf("Unable to interpret ParallelismSpec %+v", spec)
}
