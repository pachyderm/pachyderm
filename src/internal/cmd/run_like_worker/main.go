package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
	"go.uber.org/zap"
)

func main() {
	log.InitWorkerLogger()
	log.SetLevel(log.DebugLevel)
	rctx, rcancel := pctx.Interactive()

	p := &pps.PipelineInfo{
		Pipeline: &pps.Pipeline{
			Name: "test",
			Project: &pfs.Project{
				Name: "default",
			},
		},
		Type: pps.PipelineInfo_PIPELINE_TYPE_TRANSFORM,
		Details: &pps.PipelineInfo_Details{
			Transform: &pps.Transform{
				Cmd: os.Args[1:],
			},
		},
	}

	lineCh := make(chan string)
	go func() {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			lineCh <- s.Text()
		}
		close(lineCh)
		rcancel()
	}()
	for i := 0; ; i++ {
		ctx, cancel := pctx.WithCancel(pctx.Child(rctx, fmt.Sprintf("run %d", i)))
		ch := make(chan struct{})
		d := driver.NewEmptyDriver(ctx, p)
		go func() {
			select {
			case <-lineCh:
				log.Info(ctx, "killing user code")
				cancel()
			case <-ctx.Done():
			}
		}()
		go func() {
			log.Info(ctx, "starting user code")
			l := logs.New(pctx.Child(ctx, "RunUserCode"))
			err := d.RunUserCode(ctx, l, os.Environ())
			log.Info(ctx, "user code done", zap.Error(err))
			close(ch)
		}()
		<-ch
		select {
		case <-rctx.Done():
			os.Exit(0)
		default:
		}
	}
}
