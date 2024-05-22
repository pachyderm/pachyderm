// Package meters implements lightweight metrics for internal use.
//
// A meter is a name, a set of key/value pairs (set by the underlying logger), and a value.  A value
// can be an absolute value ("gauge"), an incremental value ("counter"), or a set of samples (like
// "histogram", though lossless in this implementation).  Meters are most useful when additional
// code analyzes the entire log of a program's run; this is called "analysis code" or "anaylsis
// software" throughout the documentation.
//
// Values are stored by writing the set of operations to build the value to the logs.  For example,
// each time a sample is added to a sampler with Sample, a log line is produced.  Reading the logs
// will allow analysis software to recover the entire set of samples.  Counters are similar; each
// increment event emits a log line, and and analysis code can add the deltas to see the final
// value, or the value at a particular point in time.
//
// Values are logically associated with key=value metadata.  Each value is uniquely identified by
// the set of key=value metadata; a meter foo with field foo=bar is logically a different value from
// the meter foo with field foo=baz.  The metadata is defined by the fields applied to the
// underlying logger.  For example, if you declare a counter `tx_bytes` on each HTTP request you
// handle, each time the count of transmitted bytes is updated, the log message will include fields
// inherited from the default HTTP logger, which includes a per-request ID.  Therefore, analysis
// code can calculate a tx_byte count for every request handled by the system by observing the value
// of the x-request-id field on log lines matching the tx_byte counter format.  And, of course, it
// can ignore the extra fields and add everything up to show a program-wide count of bytes
// transmitted.
//
// Because pctx.Child allows the caller to define a namespace for logs and metrics, meter names do
// not need to be namespaced.  Prefer a meter named `tx_bytes` in the `chunk_storage.Upload` logger
// over one named `chunk_storage.upload.tx_bytes`.
//
// Normally it's considered "too expensive" to store metrics with such a high cardinality (per-user,
// per-IP, per-request), but this system has no such limitation.  Cardinality has no impact on write
// performance.  Analysis code can make the decision on which fields to discard to decrease the
// cardinality for long-term storage, if desired.  Be aware that the usual log sampling rules apply
// to logged meters.
//
// Aggregating each operation on a meter over time recovers the value of the meter at a particular
// time.  Any sort of smartness or validation comes from the reader, not from this writing code.  If
// you want to treat a certain meter name as a string gauge, integer counter, and sampler of proto
// messages, that is OK.  The analysis code that processes the logs will need to be ready for that,
// or at least ready to ignore values it doesn't think are valid.
//
// To emit meters, simply call these public functions in this package:
//
//	Set(ctx, "helpful_name", value) // Set the current value of the meter to value.
//	Inc(ctx, "helpful_name", delta) // Increment the current value of the meter by delta.
//	Sample(ctx, "helpful_name", sample) // Add sample to the set of samples captured by the meter.
//
// It is safe to write to the same meter from different goroutines concurrently.
//
// Some minimal aggregation can be done in-process.  This is a compromise to reduce the "noise" in
// the logs.  If you were uploading a 1GB file, it would make sense to increment the byte count
// meter 1 billion times, as each byte of the file is passed to the TCP stack.  This, however, would
// be very noisy and make the logs difficult to read.  So, we offer aggregates to flush the value of
// gauges (Set) and counters (Inc) periodically, grouping many value-changing operations into a
// single log message.
//
// An aggregator can be registered on a context (with pctx.Child), causing all future writes to that
// meter on that context to be aggregated.  (The code that writes the meter need not be aware of the
// aggregated nature; the public Set/Inc/Sample API automatically does the right thing.)
//
// Aggregated meters are emitted to the logs based on a time interval or value delta threshold set
// at registration time.  If a write occurs, and the meter hasn't been logged for that interval,
// then a log line will be produced showing the current value of the meter.  If unflushed data
// exists when the controlling context is Done, a log line representing the final value is also
// emitted.  From an analysis standpoint, nothing changes; the emitted aggregated meters are
// indistinguishable from immediately emitted meters.  It just is a little bit easier on the eyes of
// the human reader.
//
// Aggregated meters can only be one type of meter (gauge/counter) with one type of value; if you
// create a counter named tx_bytes and then call the gauge operatoin "Set" on it, no aggregation
// will be done on that data.  (Inc calls continue to be aggregated; it doesn't "break" the
// aggregation to accidentally call the wrong value-changing function.)  Similarly, the Go type of
// the value is set at registration time; calls writing values of a different type will not be
// aggregated, but again, do not break any ongoing aggregation.
//
// Finally, a couple of implementation notes. This package never panics or returns an error; every
// possible argument you can pass to any public function is valid.  Also, functions must never
// create or contend on any global locks (except the lock on the logger to prevent goroutines from
// writing partial lines).  This is so that heavy use of meters in one component does not delay
// other more crticial code.  Locks scoped to one context chain are OK, but a global lock on
// something popular like `tx_bytes` is not desirable.
package meters

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"
	"golang.org/x/exp/constraints"
)

