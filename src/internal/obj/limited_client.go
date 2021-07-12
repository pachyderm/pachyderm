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
	blockStartedMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "pachyderm",
		Subsystem: "pfs_object_storage",
		Name:      "limited_block_start_total",
		Help:      "The number of times a blocking operation has started (even if it wouldn't block), by operation name.  The number of finished blocks is available via pachyderm_pfs_object_storage_limited_seconds_count.",
	}, []string{"op"})
	blockedSecondsMetric = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "pachyderm",
		Subsystem: "pfs_object_storage",
		Name:      "limited_seconds",
		Help:      "Distribution of time spent waiting behind the limitedClient semaphore, by operation name",
		Buckets:   []float64{0.00001, 0.00005, 0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 10, 30, 60},
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
	blockStartedMetric.WithLabelValues("put").Inc()
	t := time.Now()
	if err := loc.writersSem.Acquire(ctx, limitClientSemCost); err != nil {
		return err
	}
	blockedSecondsMetric.WithLabelValues("put").Observe(time.Since(t).Seconds())
	defer loc.writersSem.Release(limitClientSemCost)
	return loc.Client.Put(ctx, name, r)
}

func (loc *limitedClient) Get(ctx context.Context, name string, w io.Writer) error {
	blockStartedMetric.WithLabelValues("get").Inc()
	t := time.Now()
	if err := loc.readersSem.Acquire(ctx, limitClientSemCost); err != nil {
		return err
	}
	blockedSecondsMetric.WithLabelValues("get").Observe(time.Since(t).Seconds())
	defer loc.readersSem.Release(limitClientSemCost)
	return loc.Client.Get(ctx, name, w)
}
