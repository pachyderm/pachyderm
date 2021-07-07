package obj

import (
	"context"
	io "io"
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/sync/semaphore"
)

var (
	blockedSecondsMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "pachyderm",
		Subsystem: "object_storage",
		Name:      "limited_seconds_total",
		Help:      "Total",
	}, []string{"op"})
)

const limitClientSemCost = 1

var _ Client = &limitedClient{}

// limitedClient is a Client which limits the number of objects open at a time
// for reading and writing respectively.
type limitedClient struct {
	Client
	writersSem *semaphore.Weighted
	readersSem *semaphore.Weighted
}

// NewLimitedClient constructs a Client which will only ever have
//   <= maxReaders objects open for reading
//   <= maxWriters objects open for writing
// if either is < 1 then that constraint is ignored.
func NewLimitedClient(client Client, maxReaders, maxWriters int) Client {
	if maxReaders < 1 {
		maxReaders = int(math.MaxInt64)
	}
	if maxWriters < 1 {
		maxWriters = int(math.MaxInt64)
	}
	return &limitedClient{
		Client:     client,
		writersSem: semaphore.NewWeighted(int64(maxWriters)),
		readersSem: semaphore.NewWeighted(int64(maxReaders)),
	}
}

func (loc *limitedClient) Put(ctx context.Context, name string, r io.Reader) error {
	t := time.Now()
	if err := loc.writersSem.Acquire(ctx, limitClientSemCost); err != nil {
		return err
	}
	blockedSecondsMetric.WithLabelValues("put").Add(time.Since(t).Seconds())
	defer loc.writersSem.Release(limitClientSemCost)
	return loc.Client.Put(ctx, name, r)
}

func (loc *limitedClient) Get(ctx context.Context, name string, w io.Writer) error {
	t := time.Now()
	if err := loc.readersSem.Acquire(ctx, limitClientSemCost); err != nil {
		return err
	}
	blockedSecondsMetric.WithLabelValues("get").Add(time.Since(t).Seconds())
	defer loc.readersSem.Release(limitClientSemCost)
	return loc.Client.Get(ctx, name, w)
}
