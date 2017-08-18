package util

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"

	"github.com/pachyderm/pachyderm/src/client/pps"
)

// GetResourceListFromPipeline returns a list of resources that the pipeline,
// minimally requires.
func GetResourceListFromPipeline(pipelineInfo *pps.PipelineInfo) (*api.ResourceList, error) {
	var result api.ResourceList = make(map[api.ResourceName]resource.Quantity)
	resources, cacheSize := pipelineInfo.ResourceSpec, pipelineInfo.CacheSize
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
