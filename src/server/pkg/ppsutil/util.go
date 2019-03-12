// Package ppsutil contains utilities for various PPS-related tasks, which are
// shared by both the PPS API and the worker binary. These utilities include:
// - Getting the RC name and querying k8s reguarding pipelines
// - Reading and writing pipeline resource requests and limits
// - Reading and writing EtcdPipelineInfos and PipelineInfos[1]
//
// [1] Note that PipelineInfo in particular is complicated because it contains
// fields that are not always set or are stored in multiple places
// ('job_state', for example, is not stored in PFS along with the rest of each
// PipelineInfo, because this field is volatile and we cannot commit to PFS
// every time it changes. 'job_counts' is the same, and 'reason' is in etcd
// because it is only updated alongside 'job_state').  As of 12/7/2017, these
// are the only fields not stored in PFS.
package ppsutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"

	etcd "github.com/coreos/etcd/clientv3"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"
)

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

// GetRequestsResourceListFromPipeline returns a list of resources that the pipeline,
// minimally requires.
func GetRequestsResourceListFromPipeline(pipelineInfo *pps.PipelineInfo) (*v1.ResourceList, error) {
	return getResourceListFromSpec(pipelineInfo.ResourceRequests, pipelineInfo.CacheSize)
}

func getResourceListFromSpec(resources *pps.ResourceSpec, cacheSize string) (*v1.ResourceList, error) {
	var result v1.ResourceList = make(map[v1.ResourceName]resource.Quantity)
	cpuStr := fmt.Sprintf("%f", resources.Cpu)
	cpuQuantity, err := resource.ParseQuantity(cpuStr)
	if err != nil {
		log.Warnf("error parsing cpu string: %s: %+v", cpuStr, err)
	} else {
		result[v1.ResourceCPU] = cpuQuantity
	}

	memQuantity, err := resource.ParseQuantity(resources.Memory)
	if err != nil {
		log.Warnf("error parsing memory string: %s: %+v", resources.Memory, err)
	} else {
		result[v1.ResourceMemory] = memQuantity
	}

	if resources.Disk != "" { // needed because not all versions of k8s support disk resources
		diskQuantity, err := resource.ParseQuantity(resources.Disk)
		if err != nil {
			log.Warnf("error parsing disk string: %s: %+v", resources.Disk, err)
		} else {
			result[v1.ResourceStorage] = diskQuantity
		}
	}

	// Here we are sanity checking.  A pipeline should request at least
	// as much memory as it needs for caching.
	cacheQuantity, err := resource.ParseQuantity(cacheSize)
	if err != nil {
		log.Warnf("error parsing cache string: %s: %+v", cacheSize, err)
	} else if cacheQuantity.Cmp(memQuantity) > 0 {
		result[v1.ResourceMemory] = cacheQuantity
	}

	if resources.Gpu != nil {
		gpuStr := fmt.Sprintf("%d", resources.Gpu.Number)
		gpuQuantity, err := resource.ParseQuantity(gpuStr)
		if err != nil {
			log.Warnf("error parsing gpu string: %s: %+v", gpuStr, err)
		} else {
			result[v1.ResourceName(resources.Gpu.Type)] = gpuQuantity
		}
	}

	return &result, nil
}

