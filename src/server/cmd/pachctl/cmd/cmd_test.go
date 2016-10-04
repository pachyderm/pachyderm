package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	deploycmds "github.com/pachyderm/pachyderm/src/server/pkg/deploy/cmds"
)

func TestMetrics(t *testing.T) {

	// Run deploy normally, should see METRICS=true
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Args = []string{"deploy", "--dry-run"}
	err := deploycmds.DeployCmd().Execute()
	require.NoError(t, err)
	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// restore stdout
	w.Close()
	os.Stdout = old
	out := <-outC

	decoder := json.NewDecoder(strings.NewReader(out))
	for {
		var jsonObj json.RawMessage
		err = decoder.Decode(&jsonObj)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		var manifests []interface{}
		for _, obj := range manifests {
			manifest, ok := obj.(map[string]interface{})
			require.Equal(t, true, ok)
			metadata, ok := manifest["metadata"].(map[string]interface{})
			require.Equal(t, true, ok)
			name, ok := metadata["name"].(string)
			require.Equal(t, true, ok)

			if name == "pachd" {
				spec, ok := manifest["spec"].(map[string]interface{})
				require.Equal(t, true, ok)
				template, ok := spec["template"].(map[string]interface{})
				require.Equal(t, true, ok)
				innerSpec, ok := template["spec"].(map[string]interface{})
				require.Equal(t, true, ok)
				containers, ok := innerSpec["containers"].([]interface{})
				require.Equal(t, true, ok)
				container, ok := containers[0].(map[string]interface{})
				require.Equal(t, true, ok)
				env, ok := container["env"].([]interface{})
				require.Equal(t, true, ok)
				metricsSetToTrue := map[string]string{
					"name":  "METRICS",
					"value": "true",
				}
				fmt.Printf("env: %v\n", env)
				require.OneOfEquals(t, metricsSetToTrue, env)
			}
		}
	}

	// Run deploy w dev flag, should see METRICS=false
}
