package dbutil

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/jackc/pgconn"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

var (
	txStartedMetric = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pachyderm",
		Subsystem: "postgres",
		Name:      "tx_start_count",
		Help:      "Count of transactions that have been started.  One transaction may start many underlying database transactions, whose status is tracked separately.",
	})
	txFinishedMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "pachyderm",
		Subsystem: "postgres",
		Name:      "tx_finish_count",
		Help:      "Count of transactions that have finished, by outcome ('error', 'ok', etc.).",
	}, []string{"outcome"})
	txDurationMetric = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "pachyderm",
		Subsystem: "postgres",
		Name:      "tx_duration_seconds",
		Help:      "Time taken for a transaction, by outcome ('error', 'ok', etc.).",
		Buckets: []float64{
			0.0001, // 100us
			0.0005, // .5ms
			0.001,  // 1ms
			0.002,  // 2ms
			0.005,  // 5ms
			0.01,   // 10ms
			0.02,   // 20ms
			0.05,   // 50ms
			0.1,    // 100ms
			0.2,    // 200ms
			0.5,    // 500ms
			1,      // 1s
			2,      // 2s
			5,      // 5s
			30,     // 30s
			60,     // 60s
			300,    // 5m
			600,    // 10m
			3600,   // 1h
			86400,  // 1d
		},
	}, []string{"outcome"})
	triesPerTxMetric = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "pachyderm",
		Subsystem: "postgres",
		Name:      "tx_attempt_count",
		Help:      "Count of underlying database transactions to resolve each application-level transaction, by outcome ('error', 'ok').  One that works on the first try reports '1' (attempt_count, not retry_count).  Failures are included.",
		Buckets:   []float64{0, 1, 2, 3, 4, 5, 10, 50, 100, 1000},
	}, []string{"outcome"})
	underlyingTxStartedMetric = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pachyderm",
		Subsystem: "postgres",
		Name:      "tx_underlying_start_count",
		Help:      "Count of underlying database transactions that have been started.",
	})
	underlyingTxFinishMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "pachyderm",
		Subsystem: "postgres",
		Name:      "tx_underlying_finish_count",
		Help:      "Count of underlying database transactions that have finished, by outcome ('commit_ok', 'commit_failed', 'rollback_ok', 'rollback_failed', 'failed_start', etc.)",
	}, []string{"outcome"})
)

type withTxConfig struct {
	sql.TxOptions
	backoff.BackOff
}

// WithTxOption parameterizes the WithTx function
type WithTxOption func(c *withTxConfig)

// WithIsolationLevel runs the transaction with the specified isolation level.
func WithIsolationLevel(x sql.IsolationLevel) WithTxOption {
	return func(c *withTxConfig) {
		c.TxOptions.Isolation = x
	}
}

// TODO: Unused, but probably acceptable to leave in place
// WithReadOnly causes WithTx to run the transaction as read only
func WithReadOnly() WithTxOption {
	return func(c *withTxConfig) {
		c.TxOptions.ReadOnly = true
	}
}

// TODO: Unused, but probably acceptable to leave in place
// WithBackOff sets the BackOff used when retrying
func WithBackOff(bo backoff.BackOff) WithTxOption {
	return func(c *withTxConfig) {
		c.BackOff = bo
	}
}

// WithTx calls cb with a transaction,
// The transaction is committed IFF cb returns nil.
// If cb returns an error the transaction is rolled back.
func WithTx(ctx context.Context, db *pachsql.DB, cb func(tx *pachsql.Tx) error, opts ...WithTxOption) error {
	backoffStrategy := backoff.NewExponentialBackOff()
	backoffStrategy.InitialInterval = 1 * time.Millisecond
	backoffStrategy.MaxElapsedTime = 0
	backoffStrategy.Multiplier = 1.05
	c := &withTxConfig{
		TxOptions: sql.TxOptions{
			Isolation: sql.LevelSerializable,
		},
		BackOff: backoffStrategy,
	}
	for _, opt := range opts {
		opt(c)
	}

	var attempts int
	var outcome string
	start := time.Now()
	defer func() {
		txFinishedMetric.WithLabelValues(outcome).Inc()
		txDurationMetric.WithLabelValues(outcome).Observe(time.Since(start).Seconds())
		triesPerTxMetric.WithLabelValues(outcome).Observe(float64(attempts))
	}()

	txStartedMetric.Inc()
	err := backoff.RetryUntilCancel(ctx, func() error {
		ctx, cf := context.WithCancel(context.Background())
		defer cf()
		underlyingTxStartedMetric.Inc()
		attempts++
		tx, err := db.BeginTxx(ctx, &c.TxOptions)
		if err != nil {
			underlyingTxFinishMetric.WithLabelValues("failed_start").Inc()
			return errors.EnsureStack(err)
		}
		return tryTxFunc(tx, cb)
	}, c.BackOff, func(err error, _ time.Duration) error {
		if isTransactionError(err) {
			return nil
		}
		return err
	})
	if err != nil {
		// Inspecting err could yield a better outcome type than "error", but some care is
		// needed.  For example, `cb` could return "context deadline exceeded" because it
		// created a sub-context that expires, and that's a different error than 'commit'
		// failing because the deadline expired during commit.  But we can't know that here
		// without extra annotations on the error.
		outcome = "error"
		if errors.Is(err, sql.ErrTxDone) {
			outcome += "_txdone"
		}
		return err
	}
	outcome = "ok"
	return nil
}

func tryTxFunc(tx *pachsql.Tx, cb func(tx *pachsql.Tx) error) error {
	if err := cb(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			underlyingTxFinishMetric.WithLabelValues("rollback_failed").Inc()
			logrus.Error(rbErr)
			return err // The user error, not the rollback error.
		}
		underlyingTxFinishMetric.WithLabelValues("rollback_ok").Inc()
		return err
	}
	if err := tx.Commit(); err != nil {
		label := "commit_failed"
		if errors.Is(err, sql.ErrTxDone) {
			label += "_txdone"
		}
		underlyingTxFinishMetric.WithLabelValues(label).Inc()
		return errors.EnsureStack(err)
	}
	underlyingTxFinishMetric.WithLabelValues("commit_ok").Inc()
	return nil
}

func isTransactionError(err error) bool {
	pgxErr := &pgconn.PgError{}
	return errors.As(err, &pgxErr) && strings.HasPrefix(pgxErr.Code, "40")
}
