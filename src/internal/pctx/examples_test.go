package pctx

import (
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/meters"
)

func ExampleChild() {
	log.InitPachctlLogger()
	log.SetLevel(log.DebugLevel)
	ctx := Background("")
	meters.Inc(ctx, "counter", 1)
	meters.Set(ctx, "gauge", 42)
	meters.Sample(ctx, "sampler", "hi")

	ctx, c := WithCancel(ctx)
	ctx = Child(ctx, "aggregated", WithCounter("counter", 0, meters.WithFlushInterval(time.Second)))
	ctx = Child(ctx, "", WithGauge("gauge", 0, meters.WithFlushInterval(time.Second)))
	for i := 0; i < 1<<24; i++ {
		meters.Inc(ctx, "counter", 1)
		meters.Set(ctx, "gauge", i)
	}
	c()
	// Output:
}
