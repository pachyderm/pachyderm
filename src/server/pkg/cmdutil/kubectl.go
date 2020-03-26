package cmdutil


import (
	"io"
	"os"
	"fmt"
	"path"
	"io/ioutil"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
)

const configTemplate = `
apiVersion: v1
contexts:
- context:
    cluster: "%s"
    user: "%s"
    namespace: "%s"
  name: pachyderm-active-context
current-context: pachyderm-active-context
kind: Config
`

func RunKubectl(stdin io.Reader, context *config.Context, args ...string) error {
	tmpfile, err := ioutil.TempFile("", "transient-kube-config-*.yaml")
	if err != nil {
		return errors.Wrapf(err, "failed to create transient kube config")
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Write([]byte(fmt.Sprintf(configTemplate, context.ClusterName, context.AuthInfo, context.Namespace)))
	tmpfile.Close()

	var environ []string = nil
	if context != nil {
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				errors.Wrapf(err, "failed to determine user home directory")
			}
			kubeconfig = path.Join(home, ".kube", "config")
		}
		kubeconfig = fmt.Sprintf("%s:%s", kubeconfig, tmpfile.Name())

		// note that this will override `KUBECONFIG` (if it is already defined) in
		// the environment; see examples under
		// https://golang.org/pkg/os/exec/#Command
		environ = os.Environ()
		environ = append(environ, fmt.Sprintf("KUBECONFIG=%s", kubeconfig))

		if stdin == nil {
			stdin = os.Stdin
		}
	}

	ioObj := IO {
		Stdin: stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Environ: environ,
	}

	args = append([]string{"kubectl"}, args...)
	return RunIO(ioObj, args...)
}
