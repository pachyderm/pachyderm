package server

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/pachyderm/pachyderm/src/server/pkg/serde"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"
)

func makeManifest(opts *assets.AssetOpts, backend assets.Backend, secrets map[string][]byte) (string, error) {
	manifest := &strings.Builder{}
	encoder, err := serde.GetEncoder("json", manifest,
		serde.WithIndent(2),
		serde.WithOrigName(true),
	)
	if err != nil {
		return "", err
	}

	if err := assets.WriteAssets(encoder, opts, backend, assets.LocalBackend, "1", ""); err != nil {
		return "", err
	}

	if err := assets.WriteSecret(encoder, secrets, opts); err != nil {
		return "", err
	}

	return manifest.String(), nil
}

func withManifest(kubeClient *kube.Clientset, deployArgs []string, callback func(*v1.Namespace) error) (retErr error) {
	namespaceName := tu.UniqueString("deployment-test-")

	manifest, err := makeManifest(namespaceName, deployArgs)
	if err != nil {
		return err
	}

	namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespaceName}}
	if _, err := kubeClient.CoreV1().Namespaces().Create(namespace); err != nil {
		return err
	}
	/*
		defer func() {
			if err := kubeClient.CoreV1().Namespaces().Delete(namespaceName, nil); err != nil && retErr == nil {
				retErr = err
			}
		}()
	*/

	cmd := tu.Cmd("kubectl", "apply", "--namespace", namespaceName, "-f", "-")
	cmd.Stdin = strings.NewReader(manifest)
	if err := cmd.Run(); err != nil {
		return err
	}

	return callback(namespace)
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
	amazonID := os.Getenv("AMAZON_DEPLOYMENT_ID")
	amazonSecret := os.Getenv("AMAZON_DEPLOYMENT_SECRET")
	amazonBucket := os.Getenv("AMAZON_DEPLOYMENT_BUCKET")
	amazonRegion := os.Getenv("AMAZON_DEPLOYMENT_REGION")
	require.NotEqual(t, "", amazonID)
	require.NotEqual(t, "", amazonSecret)
	require.NotEqual(t, "", amazonBucket)
	require.NotEqual(t, "", amazonRegion)

	creds := fmt.Sprintf("%s,%s", amazonID, amazonSecret)
	deployArgs := []string{
		"amazon",
		amazonBucket,
		amazonRegion,
		"1", // GB of disk space
		"--credentials", creds,
		"--dynamic-etcd-nodes=1",
	}

	kubeClient := tu.GetKubeClient(t)
	err := withManifest(kubeClient, deployArgs, func(namespace *v1.Namespace) error {
		return nil
	})
	require.NoError(t, err)
}

func getPachClient(namespace *v1.Namespace) (*client.APIClient, error) {
	return nil, nil
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
