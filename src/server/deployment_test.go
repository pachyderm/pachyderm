package server

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gogo/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/serde"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

var _ = fmt.Printf

// Change this to false to keep kubernetes namespaces around after the test for debugging purposes
const cleanup = false

// Rewrites kubernetes manifest services to auto-allocate external ports and
// reduce cpu resource requests for parallel testing.
type ManifestRewriter struct {
	serde.Encoder
}

func rewriterCallback(innerCb func(map[string]interface{}) error) func(map[string]interface{}) error {
	return func(data map[string]interface{}) error {
		var err error
		if innerCb != nil {
			err = innerCb(data)
		}
		rewriteNodePort(data)
		return err
	}
}

func rewriteNodePort(data map[string]interface{}) {
	if data["kind"] == "Service" {
		spec := data["spec"].(map[string]interface{})
		if spec["type"] == "NodePort" {
			ports := spec["ports"].([]interface{})
			for _, port := range ports {
				port := port.(map[string]interface{})
				if _, ok := port["nodePort"]; ok {
					port["nodePort"] = 0
				}
			}
		}
	}

	if data["kind"] == "Deployment" {
		if spec, ok := data["spec"]; ok {
			spec := spec.(map[string]interface{})
			if template, ok := spec["template"]; ok {
				template := template.(map[string]interface{})
				if spec, ok := template["spec"]; ok {
					spec := spec.(map[string]interface{})
					if containers, ok := spec["containers"]; ok {
						containers := containers.([]interface{})
						for _, container := range containers {
							container := container.(map[string]interface{})
							if resources, ok := container["resources"]; ok {
								resources := resources.(map[string]interface{})
								if limits, ok := resources["limits"]; ok {
									limits := limits.(map[string]interface{})
									if _, ok := limits["cpu"]; ok {
										limits["cpu"] = "0"
									}
								}
								if requests, ok := resources["requests"]; ok {
									requests := requests.(map[string]interface{})
									if _, ok := requests["cpu"]; ok {
										requests["cpu"] = "0"
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func (npr *ManifestRewriter) Encode(v interface{}) error {
	return npr.EncodeTransform(v, nil)
}

func (npr *ManifestRewriter) EncodeProto(m proto.Message) error {
	return npr.EncodeProtoTransform(m, nil)
}

func (npr *ManifestRewriter) EncodeTransform(v interface{}, cb func(map[string]interface{}) error) error {
	return npr.Encoder.EncodeTransform(v, rewriterCallback(cb))
}

func (npr *ManifestRewriter) EncodeProtoTransform(m proto.Message, cb func(map[string]interface{}) error) error {
	return npr.Encoder.EncodeProtoTransform(m, rewriterCallback(cb))
}

func getPachClient(t *testing.T, kubeClient *kube.Clientset, namespace string) *client.APIClient {
	// Get the pachd service from kubernetes
	pachd, err := kubeClient.CoreV1().Services(namespace).Get("pachd", metav1.GetOptions{})
	require.NoError(t, err)

	var port int32
	for _, servicePort := range pachd.Spec.Ports {
		if servicePort.Name == "api-grpc-port" {
			port = servicePort.NodePort
		}
	}
	require.NotEqual(t, 0, port)

	// Get the IP address of the nodes (any _should_ work for the service port)
	nodes, err := kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
	require.NoError(t, err)

	// Minikube 'Hostname' address type didn't work when testing, use InternalIP
	var address string
	for _, addr := range nodes.Items[0].Status.Addresses {
		if addr.Type == "InternalIP" {
			address = addr.Address
		}
	}
	require.NotEqual(t, "", address)

	// Wait until pachd pod is ready
	waitForReadiness(t, namespace)

	// Connect to pachd
	fmt.Printf("Connecting to %s:%d\n", address, port)
	client, err := client.NewFromAddress(fmt.Sprintf("%s:%d", address, port))
	require.NoError(t, err)
	fmt.Printf("Connected\n")
	return client
}

func makeManifest(t *testing.T, backend assets.Backend, secrets map[string][]byte, opts *assets.AssetOpts) string {
	manifest := &strings.Builder{}
	jsonEncoder, err := serde.GetEncoder("json", manifest, serde.WithIndent(2), serde.WithOrigName(true))
	require.NoError(t, err)

	// Create a wrapper encoder that rewrites NodePort ports to be randomly
	// assigned so that we don't get collisions across namespaces and can run
	// these tests in parallel.
	encoder := &ManifestRewriter{Encoder: jsonEncoder}

	// Use a separate hostpath on the kubernetes host for each deployment
	hostPath := fmt.Sprintf("/var/pachyderm-%s", opts.Namespace)
	err = assets.WriteAssets(encoder, opts, backend, assets.LocalBackend, 1, hostPath)
	require.NoError(t, err)

	err = assets.WriteSecret(encoder, secrets, opts)
	require.NoError(t, err)

	return manifest.String()
}

func withManifest(t *testing.T, backend assets.Backend, secrets map[string][]byte, callback func(namespace string, pachClient *client.APIClient)) {
	namespaceName := tu.UniqueString("deployment-test-")
	opts := &assets.AssetOpts{
		StorageOpts: assets.StorageOpts{
			UploadConcurrencyLimit:  assets.DefaultUploadConcurrencyLimit,
			PutFileConcurrencyLimit: assets.DefaultPutFileConcurrencyLimit,
		},
		PachdShards:                16,
		Version:                    "latest",
		LogLevel:                   "info",
		Namespace:                  namespaceName,
		RequireCriticalServersOnly: assets.DefaultRequireCriticalServersOnly,
		WorkerServiceAccountName:   assets.DefaultWorkerServiceAccountName,
		NoDash:                     true,
	}

	manifest := makeManifest(t, backend, secrets, opts)

	kubeClient := tu.GetKubeClient(t)
	namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespaceName}}
	_, err := kubeClient.CoreV1().Namespaces().Create(namespace)
	require.NoError(t, err)

	if cleanup {
		defer func() {
			err := kubeClient.CoreV1().Namespaces().Delete(namespaceName, nil)
			require.NoError(t, err)
		}()
	}

	cmd := tu.Cmd("kubectl", "apply", "--namespace", namespaceName, "-f", "-")
	cmd.Stdin = strings.NewReader(manifest)
	err = cmd.Run()
	require.NoError(t, err)

	pachClient := getPachClient(t, kubeClient, namespaceName)
	defer pachClient.Close()

	callback(namespaceName, pachClient)
}

func runBasicTest(t *testing.T, pachClient *client.APIClient) {
	// Create an input repo
	dataRepo := "data"
	require.NoError(t, pachClient.CreateRepo(dataRepo))

	// Upload some files
	commit1, err := pachClient.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = pachClient.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, pachClient.FinishCommit(dataRepo, commit1.ID))

	// Create a pipeline
	pipelineRepo := tu.UniqueString("pipeline")
	_, err = pachClient.PpsAPIClient.CreatePipeline(context.Background(), &pps.CreatePipelineRequest{
		Pipeline: client.NewPipeline(pipelineRepo),
		Transform: &pps.Transform{
			Image: "",
			Cmd:   []string{"bash"},
			Stdin: []string{
				fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
			},
		},
		ParallelismSpec: &pps.ParallelismSpec{
			Constant: 1,
		},
		Input:                 client.NewPFSInput(dataRepo, "/*"),
		OutputBranch:          "",
		Update:                false,
		ResourceRequests:      &pps.ResourceSpec{Cpu: 0.0},
		ResourceLimits:        &pps.ResourceSpec{Cpu: 0.0},
		SidecarResourceLimits: &pps.ResourceSpec{Cpu: 0.0},
	})
	require.NoError(t, err)

	// Wait for the output commit
	commitIter, err := pachClient.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)
	commitInfos := collectCommitInfos(t, commitIter)
	require.Equal(t, 1, len(commitInfos))

	// Check the pipeline output
	var buf bytes.Buffer
	require.NoError(t, pachClient.GetFile(commitInfos[0].Commit.Repo.Name, commitInfos[0].Commit.ID, "file", 0, 0, &buf))
	require.Equal(t, "foo", buf.String())
}

func newDefaultAmazonConfig() *obj.AmazonAdvancedConfiguration {
	return &obj.AmazonAdvancedConfiguration{
		Retries:        obj.DefaultRetries,
		Timeout:        obj.DefaultTimeout,
		UploadACL:      obj.DefaultUploadACL,
		Reverse:        obj.DefaultReverse,
		PartSize:       obj.DefaultPartSize,
		MaxUploadParts: obj.DefaultMaxUploadParts,
		DisableSSL:     obj.DefaultDisableSSL,
		NoVerifySSL:    obj.DefaultNoVerifySSL,
		LogOptions:     obj.DefaultAwsLogOptions,
	}
}

func loadAmazonParameters(t *testing.T) (string, string, string, string) {
	id := os.Getenv("AMAZON_DEPLOYMENT_ID")
	secret := os.Getenv("AMAZON_DEPLOYMENT_SECRET")
	bucket := os.Getenv("AMAZON_DEPLOYMENT_BUCKET")
	region := os.Getenv("AMAZON_DEPLOYMENT_REGION")
	require.NotEqual(t, "", id)
	require.NotEqual(t, "", secret)
	require.NotEqual(t, "", bucket)
	require.NotEqual(t, "", region)

	return id, secret, bucket, region
}

func loadECSParameters(t *testing.T) (string, string, string, string, string) {
	id := os.Getenv("ECS_DEPLOYMENT_ID")
	secret := os.Getenv("ECS_DEPLOYMENT_SECRET")
	bucket := os.Getenv("ECS_DEPLOYMENT_BUCKET")
	region := os.Getenv("ECS_DEPLOYMENT_REGION") // The region is unused but needed by the object client
	endpoint := os.Getenv("ECS_DEPLOYMENT_CUSTOM_ENDPOINT")
	require.NotEqual(t, "", id)
	require.NotEqual(t, "", secret)
	require.NotEqual(t, "", bucket)
	require.NotEqual(t, "", region)
	require.NotEqual(t, "", endpoint)

	return id, secret, bucket, region, endpoint
}

func loadGoogleParameters(t *testing.T) (string, string) {
	bucket := os.Getenv("GOOGLE_DEPLOYMENT_BUCKET")
	creds := os.Getenv("GOOGLE_DEPLOYMENT_CREDS")
	require.NotEqual(t, "", bucket)
	require.NotEqual(t, "", creds)

	return bucket, creds
}

func loadMicrosoftParameters(t *testing.T) (string, string, string) {
	id := os.Getenv("MICROSOFT_DEPLOYMENT_ID")
	secret := os.Getenv("MICROSOFT_DEPLOYMENT_SECRET")
	container := os.Getenv("MICROSOFT_DEPLOYMENT_CONTAINER")
	require.NotEqual(t, "", id)
	require.NotEqual(t, "", secret)
	require.NotEqual(t, "", container)

	return id, secret, container
}

func TestAmazonDeployment(t *testing.T) {
	advancedConfig := newDefaultAmazonConfig()

	// Test the Amazon client against S3
	t.Run("AmazonObjectStorage", func(t *testing.T) {
		// t.Parallel()
		id, secret, bucket, region := loadAmazonParameters(t)
		secrets := assets.AmazonSecret(region, bucket, id, secret, "", "", "", advancedConfig)
		withManifest(t, assets.AmazonBackend, secrets, func(namespace string, pachClient *client.APIClient) {
			runBasicTest(t, pachClient)
		})
	})

	// Test the Amazon client against ECS
	t.Run("ECSObjectStorage", func(t *testing.T) {
		// t.Parallel()
		id, secret, bucket, region, endpoint := loadECSParameters(t)
		secrets := assets.AmazonSecret(region, bucket, id, secret, "", "", endpoint, advancedConfig)
		withManifest(t, assets.AmazonBackend, secrets, func(namespace string, pachClient *client.APIClient) {
			runBasicTest(t, pachClient)
		})
	})
}

func TestMinioDeployment(t *testing.T) {
	S3V2s := []bool{false, true}

	// Test the Minio client against S3 using the S3v2 and S3v4 APIs
	t.Run("AmazonObjectStorage", func(t *testing.T) {
		id, secret, bucket, region := loadAmazonParameters(t)
		endpoint := fmt.Sprintf("s3.%s.amazonaws.com", region) // Note that not all AWS regions support both http/https or both S3v2/S3v4
		for _, isS3V2 := range S3V2s {
			t.Run(fmt.Sprintf("isS3V2=%v", isS3V2), func(t *testing.T) {
				// t.Parallel()
				secrets := assets.MinioSecret(bucket, id, secret, endpoint, true, isS3V2) // TODO: need endpoint?
				withManifest(t, assets.MinioBackend, secrets, func(namespace string, pachClient *client.APIClient) {
					runBasicTest(t, pachClient)
				})
			})
		}
	})

	// Test the Minio client against ECS using the S3v2 and S3v4 APIs
	t.Run("ECSObjectStorage", func(t *testing.T) {
		id, secret, bucket, _, endpoint := loadECSParameters(t)
		for _, isS3V2 := range S3V2s {
			t.Run(fmt.Sprintf("isS3V2=%v", isS3V2), func(t *testing.T) {
				// t.Parallel()
				secrets := assets.MinioSecret(bucket, id, secret, endpoint, true, isS3V2)
				withManifest(t, assets.MinioBackend, secrets, func(namespace string, pachClient *client.APIClient) {
					runBasicTest(t, pachClient)
				})
			})
		}
	})

	/*
	// Test the Minio client against GCP using the S3v2 and S3v4 APIs
	t.Run("ECSObjectStorage", func(t *testing.T) {
		bucket, creds := loadGoogleParameters(t)
		for _, isS3V2 := range S3V2s {
			t.Run(fmt.Sprintf("isS3V2=%v", isS3V2), func(t *testing.T) {
				// t.Parallel()
				secrets := assets.MinioSecret(bucket, id, secret, endpoint, true, isS3V2)
				withManifest(t, assets.MinioBackend, secrets, func(namespace string, pachClient *client.APIClient) {
					runBasicTest(t, pachClient)
				})
			})
		}
	})
	*/
}

func TestGoogleDeployment(t *testing.T) {
	// t.Parallel()
	bucket, creds := loadGoogleParameters(t)
	secrets := assets.GoogleSecret(bucket, creds)
	withManifest(t, assets.GoogleBackend, secrets, func(namespace string, pachClient *client.APIClient) {
		runBasicTest(t, pachClient)
	})
}

func TestMicrosoftDeployment(t *testing.T) {
	// t.Parallel()
	id, secret, container := loadMicrosoftParameters(t)
	secrets := assets.MicrosoftSecret(container, id, secret)
	withManifest(t, assets.MicrosoftBackend, secrets, func(namespace string, pachClient *client.APIClient) {
		runBasicTest(t, pachClient)
	})
}
