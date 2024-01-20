package kindenv

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func TestCluster(t *testing.T) {
	os.Setenv("XDG_STATE_HOME", t.TempDir())
	xdg.Reload()

	ctx := pctx.TestContext(t)
	var r uint16
	if err := binary.Read(rand.Reader, binary.BigEndian, &r); err != nil {
		t.Fatalf("read random cluster id: %v", err)
	}
	name := fmt.Sprintf("kindenv-test-%v", r)
	c, err := New(ctx, name)
	if err != nil {
		t.Fatalf("new cluster: %v", err)
	}
	t.Cleanup(func() {
		t.Log("cleanup tmp kubeconfig")
		if err := c.Close(); err != nil {
			t.Errorf("close cluster: %v", err)
		}
	})
	registry := fmt.Sprintf("test-registry-%v", r)
	t.Cleanup(func() {
		t.Log("delete registry")
		if err := destroyRegistry(ctx, registry); err != nil {
			t.Errorf("destroy registry: %v", err)
		}
	})
	opts := &CreateOpts{
		TestNamespaceCount: 0,
		BindHTTPPorts:      false,
		StartingPort:       -1,
		kubeconfigPath:     filepath.Join(t.TempDir(), "kubeconfig"),
		registryName:       registry,
	}
	t.Cleanup(func() {
		t.Log("delete cluster")
		// It is safe to call Delete even if the cluster doesn't exist.
		if err := c.Delete(ctx); err != nil {
			t.Errorf("delete cluster: %v", err)
		}
	})
	if err := c.Create(ctx, opts); err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	pause, err := runfiles.Rlocation("_main/src/internal/kindenv/pause")
	if err != nil {
		t.Skip("can't find pause container; probably not running the test with bazel")
	}
	if err := c.PushImage(ctx, "oci:"+pause, "pause:latest"); err != nil {
		t.Fatalf("push pause image: %v", err)
	}
	k, err := c.GetKubeconfig(ctx)
	if err != nil {
		t.Fatalf("get kubeconfig: %v", err)
	}
	if err := k.KubectlCommand(ctx, "create", "deployment", "pause", fmt.Sprintf("--image=%s:5001/pause:latest", registry)).Run(); err != nil {
		t.Fatalf("create pause deployment: %v", err)
	}
	if err := k.KubectlCommand(ctx, "rollout", "status", "deployment", "pause", "--timeout=20s").Run(); err != nil {
		if err := k.KubectlCommand(ctx, "describe", "pod").Run(); err != nil {
			t.Logf("additional error running `kubectl describe pod`: %v", err)
		}
		t.Fatalf("wait for pause deployment: %v", err)
	}
}
