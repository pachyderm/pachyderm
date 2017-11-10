package util

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strings"

	etcd "github.com/coreos/etcd/clientv3"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"

	"github.com/pachyderm/pachyderm/src/client/pps"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

// GetRequestsResourceListFromPipeline returns a list of resources that the pipeline,
// minimally requires.
func GetRequestsResourceListFromPipeline(pipelineInfo *pps.PipelineInfo) (*api.ResourceList, error) {
	return getResourceListFromSpec(pipelineInfo.ResourceRequestsSpec, pipelineInfo.CacheSize)
}

func getResourceListFromSpec(resources *pps.ResourceSpec, cacheSize string) (*api.ResourceList, error) {
	var result api.ResourceList = make(map[api.ResourceName]resource.Quantity)
	cpuStr := fmt.Sprintf("%f", resources.Cpu)
	cpuQuantity, err := resource.ParseQuantity(cpuStr)
	if err != nil {
		log.Warnf("error parsing cpu string: %s: %+v", cpuStr, err)
	} else {
		result[api.ResourceCPU] = cpuQuantity
	}

	memQuantity, err := resource.ParseQuantity(resources.Memory)
	if err != nil {
		log.Warnf("error parsing memory string: %s: %+v", resources.Memory, err)
	} else {
		result[api.ResourceMemory] = memQuantity
	}

	// Here we are sanity checking.  A pipeline should request at least
	// as much memory as it needs for caching.
	cacheQuantity, err := resource.ParseQuantity(cacheSize)
	if err != nil {
		log.Warnf("error parsing cache string: %s: %+v", cacheSize, err)
	} else if cacheQuantity.Cmp(memQuantity) > 0 {
		result[api.ResourceMemory] = cacheQuantity
	}

	gpuStr := fmt.Sprintf("%d", resources.Gpu)
	gpuQuantity, err := resource.ParseQuantity(gpuStr)
	if err != nil {
		log.Warnf("error parsing gpu string: %s: %+v", gpuStr, err)
	} else {
		result[api.ResourceNvidiaGPU] = gpuQuantity
	}
	return &result, nil
}

// GetLimitsResourceListFromPipeline returns a list of resources that the pipeline,
// maximally is limited to.
func GetLimitsResourceListFromPipeline(pipelineInfo *pps.PipelineInfo) (*api.ResourceList, error) {
	return getResourceListFromSpec(pipelineInfo.ResourceLimitsSpec, pipelineInfo.CacheSize)
}

// LookupUser is a reimplementation of user.Lookup that doesn't require cgo.
func LookupUser(name string) (_ *user.User, retErr error) {
	passwd, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := passwd.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	scanner := bufio.NewScanner(passwd)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if parts[0] == name {
			return &user.User{
				Username: parts[0],
				Uid:      parts[2],
				Gid:      parts[3],
				Name:     parts[4],
				HomeDir:  parts[5],
			}, nil
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return nil, fmt.Errorf("user %s not found", name)
}

// FailPipeline updates the pipeline's state to failed and sets the failure reason
func FailPipeline(ctx context.Context, etcdClient *etcd.Client, pipelinesCollection col.Collection, pipelineName string, reason string) error {

	_, err := col.NewSTM(ctx, etcdClient, func(stm col.STM) error {
		pipelines := pipelinesCollection.ReadWrite(stm)
		pipelineInfo := new(pps.PipelineInfo)
		if err := pipelines.Get(pipelineName, pipelineInfo); err != nil {
			return err
		}
		pipelineInfo.State = pps.PipelineState_PIPELINE_FAILURE
		pipelineInfo.Reason = reason
		pipelines.Put(pipelineName, pipelineInfo)
		return nil
	})
	return err
}
