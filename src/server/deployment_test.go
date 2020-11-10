package server

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/serde"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"
)

var _ = fmt.Printf

func getPachClient(t *testing.T, kubeClient *kube.Clientset, namespace string) *client.APIClient {
	// Get the pachd service from kubernetes
	pachd, err := kubeClient.CoreV1().Services(namespace).Get("pachd", metav1.GetOptions{})
	require.NoError(t, err)

	var port int32
	for _, servicePort := range pachd.Spec.Ports {
		if servicePort.Name == "api-grpc-port" {
			port = servicePort.Port
		}
	}
	require.NotEqual(t, 0, port)
	address := fmt.Sprintf("%s:%d", pachd.Spec.ClusterIP, port)

	// Wait until pachd pod is ready
	waitForReadiness(t, namespace)

	// Connect to pachd
	fmt.Printf("connecting to pachd: %s\n", address)
	client, err := client.NewFromAddress(address)
	require.NoError(t, err)
	return client
}

func makeManifest(t *testing.T, backend assets.Backend, secrets map[string][]byte, opts *assets.AssetOpts) string {
	manifest := &strings.Builder{}
	encoder, err := serde.GetEncoder("json", manifest,
		serde.WithIndent(2),
		serde.WithOrigName(true),
	)
	require.NoError(t, err)

	err = assets.WriteAssets(encoder, opts, backend, assets.LocalBackend, 1, "/var/pachyderm")
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
		/*
			PachdCPURequest:            "",
			PachdNonCacheMemRequest:    "",
			BlockCacheSize:             "",
			EtcdCPURequest:             "",
			EtcdMemRequest:             "",
			EtcdNodes:                  0,
			EtcdVolume:                 "",
			EtcdStorageClassName:       "",
			DashOnly:                   false,
			DashImage:                  "",
			Registry:                   "",
			ImagePullSecret:            "",
			NoGuaranteed:               false,
			NoRBAC:                     false,
			LocalRoles:                 false,
			NoExposeDockerSocket:       false,
			ExposeObjectAPI:            false,
			Metrics:                    false,
			ClusterDeploymentID:        "",
		*/
	}

	manifest := makeManifest(t, backend, secrets, opts)

	fmt.Printf("manifest:\n%s\n", manifest)

	kubeClient := tu.GetKubeClient(t)
	namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespaceName}}
	_, err := kubeClient.CoreV1().Namespaces().Create(namespace)
	require.NoError(t, err)
	/*
		defer func() {
			err := kubeClient.CoreV1().Namespaces().Delete(namespaceName, nil)
			require.NoError(t, err)
		}()
	*/

	cmd := tu.Cmd("kubectl", "apply", "--namespace", namespaceName, "-f", "-")
	cmd.Stdin = strings.NewReader(manifest)
	err = cmd.Run()
	require.NoError(t, err)

	pachClient := getPachClient(t, kubeClient, namespaceName)
	defer pachClient.Close()

	callback(namespaceName, pachClient)
}

/*
func makeManifest(namespace string, args []string) (string, error) {
	cmdArgs := []string{"deploy", args[0], "--dry-run", "--namespace", namespace}
	cmdArgs = append(cmdArgs, args[1:]...)
	fmt.Printf("calling deploy with args: %v\n", cmdArgs)
	cmd := tu.Cmd("pachctl", cmdArgs...)
	manifest := &strings.Builder{}
	cmd.Stdout = manifest
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return manifest.String(), nil
}
*/

func TestAmazonDeployment(t *testing.T) {
	id := os.Getenv("AMAZON_DEPLOYMENT_ID")
	secret := os.Getenv("AMAZON_DEPLOYMENT_SECRET")
	bucket := os.Getenv("AMAZON_DEPLOYMENT_BUCKET")
	region := os.Getenv("AMAZON_DEPLOYMENT_REGION")
	require.NotEqual(t, "", id)
	require.NotEqual(t, "", secret)
	require.NotEqual(t, "", bucket)
	require.NotEqual(t, "", region)

	advancedConfig := &obj.AmazonAdvancedConfiguration{
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

	secrets := assets.AmazonSecret(region, bucket, id, secret, "", "", "", advancedConfig)
	withManifest(t, assets.AmazonBackend, secrets, func(namespace string, pachClient *client.APIClient) {
	})
}

func runBasicTest(t *testing.T) {
}

func TestGoogleDeployment(t *testing.T) {
}

func TestMicrosoftDeployment(t *testing.T) {
}

func TestECSDeployment(t *testing.T) {
}

func TestMinioDeployment(t *testing.T) {
}
