// Package pctx implements contexts for Pachyderm.
//
// Contexts are the root of most tracing, logging, auth, etc.  This package manages a context that
// is set up to do all of those things.  (At the time of this writing, actually just logging, but
// the FUTURE is bright.)
//
// # GETTING A CONTEXT
//
// If you are creating a new application, use `Background` to get the root context and derive all
// future contexts from that.  HTTP and gRPC servers have an interceptor to do this already.
//
// If you need to gradually introduce contexts into an existing package, use `TODO` to get one to
// use as needed.
//
// # DERIVED CONTEXTS
//
// It goes without saying that it is perfectly safe to use context.WithTimeout, context.WithCancel,
// context.WithValue, etc.  The derived context will inherit the capabilities of its parent.
//
// Sometimes you want to spin-off a long-running operation with a name and some fields.  The Child
// function in this package takes care of this for you.  `Child(parent, "name", [options...])`.
// Each Child call changes the name of the underlying instrumentation, concatenating its own name
// with the parent's name and a dot.  So as you get deeper into children, you might see that the
// logs come from something like "PPS.master.pipelineController(default/edges).outgoingHttp",
// allowing you to see where about in the stack the data came from.
//
// The convention is to use oneCamelCaseWord for the logger name, and for parents to name their
// children.  Instead of:
//
//	func worker(ctx context.Context) {
//	    ctx = pctx.Child(ctx, "worker")
//	    ...
//	}
//
// Prefer:
//
//	go s.worker(pctx.Child(ctx, "worker"))
package pctx
