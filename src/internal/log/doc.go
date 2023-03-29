// Package log is Pachyderm's logger.
//
// As a developer in the Pachyderm codebase, logging should be very easy.  All context.Contexts
// available throughout Pachyderm are capable of logging a message associated with that request or
// subsystem.
//
// # HOW TO LOG
//
// The API surface that we provide is simply:
//
//	Debug(ctx, message, [fields...])
//	Info(ctx, message, [fields...])
//	Error(ctx, message, [fields...])
//
// The fields are zap.Field objects; 99% of the time the field you're looking for is in the zap
// package, but we have a few in this package as well.  "go doc zap.Field" and "go doc log.Field"
// will list them all; there are about 100 of them.
//
// # LEVELS
//
// Debug messages are interesting to you, the developer.  Info messages are interesting to the team
// running Pachyderm day-to-day.  If you want an operator or support to see a message, choose the
// Info level.  Error messages are for actionable errors; if you can't handle a request and there is
// no way for an end user to see an error message, choose Error.  (It's assumed that the operator or
// Pachyderm support will have to jump in whenever these show up.  Misleading errors, especially
// errors that resolve themselves after a retry, are a big waste of customer time.)
//
// Error messages that end users should see must be in the API response or UI, not in log messages,
// of course.  End users should never have to read the logs.
//
// # STRUCTURED LOGGING
//
// All logs are structured; prefer:
//
//	Debug(ctx, "did a thing", zap.String("thing", thing), zap.String("reason", "because"))
//
// over a non-structured:
//
//	Debugf(ctx, "did a thing (%v) because %v", thing, "because")
//
// Structured logging is nice because you don't have to think about how to nicely phrase the message
// to make the fields make sense; just stick as much data in there as you like.  And, automation or
// people debugging can filter for relevant messages with tools like jlog
// <https://github.com/jrockway/json-logs> or jq.
//
// It is conventional to name your fields with camelCase, not snake_case or whatever-this-is-case.
// When unpacking a compound object into individual fields, object.key1, object.key2 are acceptable.
// (Example: grpc.code, grpc.message in gRPC responses.)
//
// # SPANS
//
// A common pattern is logging how long a certain operation ran for, and whether or not it
// succeeded.  Most uses are covered by:
//
//	func doSomething(ctx context.Context, to *Object) (retErr error) {
//	    defer log.Span(ctx, "doSomething", zap.Object("to", to))(log.Errorp(&retErr))
//	    do()
//	    things()
//	    and()
//	    return them()
//	}
//
// This will log:
//
//	DEBUG doSomething; span starting {to: {my: object}}
//	DEBUG doSomething; span finished ok {span_duration: 100ms, to: {my: object}}
//
// or
//
//	DEBUG doSomething; span starting {to: {my: object}}
//	DEBUG doSomething; span failed {span_duration: 100ms, error: "them failed", errorVerbose: "<stack trace>", to: {my: object}}
//
// There is `SpanL` if you want to set a log level other than DEBUG.  log.Span takes fields as well,
// if you want to annotate the span with an ID or other information to help you assocaite the "end"
// log with the "start" log.  The fields will be logged for both the start and end events.
//
// There are the fields `Errorp` (for failing if a pointer to an error is non-nil, useful with
// retErr as above), `ErrorpL` (to change the log level on failure), zap.Error (for an error that is
// ready when you call the callback), and `ErrorL` (to change the log level if the error is
// non-nil).  (As an aside, `zap.Error` won't log anything if its argument is nil, so you can always
// include an error without checking for nil-ness first.  Our wrappers also conform to that
// convention.)
//
// There is also SpanContext, where the operation itself wants to add details to the span:
//
//	func doSomething(ctx context.Context) (retErr error) {
//	    ctx, end := log.SpanContext(ctx, "doSomething", zap.String("theThing", "coolThing"))
//	    defer end(log.ErrorpL(&retErr, log.LevelError))
//	    do(ctx)
//	    things(ctx)
//	    and(ctx)
//	    return them(ctx)
//	}
//
// This will log:
//
//	DEBUG doSomething doSomething; span starting {theThing: coolThing}
//	DEBUG doSomething inside do {theThing: coolThing}
//	DEBUG doSomething inside things {theThing: coolThing}
//	INFO  doSomething them is having trouble {theThing: coolThing, howMuchTrouble: a lot}
//	ERROR doSomething doSomething; span failed {span_duration: 1ms, theThing: coolThing, error: "them failed"}
//
// There is `SpanContextL` for controlling the level of the starting message.
//
// Finally, there is `LogStep`, which used to be called middleware.LogStep.  It takes a callback
// instead of requiring you to defer things with pointers.
//
// That is all you need to know to start logging things!  The rest of the functionality is more
// advanced.
//
// # SAMPLING
//
// Some subsystems can be pretty noisy.  The loggers we provide are configured to rate-limit
// messages over a short period of time (on the order of 2 messages every 3 seconds).  This lets you
// see the first few log messages of a noisy routine, without filling up all storage space with one
// log message.
//
// Rate limiting is keyed on message and severity; one noisy message won't rate limit other less
// frequent messages.  Design your messages and log levels accordingly.  For example, in the GRPC
// logging interceptor, we include the service name and method in the log message (in addition to
// fields), simply to ensure that one frequent GRPC call doesn't prevent other calls from being
// logged.  Severity counts as well, so if you're writing a backoff routine, the first attempts
// might log at level DEBUG, but the last attempt can log at INFO.  This ensures that even after a
// bunch of retries in a short period of time, the final outcome will still appear in the logs.
// (The backoff package does not yet do this.)
//
// At the end of the day, it's about 100ns per message to log to a rate-limited logger (including
// some non-rate-limited messages; see BenchmarkFieldsSampled), so don't worry about logging too
// much.  (Normal messages take about 2µs to process and render; most of this time is spent
// providing the "caller" field.)
//
// You should feel perfectly free to log things like:
//
//	for i := 0; i < 10000; i++ {
//	    DoThing(items, i)
//	    Debug(ctx, "did a thing to items", zap.Int("i", i), zap.Int("max", 10000))
//	}
//
// If it proceeds quickly, there will only be a few log lines.  If DoThing is unexpectedly slow
// sometimes, then those logs will magically appear.  Just don't change the message or level every
// iteration.
//
// # GETTING A LOGGER
//
// By the time you read this, all code should have been converted to transmit loggers through the
// context.  If you run into some code that can't accept contexts, use `pctx.TODO()` if you can't fix
// the API today.  TODO uses the global logger, but doesn't warn about the logger not being
// initialized, as Child() would do on an empty context.  (For saftey, there is no circumstance
// where a log function will ever panic, or not log, but we do warn if the provided context doesn't
// have a logger inside of it.  It's probably a bug, and the stack trace attached to the warning
// should let you track it down easily.)
//
// Top-level applications need to initialize the global logger, from which background services and
// pctx.TODO() derive the global logger from.  We have three ways of doing this, InitPachdLogger,
// InitWorkerLogger, and InitPachctlLogger.  The Pachd Logger is designed for compatability with
// third-party logging systems, like Stackdriver, and should be the default for any server-side
// applications.  The Worker logger is designed to only emit messages that can be parsed by `pachctl
// logs`, and is used only by the worker (user code container, not the storage container).  The
// Pachctl logger is for command-line apps, and pretty-prints messages similar to those in this
// documentation.  (It also adds color if the terminal is capable and it's not disabled by pachctl
// flags.)  If you're writing a new CLI utility, you probably want the Pachctl logger.
//
// The root context in applications is `pctx.Background()`.  Call it exactly once after
// Init(Pachd|Worker|Pachctl)Logger.
//
// In tests, use pctx.TestContext(t).  In the future we plan to add other accouterments to contexts
// inside Pachyderm; use of this function will ensure that your tests are ready.  The test context
// contains a logger scoped to that test; it is safe to run tests in parallel that both use
// different TestContexts.  Anything logged to that logger will appear in the test logs (go test -v
// ...) in a format similar to pachctl's logs.  For logging inside tests, rather than in production
// code called by tests, just use `t.Log`.
//
// Do not use `zap.L()` and `zap.S()` outside of this package.  They are set up and work, but prefer
// the context propagation.
//
// # CONFIGURING LOGGING
//
// Logging is configured by the environment, unrelated to service environment initialization.
//
// The knobs you can tweak are:
//
//   - PACHYDERM_LOG_LEVEL={debug, info, error}. That controls the minimum serverity log messages
//     that will be printed.  We default to info, but it's configurable in the helm chart.
//
//   - PACHYDERM_DEVELOPMENT_LOGGER=1.  If set to "1" or "true", then any warnings generated by the
//     logging system will panic.  It can't be turned on in the helm chart; these warnings should
//     never reach production releases, but if they do, we just want to carry on and not crash.
//
//   - PACHYDERM_DISABLE_LOG_SAMPLING=1.  If set to "1" or "true", no log sampling will occur.  This
//     might be helpful if there are a lot of duplicate messages that you really need to see in
//     order to debug.  It can't be turned on in the helm chart.
//
// If you mess up the syntax of any of these environment variables, a warning will be the first
// message printed to the logger after it starts up.  These will also panic in development mode.
//
// At startup, the values of these configuration items will be stored and propagated to all workers.
// That's why they have the PACHYDERM_ prefix, so as not to conflict with user code.
//
// You can change the log level across the entire fleet of services at runtime with `pachctl debug
// log-level ...`.
//
// # MISCELLANEOUS
//
// Some third-party packages expect a certain type of logger.  We typically add adaptors to this
// package to convert a context into a logger for that subsystem.
//
// See:
//
//	NewAmazonLogger(ctx)         # For the S3 client
//	NewLogrus(ctx)               # For logrus users, like dex and snowflake.
//	NewStdLogAt(ctx, level)      # For "log" users, like the net/http server.
//	AddLoggerToEtcdServer(ctx, server)        # To add a logger to an in-memory etcd.
//	AddLoggerToHttpServer(ctx, name, server)  # To add a logger to incoming requests in a net/http server.
//
// Do not use a NewLogrus or a NewStdLogAt as the primary logger in any code you're writing.  They
// are only for adapting to the needs of third-party packages.
//
// # EXTENDING THE LOGGING PACKAGE
//
// Sometimes the loggers available might not be quite what you need.  Add the utility function you
// need to this package.  To test it, there are private methods in testing.go that create a logger
// for testing the logging system.  They create a pachd-style logger, that logs to an object you can
// iterate over to extract the parsed messages.
//
// Because logging is widespread throughout the codebase, this package should not have any
// dependencies inside Pachyderm.  For example, there is a function in here that generates UUIDs,
// but it can't depend on internal/uuid, because internal/uuid depends on internal/backoff, which
// depends on this package.  It should always be safe to add logging no matter where you are in the
// codebase.
package log
