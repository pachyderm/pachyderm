package jobs

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/zeebo/xxh3"
	"go.uber.org/zap"
)

type GoBinary struct {
	Workdir  string   // Where to invoke Go.
	CGO      bool     // If true, enable cgo.
	Target   string   // The "go build ..." target.
	Platform Platform // The platform to build for.
}

func (g GoBinary) String() string {
	return fmt.Sprintf("<go_binary cd %v; %v GOOS=%v GOARCH=%v go build %s>", g.Workdir, g.cgo(), g.Platform.GOOS(), g.Platform.GOARCH(), g.Target)
}

func (g GoBinary) ID() uint64 {
	return xxh3.HashString(g.Workdir + g.Target + string(g.Platform))
}

func (g GoBinary) Inputs() []Reference {
	return nil
}

func (g GoBinary) output() NameAndPlatform {
	return NameAndPlatform{
		Name:     "go_binary:" + g.Workdir + ":" + g.Target,
		Platform: g.Platform,
	}
}

func (g GoBinary) cgo() string {
	if g.CGO {
		return "CGO_ENABLED=1"
	}
	return "CGO_ENABLED=0"
}

func (g GoBinary) Outputs() []Reference {
	return []Reference{g.output()}
}

func (g GoBinary) Run(ctx context.Context, jc *JobContext, inputs []Artifact) (_ []Artifact, retErr error) {
	fh, err := os.CreateTemp("", "go_binary-*")
	if err != nil {
		return nil, WrapRetryable(errors.Wrap(err, "create output file"))
	}
	defer errors.Close(&retErr, fh, "close output")

	cmd := exec.CommandContext(ctx, "go", "build", "-v", "-o", fh.Name(), g.Target)
	cmd.Dir = g.Workdir
	if arm, ok := g.Platform.GOARM(); ok {
		cmd.Env = append(cmd.Env, "GOARM="+arm)
	}
	cmd.Env = append(cmd.Env, "GOARCH="+g.Platform.GOARCH())
	cmd.Env = append(cmd.Env, "GOOS="+g.Platform.GOOS())
	cmd.Env = append(cmd.Env, "GOPATH=/tmp/.gopath")
	cmd.Env = append(cmd.Env, "GOCACHE=/tmp/.go-cache")
	cmd.Env = append(cmd.Env, g.cgo())

	execCtx, done := log.SpanContext(ctx, "go build", zap.Stringer("platform", g.Platform))
	stdout, stderr := log.WriterAt(pctx.Child(execCtx, "stdout"), log.DebugLevel), log.WriterAt(pctx.Child(execCtx, "stderr"), log.InfoLevel)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	defer errors.Close(&retErr, stdout, "close stdout")
	defer errors.Close(&retErr, stderr, "close stderr")

	err = cmd.Run()
	done(zap.Error(err))
	if err != nil {
		return nil, errors.Wrap(err, "go build")
	}
	return []Artifact{
		&GoBinaryFile{
			NameAndPlatform: g.output(),
			File: &File{
				Path: fh.Name(),
			},
		},
	}, nil
}

type GoBinaryFile struct {
	NameAndPlatform
	*File
}

func (g GoBinaryFile) FS() fs.FS {
	return g.File.FS()
}

func GenerateGoBinaryJobs(g GoBinary) []Job {
	var result []Job
	for _, p := range KnownPlatforms {
		g.Platform = p
		result = append(result, g)
	}
	return result
}
