package util

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"

	"github.com/pachyderm/pachyderm/src/client/pps"
)

// GetResourceListFromPipeline returns a list of resources that the pipeline,
// minimally requires.
func GetResourceListFromPipeline(pipelineInfo *pps.PipelineInfo) (*api.ResourceList, error) {
	resources, cacheSize := pipelineInfo.ResourceSpec, pipelineInfo.CacheSize
	cpuQuantity, err := resource.ParseQuantity(fmt.Sprintf("%f", resources.Cpu))
	if err != nil {
		return nil, fmt.Errorf("could not parse cpu quantity: %s", err)
	}

	memQuantity, err := resource.ParseQuantity(resources.Memory)
	if err != nil {
		return nil, fmt.Errorf("could not parse memory quantity: %s", err)
	}
	cacheQuantity, err := resource.ParseQuantity(cacheSize)
	if err != nil {
		return nil, fmt.Errorf("could not parse cache quantity: %s", err)
	}
	// Here we are sanity checking.  A pipeline should request at least
	// as much memory as it needs for caching.
	if cacheQuantity.Cmp(memQuantity) > 0 {
		memQuantity = cacheQuantity
	}

	gpuQuantity, err := resource.ParseQuantity(fmt.Sprintf("%d", resources.Gpu))
	if err != nil {
		return nil, fmt.Errorf("could not parse gpu quantity: %s", err)
	}
	var result api.ResourceList = map[api.ResourceName]resource.Quantity{
		api.ResourceCPU:       cpuQuantity,
		api.ResourceMemory:    memQuantity,
		api.ResourceNvidiaGPU: gpuQuantity,
	}
	return &result, nil
}