// aggregatedMetersKey is the context key identifying aggregated meters.  The value of this key is
// a gauge[T] or counter[T].
type aggregatedMetersKey struct{ meter string }

func meterFromContext[T any](ctx context.Context, meter string) *T {
	if x, ok := ctx.Value(aggregatedMetersKey{meter: meter}).(*T); ok {
		return x
	}
	return nil
}

// allParentsMetersKey is the context key that tracks all aggregated meters in the context chain.
// The value of this key is a parentMeters slice.  When changing the set of active fields with
// WithNewFields, each parent meter is re-created for the resulting child context.  (There isn't a
// way to walk the chain of contexts, so we have to keep an index ourselves for the copy that .)
type allParentMeters struct{}

// parentMeters is a slice of functions that create an aggregated meter and return a child context
// with the aggregated meter in effect.  This is so that when changing log fields, we can copy
// every aggregated meter in the context chain, and start tracking the values for the new set of
// fields.  (The pctx package knows when this is necessary.)
type parentMeters []func(context.Context) context.Context

// String implements fmt.Stringer.  (For debugging context chains; reflect.ValueOf(ctx).String()
// will show you everything inside the context chain.)
func (p parentMeters) String() string {
	return fmt.Sprintf("{%v parent meters}", len(p))
}

// valueType returns a type hint for parsers that need to know the value type to synthesize a meter
// in some other stricter monitoring system.  The parser will have to choose what it wants to do
// when different set/increment/sample operations emit different types; we don't care, but they
// might.
func valueType(val any) string {
	switch val.(type) {
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
		return "int"
	case float32, float64:
		return "float"
	case string, []byte:
		// rune is an alias for int32, byte is an alias for uint8, so those types can't be
		// (single-character) strings.  Note that []byte gets marshaled as zap.Binary, so it
		// will be base64-encoded.
		return "string"
	default:
		return fmt.Sprintf("%T", val)
	}
}

func logMeter(ctx context.Context, skip int, meter string, fields ...zap.Field) {
	fields = append(fields, zap.String("meter", meter))
	log.Debug(log.ChildLogger(ctx, "", log.WithOptions(zap.AddCallerSkip(skip))), "meter: "+meter, fields...)
}

// Set sets the value of a signed numeric meter.  (This exists because Delta gauges have more
// constraints than gauges in general.)
func Set[T Signed](ctx context.Context, meter string, val T) {
	if g := meterFromContext[gauge[T]](ctx, meter); g != nil {
		g.set(ctx, val)
		return
	}
	if g := meterFromContext[delta[T]](ctx, meter); g != nil {
		g.set(ctx, val)
		return
	}
	logGauge(ctx, 0, meter, val)
}

// SetGauge sets the value of an arbitrary-type gauge meter.
func SetGauge[T any](ctx context.Context, meter string, val T) {
	if g := meterFromContext[gauge[T]](ctx, meter); g != nil {
		g.set(ctx, val)
		return
	}
	logGauge(ctx, 0, meter, val)
}

func logGauge[T any](ctx context.Context, skip int, meter string, val T) {
	logMeter(ctx, 3+skip, meter, zap.String("type", valueType(val)), zap.Any("value", val))
}

// Monoid is a constraint matching types that have an "empty" value and "append" operation.
// Consider an integer; 0 is "empty", and "append" is addition.  Any Monoid can be a Counter meter
// value.
type Monoid interface {
	constraints.Integer | constraints.Float | constraints.Complex
}

// Signed is a constraint over signed numbers.
type Signed interface {
	constraints.Signed | constraints.Float
}

