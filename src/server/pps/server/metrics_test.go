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
# install the podmonitor we document for users
git clone https://github.com/pachyderm/docs-content.git --depth=1 '{{.docsContent}}'
## update the spec for this namespace & Helm release
sed -e 's/- default/- {{.namespace}}/' '{{.docsContent}}/latest/manage/prometheus/podmonitor.yaml' -e 's/release: default/release: {{.namespace}}-prometheus/' | kubectl -n {{.namespace}} apply -f -
kubectl wait -n {{.namespace}} --for=condition=ready pod -l app.kubernetes.io/name=prometheus --timeout=5m
`,
		"namespace", namespace,
		"docsContent", docsDir,
		"tmpDir", tmpDir)
	cmd.Stdout = os.Stdout

	require.NoError(t, cmd.Run())
	var resp *http.Response
	require.NoErrorWithinTRetry(t, time.Minute, func() error {
		var err error
		addr := c.GetAddress()
		uri := fmt.Sprintf("http://%s:%d/api/v1/query?query=pachyderm_auth_dex_approval_errors_total", addr.Host, addr.Port+10)
		resp, err = http.Get(uri)
		return errors.Wrap(err, "could not fetch %s", uri)
	})

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
}
