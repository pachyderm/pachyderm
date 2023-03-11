package pctx

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/m"
)

func ExampleChild() {
	log.InitPachctlLogger()
	log.SetLevel(log.DebugLevel)
	ctx := Background("")
	m.Inc(ctx, "counter", 1)
	m.Set(ctx, "gauge", 42)
	m.Sample(ctx, "sampler", "hi")

	ctx, c := context.WithCancel(ctx)
	ctx = Child(ctx, "aggregated", WithCounter("counter", 0, m.WithFlushInterval(time.Second)))
	ctx = Child(ctx, "", WithGauge("gauge", 0, m.WithFlushInterval(time.Second)))
	for i := 0; i < 1<<24; i++ {
		m.Inc(ctx, "counter", 1)
		m.Set(ctx, "gauge", i)
	}
	c()
	// Output:
}