// Inc changes the value of a meter by the provided delta.
func Inc[T Monoid](ctx context.Context, meter string, delta T) {
	if c := meterFromContext[counter[T]](ctx, meter); c != nil {
		c.inc(ctx, delta)
		return
	}
	// In immediate mode, a delta of 0 is a no-op; the absence of a log line also indicates that
	// the counter was incremented by 0.  In aggregate mode above, that's not quite true; it is
	// potentially an opportunity to flush the aggregated value to the logs.
	if delta == 0 {
		return
	}
	logCounter(ctx, 0, meter, delta)
}

func logCounter[T Monoid](ctx context.Context, skip int, meter string, delta T) {
	logMeter(ctx, 3+skip, meter, zap.String("type", valueType(delta)), zap.Any("delta", delta))
}

// Sample adds a sample to the value of the meter.
func Sample[T any](ctx context.Context, meter string, val T) {
	logMeter(ctx, 2, meter, zap.String("type", valueType(val)), zap.Any("sample", val))
}

// aggregateOption contains optional parameters for customizing an aggregating meter.
type aggregateOptions struct {
	flushInterval time.Duration // How long to wait, at a minimum, before reporting the value of the meter.
	doneCh        chan struct{} // only for testing
}

// Option supplies optional configuration to aggregated meters.
type Option func(o *aggregateOptions)

// WithFlushInterval is an Option that sets the amount of time to aggregate a meter for before emitting.
func WithFlushInterval(interval time.Duration) Option {
	return func(o *aggregateOptions) {
		o.flushInterval = interval
	}
}

// Deferred sets a meter to be aggregated until the underlying context is Done.
func Deferred() Option {
	return func(o *aggregateOptions) {
		o.flushInterval = time.Duration(math.MaxInt64)
	}
}

// withDoneCh sets a channel to be closed when a meter is flushed for the last time.  It's only
// used for testing.
func withDoneCh(ch chan struct{}) Option {
	return func(o *aggregateOptions) {
		o.doneCh = ch
	}
}

const defaultFlushInterval = 10 * time.Second

// aggregatedMeter is a meter that emits on write only if flushInterval has passed.
type aggregatedMeter struct {
	sync.Mutex
	aggregateOptions           // Config options.
	meter            string    // The name of the meter.
	dirty            bool      // Whether or not the meter needs to be flushed.
	last             time.Time // When the meter was last flushed.
}

func (m *aggregatedMeter) init(meter string, options []Option) {
	m.meter = meter
	m.last = time.Now()
	m.flushInterval = defaultFlushInterval
	for _, opt := range options {
		opt(&m.aggregateOptions)
	}
}

func (m *aggregatedMeter) closeDoneCh() {
	if m.doneCh != nil {
		close(m.doneCh)
	}
}

type gauge[T any] struct {
	aggregatedMeter
	value T
}

type counter[T Monoid] struct {
	aggregatedMeter
	delta T
}

type delta[T Signed] struct {
	aggregatedMeter
	flushEvery T // Flush immediately if delta is > this.
	prev       T // The flushed absolute value.
	current    T // The current unflushed value.
}

func shouldFlush(now bool, m *aggregatedMeter) bool { // must hold lock
	return m.dirty && (now || time.Since(m.last) > m.flushInterval)
}

func (g *gauge[T]) flush(ctx context.Context, skip int, now bool) {
	g.Lock()
	defer g.Unlock()
	if shouldFlush(now, &g.aggregatedMeter) {
		logGauge(ctx, skip, g.meter, g.value)
		g.dirty = false
		g.last = time.Now()
	}
}

func (c *counter[T]) flush(ctx context.Context, skip int, now bool) {
	c.Lock()
	defer c.Unlock()
	if shouldFlush(now, &c.aggregatedMeter) {
		logCounter(ctx, skip, c.meter, c.delta)
		c.delta = 0
		c.dirty = false
		c.last = time.Now()
	}
}

func (d *delta[T]) flush(ctx context.Context, skip int, now bool) {
	d.Lock()
	defer d.Unlock()
	delta := d.current - d.prev
	if delta != 0 && (shouldFlush(now, &d.aggregatedMeter) || (d.flushEvery != 0 && (delta >= d.flushEvery || delta <= -d.flushEvery))) {
		if d.prev == 0 {
			logGauge(ctx, skip, d.meter, d.current)
		} else {
			logCounter(ctx, skip, d.meter, delta)
		}
		d.prev, d.current = d.current, 0
		d.dirty = false
		d.last = time.Now()
	}
}

func (g *gauge[T]) set(ctx context.Context, val T) {
	g.Lock()
	g.value = val
	g.dirty = true
	g.Unlock()
	g.flush(ctx, 2, false)
}

