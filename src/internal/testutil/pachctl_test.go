//go:build k8s

package testutil_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestPachctl(t *testing.T) {
	ctx := context.Background()

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	dirPath := t.TempDir()
	configPath := filepath.Join(dirPath, "test-config.json")
	p, err := testutil.NewPachctl(ctx, c, configPath)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	cmd, err := p.Command(ctx, `pachctl version`)
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Run(); err != nil {
		t.Log("stdout:", cmd.Stdout())
		t.Log("stderr:", cmd.Stderr())
		t.Fatal(err)
	}

	cmd, err = p.CommandTemplate(ctx,
		`echo "{{.foo}}" | match '^bar$'`,
		map[string]string{
			"foo": "bar",
		})
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Run(); err != nil {
		t.Log("stdout:", cmd.Stdout())
		t.Log("stderr:", cmd.Stderr())
		t.Fatal(err)
	}
}
