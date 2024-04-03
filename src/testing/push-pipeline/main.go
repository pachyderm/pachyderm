package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/kindenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/protoutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/randutil"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/testing/protocmp"
)

var (
	image    = flag.String("image", "", "bazel path of image to push")
	pipeline = flag.String("pipeline", "", "bazel path of pipeline to push")
)

func main() {
	flag.Parse()
	done := log.InitBatchLogger("")
	defer done(nil)

	ctx, cancel := pctx.Interactive()
	defer cancel()

	// iloc, err := runfiles.Rlocation(*image)
	// if err != nil {
	// 	log.Exit(ctx, "cannot find image runfiles", zap.Stringp("image", image), zap.Error(err))
	// }

	// ploc, err := runfiles.Rlocation(*pipeline)
	// if err != nil {
	// 	log.Exit(ctx, "cannot find pipeline runfiles", zap.Stringp("pipeline", pipeline), zap.Error(err))
	// }

	specfile, err := os.ReadFile(*pipeline)
	if err != nil {
		log.Exit(ctx, "problem reading pipeline spec", zap.Stringp("pipeline", pipeline), zap.Error(err))
	}
	spec := new(pps.CreatePipelineRequest)
	if err := protojson.Unmarshal(specfile, spec); err != nil {
		log.Exit(ctx, "problem unmarshaling pipeline", zap.Error(err))
	}
	req := protoutil.Clone(spec)

	cluster, err := kindenv.New(ctx, "")
	if err != nil {
		log.Exit(ctx, "problem finding pachdev cluster", zap.Error(err))
	}
	ccfg, err := cluster.GetConfig(ctx, "default")
	if err != nil {
		log.Exit(ctx, "problem getting cluster config", zap.Error(err))
	}

	id := randutil.UniqueString("")
	if err := cluster.PushImage(ctx, "oci:"+*image, "test:"+id); err != nil {
		log.Exit(ctx, "problem pushing image", zap.Error(err))
	}
	req.Transform.Image = path.Join(ccfg.ImagePullPath, "test:"+id)
	req.Update = true
	req.Reprocess = true

	diff := cmp.Diff(spec, req, protocmp.Transform())
	log.Info(ctx, "about to push new pipeline")
	fmt.Fprintf(os.Stderr, "\n%s\n", diff)

	pccfg := pachctl.Config{Verbose: true}
	c, err := pccfg.NewOnUserMachine(ctx, false)
	if err != nil {
		log.Exit(ctx, "problem making pachctl client", zap.Error(err))
	}
	if _, err := c.PpsAPIClient.CreatePipeline(c.Ctx(), req); err != nil {
		log.Exit(ctx, "problem pushing pipeline", zap.Error(err))
	}
	log.Info(ctx, "done")
}
