package kindenv

// NOTE(jrockway): This should live in //tools/skopeo and the go_library should depend on the
// binary... but we want to use internal/log, and //tools/skopeo can't see //src/internal.

import (
	"context"
	"os/exec"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func SkopeoCommand(ctx context.Context, args ...string) *exec.Cmd {
	skopeo, ok := bazel.FindBinary("//tools/skopeo", "_skopeo")
	if !ok {
		log.Error(ctx, "binary not built with bazel; falling back to host skopeo")
		skopeo = "skopeo"
	}
	defaultArgs := []string{"--insecure-policy"} // Avoid reading /etc/containers/policy.json, which does not exist in CI.
	defaultArgs = append(defaultArgs, args...)
	cmd := exec.CommandContext(ctx, skopeo, defaultArgs...)
	cmd.Args[0] = "skopeo"
	cmd.Stdout = log.WriterAt(pctx.Child(ctx, "skopeo.stdout"), log.InfoLevel)
	cmd.Stderr = log.WriterAt(pctx.Child(ctx, "skopeo.stderr"), log.InfoLevel)
	return cmd
}