// GetLimitsResourceListFromPipeline returns a list of resources that the pipeline,
// maximally is limited to.
func GetLimitsResourceListFromPipeline(pipelineInfo *pps.PipelineInfo) (*v1.ResourceList, error) {
	return getResourceListFromSpec(pipelineInfo.ResourceLimits, pipelineInfo.CacheSize)
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

// GetExpectedNumHashtrees computes the expected number of hashtrees that
// Pachyderm will create given the HashtreeSpec 'spec'.
func GetExpectedNumHashtrees(spec *ppsclient.HashtreeSpec) (int64, error) {
	if spec == nil || spec.Constant == 0 {
		return 1, nil
	} else if spec.Constant > 0 {
		return int64(spec.Constant), nil
	}
	return 0, fmt.Errorf("unable to interpret HashtreeSpec %+v", spec)
}

// GetPipelineInfo retrieves and returns a valid PipelineInfo from PFS. It does
// the PFS read/unmarshalling of bytes as well as filling in missing fields
func GetPipelineInfo(pachClient *client.APIClient, ptr *pps.EtcdPipelineInfo, full bool) (*pps.PipelineInfo, error) {
	result := &pps.PipelineInfo{}
	if full {
		buf := bytes.Buffer{}
		if err := pachClient.GetFile(ppsconsts.SpecRepo, ptr.SpecCommit.ID, ppsconsts.SpecFile, 0, 0, &buf); err != nil {
			return nil, fmt.Errorf("could not read existing PipelineInfo from PFS: %v", err)
		}
		if err := result.Unmarshal(buf.Bytes()); err != nil {
			return nil, fmt.Errorf("could not unmarshal PipelineInfo bytes from PFS: %v", err)
		}
	}
	result.State = ptr.State
	result.Reason = ptr.Reason
	result.JobCounts = ptr.JobCounts
	result.LastJobState = ptr.LastJobState
	result.SpecCommit = ptr.SpecCommit
	return result, nil
}

// FailPipeline updates the pipeline's state to failed and sets the failure reason
func FailPipeline(ctx context.Context, etcdClient *etcd.Client, pipelinesCollection col.Collection, pipelineName string, reason string) error {
	_, err := col.NewSTM(ctx, etcdClient, func(stm col.STM) error {
		pipelines := pipelinesCollection.ReadWrite(stm)
		pipelinePtr := new(pps.EtcdPipelineInfo)
		if err := pipelines.Get(pipelineName, pipelinePtr); err != nil {
			return err
		}
		pipelinePtr.State = pps.PipelineState_PIPELINE_FAILURE
		pipelinePtr.Reason = reason
		pipelines.Put(pipelineName, pipelinePtr)
		return nil
	})
	return err
}

// JobInput fills in the commits for a JobInfo. branchToCommit should map strings of the form "<repo>/<branch>" to PFS commits
func JobInput(pipelineInfo *pps.PipelineInfo, branchToCommit map[string]*pfs.Commit) *pps.Input {
	key := path.Join
	jobInput := proto.Clone(pipelineInfo.Input).(*pps.Input)
	pps.VisitInput(jobInput, func(input *pps.Input) {
		if input.Atom != nil {
			if commit, ok := branchToCommit[key(input.Atom.Repo, input.Atom.Branch)]; ok {
				input.Atom.Commit = commit.ID
			}
		}
		if input.Pfs != nil {
			if commit, ok := branchToCommit[key(input.Pfs.Repo, input.Pfs.Branch)]; ok {
				input.Pfs.Commit = commit.ID
			}
		}
		if input.Cron != nil {
			if commit, ok := branchToCommit[key(input.Cron.Repo, "master")]; ok {
				input.Cron.Commit = commit.ID
			}
		}
		if input.Git != nil {
			if commit, ok := branchToCommit[key(input.Git.Name, input.Git.Branch)]; ok {
				input.Git.Commit = commit.ID
			}
		}
	})
	return jobInput
}

// PipelineReqFromInfo converts a PipelineInfo into a CreatePipelineRequest.
func PipelineReqFromInfo(pipelineInfo *ppsclient.PipelineInfo) *ppsclient.CreatePipelineRequest {
	return &ppsclient.CreatePipelineRequest{
		Pipeline:           pipelineInfo.Pipeline,
		Transform:          pipelineInfo.Transform,
		ParallelismSpec:    pipelineInfo.ParallelismSpec,
		HashtreeSpec:       pipelineInfo.HashtreeSpec,
		Egress:             pipelineInfo.Egress,
		OutputBranch:       pipelineInfo.OutputBranch,
		ScaleDownThreshold: pipelineInfo.ScaleDownThreshold,
		ResourceRequests:   pipelineInfo.ResourceRequests,
		ResourceLimits:     pipelineInfo.ResourceLimits,
		Input:              pipelineInfo.Input,
		Description:        pipelineInfo.Description,
		CacheSize:          pipelineInfo.CacheSize,
		EnableStats:        pipelineInfo.EnableStats,
		Batch:              pipelineInfo.Batch,
		MaxQueueSize:       pipelineInfo.MaxQueueSize,
		Service:            pipelineInfo.Service,
		ChunkSpec:          pipelineInfo.ChunkSpec,
		DatumTimeout:       pipelineInfo.DatumTimeout,
		JobTimeout:         pipelineInfo.JobTimeout,
		Salt:               pipelineInfo.Salt,
	}
}

// PipelineManifestReader helps with unmarshalling pipeline configs from JSON. It's used by
// create-pipeline and update-pipeline
type PipelineManifestReader struct {
	buf     bytes.Buffer
	decoder *json.Decoder
}

// NewPipelineManifestReader creates a new manifest reader from a path.
func NewPipelineManifestReader(path string) (result *PipelineManifestReader, retErr error) {
	result = &PipelineManifestReader{}
	var pipelineReader io.Reader
	if path == "-" {
		pipelineReader = io.TeeReader(os.Stdin, &result.buf)
		fmt.Print("Reading from stdin.\n")
	} else if url, err := url.Parse(path); err == nil && url.Scheme != "" {
		resp, err := http.Get(url.String())
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := resp.Body.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		rawBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		pipelineReader = io.TeeReader(strings.NewReader(string(rawBytes)), &result.buf)
	} else {
		rawBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		pipelineReader = io.TeeReader(strings.NewReader(string(rawBytes)), &result.buf)
	}
	result.decoder = json.NewDecoder(pipelineReader)
	return result, nil
}

// NextCreatePipelineRequest gets the next request from the manifest reader.
func (r *PipelineManifestReader) NextCreatePipelineRequest() (*ppsclient.CreatePipelineRequest, error) {
	var result ppsclient.CreatePipelineRequest
	if err := jsonpb.UnmarshalNext(r.decoder, &result); err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("malformed pipeline spec: %s", err)
	}
	return &result, nil
}

