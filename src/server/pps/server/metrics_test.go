//go:build k8s

package server_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
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
pachctl create project video-to-frame-traces
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
kubectl -n {{.namespace}} apply -f - <<EOF
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: worker-scraper
  labels:
    release: {{.namespace}}-prometheus
spec:
  selector:
    matchLabels:
      app: pipeline
      component: worker
  namespaceSelector:
    matchNames:
      - {{.namespace}}
  podMetricsEndpoints:
  - port: metrics-storage
    path: /metrics
  - port: metrics-user
    path: /metrics
EOF
`,
		"namespace", namespace,
		"docsContent", docsDir,
		"tmpDir", tmpDir)
	cmd.Stdout = os.Stdout

	require.NoError(t, cmd.Run())
	require.NoErrorWithinTRetry(t, 2*time.Minute, func() error {
		var resp *http.Response
		var err error
		addr := c.GetAddress()
		uri := fmt.Sprintf("http://%s:%d/api/v1/query?query=pachyderm_auth_dex_approval_errors_total", addr.Host, addr.Port+10)
		if resp, err = http.Get(uri); err != nil {
			return errors.Wrapf(err, "could not fetch %s", uri)
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "could not read response")
		}
		defer resp.Body.Close()
		type response struct {
			Status string `json:"status"`
			Data   struct {
				ResultType string `json:"resultType"`
				Result     []any  `json:"result"`
			} `json:"data"`
		}
		var pResp response
		if err := json.Unmarshal(b, &pResp); err != nil {
			return errors.Wrap(err, "could not parse response")
		}
		if pResp.Status != "success" {
			return errors.Errorf("got %q; expected \"success\"", pResp.Status)
		}
		if len(pResp.Data.Result) == 0 {
			return errors.New("no result")
		}
		return nil
	})
}
