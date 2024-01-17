//go:build k8s

package server_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestContainerMetrics(t *testing.T) {
	ctx, cancel := pctx.WithCancel(pctx.TestContext(t))
	defer cancel()
	t.Parallel()
	c, namespace := minikubetestenv.AcquireCluster(t, minikubetestenv.WithPrometheus)

	docsDir := t.TempDir()
	tmpDir := t.TempDir()
	cmd := testutil.PachctlBashCmdCtx(ctx, t, c, `
# run a pipeline in order to have a container to scrape metrics from
pachctl create project "video-to-frame-traces"
pachctl create repo raw_videos_and_images
pachctl put file raw_videos_and_images@master:robot.png -f https://raw.githubusercontent.com/pachyderm/docs-content/main/images/opencv/robot.jpg
pachctl create pipeline -f - <<EOF
pipeline:
  name: video_mp4_converter
input:
  pfs:
    repo: raw_videos_and_images
    glob: "/*"
transform:
  image: lbliii/video_mp4_converter:1.0.14
  cmd:
    - python3
    - /video_mp4_converter.py
    - --input
    - /pfs/raw_videos_and_images/
    - --output
    - /pfs/out/
EOF
# install prometheus
#helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
# install the podmonitor we document for users
git clone https://github.com/pachyderm/docs-content.git --depth=1 '{{.docsContent}}'
## update the spec for this namespace & Helm release
sed -e 's/- default/- {{.namespace}}/' '{{.docsContent}}/latest/manage/prometheus/podmonitor.yaml' -e 's/release: default/release: {{.namespace}}-prometheus/' | kubectl -n {{.namespace}} apply -f -
kubectl wait -n {{.namespace}} --for=condition=ready pod -l app.kubernetes.io/name=prometheus --timeout=5m
## listen on a random local port; use sed to extract port and write to file; use tee -p to prevent pipe-write error
kubectl port-forward -n {{.namespace}} service/{{.namespace}}-prometheus-prometheus :9090 | tee -p | head -1 | sed -e 's/Forwarding from 127.0.0.1://' -e 's/ -> 9090//' > {{.tmpDir}}/port
`,
		"namespace", namespace,
		"docsContent", docsDir,
		"tmpDir", tmpDir)
	cmd.Stdout = os.Stdout
	// This is a little tricky.  It would be appealing to run kubectl
	// port-forward in the background and complete the shell script, but
	// cmd.Run will not return until the subprocess die.  Daemonizing it
	// would mean having to record the PID and kill it manually.  Instead,
	// we’ll just run the shell script asynchronously, and poll until
	// we can read the randomly-allocated local port.
	require.NoError(t, cmd.Start())
	portCh := make(chan int)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				break
			default:
			}
			f, err := os.Open(filepath.Join(tmpDir, "port"))
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			b, err := bufio.NewReader(f).ReadString('\n')
			if err == io.EOF {
				// Could potentially get an EOF error if the
				// shell redirection is not done writing a full
				// line; in that case just retry.
				continue
			}
			require.NoError(t, err)
			i, err := strconv.Atoi(strings.TrimSpace(string(b)))
			require.NoError(t, err)
			portCh <- i
			break
		}
		close(portCh)
	}(ctx)
	port := <-portCh // wait until the async process has written the local port
	require.NotEqual(t, 0, portCh)
	// if we can see a pachyderm log, then metrics are being scraped from the container
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/v1/query?query=pachyderm_auth_dex_approval_errors_total", port))
	require.NoError(t, err)
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	type response struct {
		Status string `json:"status"`
		Data   any    `json:"data"`
	}
	var pResp response
	err = json.Unmarshal(b, &pResp)
	require.NoError(t, err)
	require.Equal(t, "success", pResp.Status)
	require.NoError(t, cmd.Process.Kill())
}
