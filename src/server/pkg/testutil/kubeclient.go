package testutil

import (
	"bytes"
	"fmt"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"os"
	"os/exec"
	"strings"
	"testing"

	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GetKubeClient connects to the Kubernetes API server either from inside the
// cluster or from a test binary running on a machine with kubectl (it will
// connect to the same cluster as kubectl)
func GetKubeClient(t testing.TB) *kube.Clientset {
	var config *rest.Config
	wd, err := os.Getwd()
	fmt.Printf("cur wd: %v (err: %v)\n", wd, err)
	fmt.Printf("KUBECONFIG: %v\n", os.Getenv("KUBECONFIG"))
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	if host != "" {
		var err error
		config, err = rest.InClusterConfig()
		require.NoError(t, err)
	} else {
		// Use kubectl binary to parse .kube/config and get address of current
		// cluster. Hopefully, once we upgrade to k8s.io/client-go, we will be able
		// to do this in-process with a library
		// First, figure out if we're talking to minikube or localhost
		cmd := exec.Command("kubectl", "config", "current-context")
		cmd.Stderr = os.Stderr
		if context, err := cmd.Output(); err == nil {
			context = bytes.TrimSpace(context)
			// kubectl has a context -- not talking to localhost
			// Get cluster and user name from kubectl
			buf := &bytes.Buffer{}
			cmd := BashCmd(strings.Join([]string{
				`kubectl config get-contexts "{{.context}}" | tail -n+2 | awk '{print $3}'`,
				`kubectl config get-contexts "{{.context}}" | tail -n+2 | awk '{print $4}'`,
			}, "\n"),
				"context", string(context))
			cmd.Stdout = buf
			require.NoError(t, cmd.Run(), "couldn't get kubernetes context info")
			lines := strings.Split(buf.String(), "\n")
			clustername, username := lines[0], lines[1]
			fmt.Printf("clustername: %v\nusername: %v\n", clustername, username)

			// Get user info
			buf.Reset()
			cmd = BashCmd(strings.Join([]string{
				`cluster="$(kubectl config view -o json | jq -r '.users[] | select(.name == "{{.user}}") | .user' )"`,
				`echo "${cluster}" | jq -r '.["client-certificate"]'`,
				`echo "${cluster}" | jq -r '.["client-key"]'`,
			}, "\n"),
				"user", username)
			cmd.Stdout = buf
			require.NoError(t, cmd.Run(), "couldn't get kubernetes user info")
			lines = strings.Split(buf.String(), "\n")
			clientCert, clientKey := lines[0], lines[1]

			// Get cluster info
			buf.Reset()
			cmd = BashCmd(strings.Join([]string{
				`cluster="$(kubectl config view -o json | jq -r '.clusters[] | select(.name == "{{.cluster}}") | .cluster')"`,
				`echo "${cluster}" | jq -r .server`,
				`echo "${cluster}" | jq -r '.["certificate-authority"]'`,
			}, "\n"),
				"cluster", clustername)
			cmd.Stdout = buf
			require.NoError(t, cmd.Run(), "couldn't get kubernetes cluster info: %s", buf.String())
			lines = strings.Split(buf.String(), "\n")
			address, CAKey := lines[0], lines[1]
			fmt.Printf("address: %v\n", address)

			// Generate config
			config = &rest.Config{
				Host: address,
				TLSClientConfig: rest.TLSClientConfig{
					CertFile: clientCert,
					KeyFile:  clientKey,
					CAFile:   CAKey,
				},
			}
		} else {
			fmt.Printf("kubectl config current-context yielded: %v\n", err)
			// no context -- talking to localhost
			config = &rest.Config{
				Host: "http://0.0.0.0:8080",
				TLSClientConfig: rest.TLSClientConfig{
					Insecure: false,
				},
			}
		}
	}
	k, err := kube.NewForConfig(config)
	require.NoError(t, err)
	return k
}
