package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"strings"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pachyderm/pachyderm/v2/src/internal/imagebuilder/jobs"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
)

func mustDecodeHex(s string) []byte {
	hash, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return hash
}

func main() {
	log.InitPachctlLogger()
	log.SetLevel(log.DebugLevel)
	ctx, c := signal.NotifyContext(pctx.Background(""), os.Interrupt)
	defer c()

	todo := []jobs.Job{
		&jobs.Download{
			Name:     "dumb-init",
			Platform: "linux/amd64",
			URL:      "https://github.com/Yelp/dumb-init/releases/download/v1.2.5/dumb-init_1.2.5_x86_64",
			WantDigest: jobs.Digest{
				Algorithm: "blake3",
				Value:     mustDecodeHex("9a520c3860a67bca23323e2dfa9e263f8dd54000b1c890b44db2a5316c607284"),
			},
		},
		&jobs.Download{
			Name: "dumb-init",
			URL:  "https://github.com/Yelp/dumb-init/releases/download/v1.2.5/dumb-init_1.2.5_aarch64",
			WantDigest: jobs.Digest{
				Algorithm: "blake3",
				Value:     mustDecodeHex("8df4e75473552405410b17e555da4b4c493857627bfdb3c0fd9be45c79827182"),
			},
			Platform: "linux/arm64",
		},
	}
	goBuilds := []jobs.GoBinary{
		{
			Workdir: "/home/jrockway/pach/pachyderm",
			Target:  "./src/server/cmd/pachd",
		},
		{
			Workdir: "/home/jrockway/pach/pachyderm",
			Target:  "./src/server/cmd/worker",
		},
		{
			Workdir: "/home/jrockway/pach/pachyderm",
			Target:  "./src/server/cmd/pachctl",
		},
	}
	for _, g := range goBuilds {
		binaries := jobs.GenerateGoBinaryJobs(g)
		todo = append(todo, binaries...)
		for _, bin := range binaries {
			for _, l := range jobs.PlatformLayers(bin.Outputs()) {
				todo = append(todo, l)
				todo = append(todo, jobs.Manifest{
					Name:     jobs.Name(l.Input.Name),
					Config:   jobs.Name("file:docker-config"),
					Platform: l.Input.Platform,
					Layers:   l.Outputs(),
				})
			}
		}
	}
	todo = append(todo,
		&jobs.TestJob{
			Name: "combine",
			Ins:  append(todo[0].Outputs(), todo[1].Outputs()...),
			Outs: []jobs.Reference{jobs.Name("output")},
			F: func(ctx context.Context, jc *jobs.JobContext, inputs []jobs.Artifact) ([]jobs.Artifact, error) {
				b := new(strings.Builder)
				for i, in := range inputs {
					fmt.Fprintf(b, "%#v", in)
					if i < len(inputs)-1 {
						b.WriteString(" + ")
					}
				}
				return []jobs.Artifact{struct {
					jobs.Name
					Data string
				}{Name: jobs.Name("output"), Data: b.String()}}, nil
			},
		},
	)
	config, err := jobs.JSONFile("file:docker-config", "config.json", &v1.ImageConfig{})
	if err != nil {
		log.Exit(ctx, "create in-memory docker config")
	}

	want := []jobs.Reference{
		jobs.Name("output"),
		jobs.NameAndPlatform{
			Name:     "manifest:go_binary:/home/jrockway/pach/pachyderm:./src/server/cmd/pachd",
			Platform: "linux/amd64",
		},
	}
	plan, err := jobs.Plan(ctx, todo, want, jobs.RunnerOption{Artifacts: []jobs.Artifact{config}})
	if err != nil {
		log.Exit(ctx, "plan", zap.Error(err))
	}
	fmt.Printf("Plan:\n%s\n", plan)

	// fmt.Println("Now running...")
	// outputs, err := jobs.Resolve(ctx, todo, want)
	// if err != nil {
	// 	log.Exit(ctx, "resolve", zap.Error(err))
	// }
	// for i, output := range outputs {
	// 	fmt.Printf("%#v -> %#v\n", want[i], output)
	// }
}
