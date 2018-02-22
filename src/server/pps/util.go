package pps

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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"

	"github.com/gogo/protobuf/jsonpb"
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

// PipelineReqFromInfo converts a PipelineInfo into a CreatePipelineRequest.
func PipelineReqFromInfo(pipelineInfo *ppsclient.PipelineInfo) *ppsclient.CreatePipelineRequest {
	return &ppsclient.CreatePipelineRequest{
		Pipeline:           pipelineInfo.Pipeline,
		Transform:          pipelineInfo.Transform,
		ParallelismSpec:    pipelineInfo.ParallelismSpec,
		Egress:             pipelineInfo.Egress,
		OutputBranch:       pipelineInfo.OutputBranch,
		ScaleDownThreshold: pipelineInfo.ScaleDownThreshold,
		ResourceRequests:   pipelineInfo.ResourceRequests,
		ResourceLimits:     pipelineInfo.ResourceLimits,
		Input:              pipelineInfo.Input,
		Description:        pipelineInfo.Description,
		Incremental:        pipelineInfo.Incremental,
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
