package main

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type imageTags struct {
	pachd, worker string
}

func readRunfile(rlocation string) ([]byte, error) {
	path, err := runfiles.Rlocation(rlocation)
	if err != nil {
		return nil, errors.Wrapf(err, "get runfile %v", rlocation)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "read runfile %v", rlocation)
	}
	return content, nil
}

func getVersions() (imageTags, error) {
	var errs error
	pachd, err := readRunfile("_main/oci/pachd_image.json.sha256")
	if err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "get pachd image digest"))
	}
	worker, err := readRunfile("_main/oci/worker_image.json.sha256")
	if err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "get worker image digest"))
	}
	return imageTags{pachd: string(pachd), worker: string(worker)}, errs
}

func runSkopeo(ctx context.Context, args ...string) (retErr error) {
	ctx, done := log.SpanContext(ctx, "skopeo")
	defer done(log.Errorp(&retErr))
	cmd := exec.CommandContext(ctx, "skopeo", args...)
	cmd.Stdout = log.WriterAt(pctx.Child(ctx, "stdout"), log.DebugLevel)
	cmd.Stderr = log.WriterAt(pctx.Child(ctx, "stderr"), log.ErrorLevel)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "run skopeo")
	}
	return nil
}

func waitForPachd(ctx context.Context) error {
	cl, err := client.NewOnUserMachine(ctx, "")
	if err != nil {
		return errors.Wrap(err, "NewOnUserMachine")
	}
	tryOnce := func() (*versionpb.Version, error) {
		ctx, c := context.WithTimeout(ctx, time.Second)
		cl := cl.WithCtx(ctx)
		defer c()
		v, err := cl.VersionAPIClient.GetVersion(cl.Ctx(), &emptypb.Empty{})
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	for {
		v, err := tryOnce()
		if err != nil {
			if strings.Contains(err.Error(), context.Canceled.Error()) {
				return context.Cause(ctx)
			}
			log.Error(ctx, "problem getting version; retrying...", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}
		if got, want := v.Canonical(), version.Version.Canonical(); got != want {
			log.Debug(ctx, "version not matched yet; retrying...", zap.String("got", got), zap.String("want", want))
			time.Sleep(time.Second)
			continue
		}
		log.Info(ctx, "server is running desired version", zap.String("version", v.Canonical()))
		break
	}
	return nil
}

func main() {
	log.InitPachctlLogger()
	log.SetLevel(log.DebugLevel)
	ctx, c := signal.NotifyContext(pctx.Background("push"), os.Interrupt)
	defer c()

	localRegistry := "oci:/tmp/zot/"
	remoteRegistry := "localhost:5001"

	// Find images.
	log.Info(ctx, "locating images")
	pachdImage, err := runfiles.Rlocation("_main/oci/pachd_image")
	if err != nil {
		log.Exit(ctx, "problem getting pachd image from runfiles", zap.Error(err))
	}
	workerImage, err := runfiles.Rlocation("_main/oci/worker_image")
	if err != nil {
		log.Exit(ctx, "problem getting worker image from runfiles", zap.Error(err))
	}

	// Find versions (because the pachyderm helm chart doesn't support referencing a build with @sha256:...).
	versions, err := getVersions()
	if err != nil {
		log.Exit(ctx, "problem getting container digests from runfiles", zap.Error(err))
	}
	versions.pachd = strings.TrimSpace(versions.pachd[7:])
	versions.worker = strings.TrimSpace(versions.worker[7:])

	// Find local helm values.
	localValues, err := runfiles.Rlocation("_main/src/cmd/push/local.values.json")
	if err != nil {
		log.Exit(ctx, "problem getting helm values from runfiles", zap.Error(err))
	}
	valuesFiles, err := filepath.Glob(filepath.Join(filepath.Dir(localValues), "*.values.json"))
	if err != nil {
		log.Exit(ctx, "glob *.values.json in runfiles directory", zap.String("rlocation", localValues), zap.Error(err))
	}

	// Copy images to registry.
	log.Info(ctx, "pushing images")
	if err := runSkopeo(ctx, "copy", "oci:"+pachdImage, localRegistry+"pachd:"+versions.pachd); err != nil {
		log.Exit(ctx, "problem copying pachd image to registry", zap.Error(err))
	}
	if err := runSkopeo(ctx, "copy", "oci:"+workerImage, localRegistry+"worker"+versions.worker); err != nil {
		log.Exit(ctx, "problem copying worker image to registry", zap.Error(err))
	}

	// Run "helm upgrade".
	helmArgs := []string{
		"upgrade",
		"pachyderm",
		filepath.Join(os.Getenv("BUILD_WORKSPACE_DIRECTORY"), "etc/helm/pachyderm"),
		"--set",
		strings.Join([]string{
			"pachd.image.repository=" + remoteRegistry + "/pachd",
			"pachd.image.tag=" + versions.pachd,
			"worker.image.repository=" + remoteRegistry + "/worker",
			"worker.image.tag=" + versions.worker,
		}, ","),
	}
	for _, f := range valuesFiles {
		helmArgs = append(helmArgs, "-f", f)
	}

	log.Info(ctx, "running helm upgrade", zap.Strings("args", helmArgs))
	cmd := exec.CommandContext(ctx, "helm", helmArgs...)
	cmd.Stdout = log.WriterAt(pctx.Child(ctx, "helm.stdout"), log.DebugLevel)
	cmd.Stderr = log.WriterAt(pctx.Child(ctx, "helm.stderr"), log.ErrorLevel)
	if err := cmd.Run(); err != nil {
		log.Exit(ctx, "problem running helm upgrade", zap.Error(err))
	}

	log.Info(ctx, "waiting for pachd to be serving traffic")
	if err := waitForPachd(ctx); err != nil {
		log.Exit(ctx, "problem waiting for pachd", zap.Error(err))
	}
	log.Info(ctx, "done")
}