// DescribeSyntaxError describes a syntax error encountered parsing json.
func DescribeSyntaxError(originalErr error, parsedBuffer bytes.Buffer) error {

	sErr, ok := originalErr.(*json.SyntaxError)
	if !ok {
		return originalErr
	}

	buffer := make([]byte, sErr.Offset)
	parsedBuffer.Read(buffer)

	lineOffset := strings.LastIndex(string(buffer[:len(buffer)-1]), "\n")
	if lineOffset == -1 {
		lineOffset = 0
	}

	lines := strings.Split(string(buffer[:len(buffer)-1]), "\n")
	lineNumber := len(lines)

	descriptiveErrorString := fmt.Sprintf("Syntax Error on line %v:\n%v\n%v^\n%v\n",
		lineNumber,
		string(buffer[lineOffset:]),
		strings.Repeat(" ", int(sErr.Offset)-2-lineOffset),
		originalErr,
	)

	return errors.New(descriptiveErrorString)
}

// IsTerminal returns 'true' if 'state' indicates that the job is done (i.e.
// the state will not change later: SUCCESS, FAILURE, KILLED) and 'false'
// otherwise.
func IsTerminal(state pps.JobState) bool {
	switch state {
	case pps.JobState_JOB_SUCCESS, pps.JobState_JOB_FAILURE, pps.JobState_JOB_KILLED:
		return true
	case pps.JobState_JOB_STARTING, pps.JobState_JOB_RUNNING, pps.JobState_JOB_MERGING:
		return false
	default:
		panic(fmt.Sprintf("unrecognized job state: %s", state))
	}
}

// UpdateJobState performs the operations involved with a job state transition.
func UpdateJobState(pipelines col.ReadWriteCollection, jobs col.ReadWriteCollection, jobPtr *pps.EtcdJobInfo, state pps.JobState, reason string) error {
	// Update pipeline
	pipelinePtr := &pps.EtcdPipelineInfo{}
	if err := pipelines.Get(jobPtr.Pipeline.Name, pipelinePtr); err != nil {
		return err
	}
	if pipelinePtr.JobCounts == nil {
		pipelinePtr.JobCounts = make(map[int32]int32)
	}
	if pipelinePtr.JobCounts[int32(jobPtr.State)] != 0 {
		pipelinePtr.JobCounts[int32(jobPtr.State)]--
	}
	pipelinePtr.JobCounts[int32(state)]++
	pipelinePtr.LastJobState = state
	if err := pipelines.Put(jobPtr.Pipeline.Name, pipelinePtr); err != nil {
		return err
	}

	// Update job info
	var err error
	if state == pps.JobState_JOB_STARTING {
		jobPtr.Started, err = types.TimestampProto(time.Now())
	} else if IsTerminal(state) {
		jobPtr.Finished, err = types.TimestampProto(time.Now())
	}
	if err != nil {
		return err
	}
	jobPtr.State = state
	jobPtr.Reason = reason
	return jobs.Put(jobPtr.Job.ID, jobPtr)
}