func (c *counter[T]) inc(ctx context.Context, val T) {
	c.Lock()
	c.delta += val // val == 0 means to try flushing
	c.dirty = true
	c.Unlock()
	c.flush(ctx, 2, false)
}

func (d *delta[T]) set(ctx context.Context, val T) {
	d.Lock()
	d.current = val
	if d.current != d.prev {
		d.dirty = true
	}
	d.Unlock()
	d.flush(ctx, 2, false)
}

func (g *gauge[T]) String() string {
	g.Lock()
	defer g.Unlock()
	return fmt.Sprintf("{%p gauge %q: %v}", g, g.meter, g.value)
}

func (c *counter[T]) String() string {
	c.Lock()
	defer c.Unlock()
	return fmt.Sprintf("{%p counter %q: %v}", c, c.meter, c.delta)
}

func (d *delta[T]) String() string {
	d.Lock()
	defer d.Unlock()
	return fmt.Sprintf("{%p delta %q: prev=%v cur=%v delta=%v}", d, d.meter, d.prev, d.current, d.current-d.prev)
}

// NewAggregatedGauge returns a context configured in such a way as to cause all calls to Set on
// this meter to be aggregated.
//
// Do not call this directly; use pctx.WithGauge.
func NewAggregatedGauge[T any](ctx context.Context, meter string, zero T, options ...Option) context.Context {
	create := func(ctx context.Context) context.Context {
		g := new(gauge[T])
		g.init(meter, options)
		go func() {
			<-ctx.Done()
			g.flush(ctx, 0, true)
			g.closeDoneCh()
		}()
		return context.WithValue(ctx, aggregatedMetersKey{meter: meter}, g)
	}
	return updateMeterIndex(ctx, create)
}

// NewAggregatedCounter returns a context configured in such a way as to cause all calls to Inc on
// this meter to be aggregated.
//
// Do not call this directly; use pctx.WithCounter.
func NewAggregatedCounter[T Monoid](ctx context.Context, meter string, zero T, options ...Option) context.Context {
	create := func(ctx context.Context) context.Context {
		c := new(counter[T])
		c.init(meter, options)
		go func() {
			<-ctx.Done()
			c.flush(ctx, 0, true)
			c.closeDoneCh()
		}()
		return context.WithValue(ctx, aggregatedMetersKey{meter: meter}, c)
	}
	return updateMeterIndex(ctx, create)
}

// NewAggregatedDelta returns a context configured in such a way as to cause all calls to Set on
// this meter to be aggregated, and output as a counter instead of a gauge.  For example, Set(0),
// Set(10), Set(20) will result in delta=10 being logged twice.  The parameter threshold will cause
// an immediate flush when the delta equals or exceeds this value; 0 turns off immediate flushing
// and the delta is only logged when it is time to do so (based on WithFlushInterval/Deferred).
//
// Note that deltas are constrained to signed types; that's because while the underlying value might
// be unsigned, the sign appears while taking deltas.  (Consider Set(100), Set(50); the delta is now
// no longer the same as the underlying type!)
//
// Do not call this directly; use pctx.WithDelta.
func NewAggregatedDelta[T Signed](ctx context.Context, meter string, threshold T, options ...Option) context.Context {
	create := func(ctx context.Context) context.Context {
		d := new(delta[T])
		d.init(meter, options)
		d.flushEvery = threshold
		go func() {
			<-ctx.Done()
			d.flush(ctx, 0, true)
			d.closeDoneCh()
		}()
		return context.WithValue(ctx, aggregatedMetersKey{meter: meter}, d)
	}
	return updateMeterIndex(ctx, create)
}

func updateMeterIndex(ctx context.Context, create func(context.Context) context.Context) context.Context {
	var result parentMeters
	if p, ok := ctx.Value(allParentMeters{}).(parentMeters); ok {
		result = append(result, p...)
	}
	result = append(result, create)
	return context.WithValue(create(ctx), allParentMeters{}, result)
}

// WithNewFields returns a context with new copies of all aggregated meters contained in the
// context chain.
func WithNewFields(ctx context.Context) context.Context {
	creators, ok := ctx.Value(allParentMeters{}).(parentMeters)
	if !ok {
		return ctx
	}
	for _, create := range creators {
		ctx = create(ctx)
	}
	return context.WithValue(ctx, allParentMeters{}, creators)
}
