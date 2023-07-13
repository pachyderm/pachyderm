package task

import (
	"context"

	"google.golang.org/protobuf/types/known/anypb"
)

// Service manages the distributed processing of tasks.

// Task:
// A task is a unit of work that can be processed by a worker.
// The input / output of a task should be fully specifiable by an Any proto.
// The processing of a task should produce a single output Any or an error.
// A task when run may have side effects, but is idempotent.
// A task can be run 0 or more times.
// A task can be run 0 times when a task with the same description (hash of the input Any) has been
// processed already (only applicable when using a caching layer).
// A task can be run multiple times if issues outside of the task processing occur.
// A task must allow for two or more workers to process it concurrently and perform side effects correctly.
// Tasks are managed by Doers and provided to client workers for processing through Sources.
// NOTE: In general, tasks should be designed to take at least seconds and ideally minutes because
// that is usually a good balance of being long enough to minimize the overhead of laying out the
// work and not being too long to restart from scratch.

// Namespace:
// A task managed by a Service has a namespace.
// The namespace is used for scoping the Doers and Sources of tasks such that clients can create workers
// that only process tasks created by clients in the same namespace.

// Scheduling:
// A task managed by a Service has a group.
// The group is used for scheduling purposes.
// Scheduling is based on maximizing fairness for the groups, and the schedulable unit is a task.
type Service interface {
	// NewDoer creates a Doer with the provided namespace and group.
	NewDoer(namespace, group string, cache Cache) Doer
	// NewSource creates a Source with the provided namespace.
	NewSource(namespace string) Source
	// List calls a function on every task under a namespace and group
	List(ctx context.Context, namespace, group string, cb func(namespace, group string, data *Task, claimed bool) error) error
}

// Doer is a doer of tasks.
// Refer to the DoOne and DoBatch helper functions if a simpler interface is desired.
type Doer interface {
	// Do creates and returns the results of processing a stream of tasks provided
	// by the input channel. The client should close the input channel when all tasks have
	// been sent (it does not need to be closed if the context is canceled). For each
	// task, the collect function will be called with the results.
	Do(ctx context.Context, inputChan chan *anypb.Any, cb CollectFunc) error
}

// Source is a source of tasks.
type Source interface {
	// Iterate iterates through tasks until the provided context is canceled.
	// For each task, the process function will be called and the results
	// will be returned to the Doer that created the task.
	Iterate(ctx context.Context, cb ProcessFunc) error
}

// CollectFunc is the type of a function that is used for collecting the output of a stream / batch of tasks.
// Index is the index of a task with respect to the order in which the task was created in the stream / batch.
type CollectFunc = func(index int64, output *anypb.Any, _ error) error

// ProcessFunc is the type of a function that is use for processing a task.
// If an error occurs, then it should be returned.
// This error will be propagated back to the Doer that created the task.
type ProcessFunc = func(ctx context.Context, input *anypb.Any) (output *anypb.Any, _ error)

type Cache interface {
	Get(ctx context.Context, key string) (output *anypb.Any, _ error)
	Put(ctx context.Context, key string, output *anypb.Any) error
}
