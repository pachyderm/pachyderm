package pachd

import (
	"context"
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"golang.org/x/sync/errgroup"
)

// setupSteps are run serially before the any of the background tasks
type setupStep struct {
	Name string
	Fn   func(context.Context) error
}

// base structures the setup and run phases of pachds
type base struct {
	setup      []setupStep
	background map[string]func(context.Context) error

	done bool
}

// addSetup adds a setupStep
func (b *base) addSetup(name string, fn func(context.Context) error) {
	b.setup = append(b.setup, setupStep{
		Name: name,
		Fn:   fn,
	})
}

// addBackground adds a background task
func (b *base) addBackground(name string, fn func(context.Context) error) {
	if b.background == nil {
		b.background = make(map[string]func(context.Context) error)
	}
	if _, exists := b.background[name]; exists {
		panic(fmt.Sprintf("2 background functions with same name %q", name))
	}
	b.background[name] = fn
}

// Run runs all the setup steps serially, returning the first error encountered.
// Then all the background tasks are run concurrently, if any of them error then
// the others are cancelled, and Run returns the error.
func (b *base) Run(ctx context.Context) error {
	if b.done {
		panic("pachd has already been run")
	}
	defer func() { b.done = true }()

	log.Info(ctx, "pachd begin setup")
	for _, step := range b.setup {
		endf := log.Span(ctx, step.Name)
		if err := step.Fn(ctx); err != nil {
			return errors.Errorf("during setup step %s: %w", step.Name, err)
		}
		endf()
	}
	log.Info(ctx, "pachd setup complete")

	log.Info(ctx, "pachd running services")
	eg, ctx := errgroup.WithContext(ctx)
	for name, fn := range b.background {
		name := name
		fn := fn
		ctx := pctx.Child(ctx, name)
		eg.Go(func() error {
			return fn(ctx)
		})
	}
	return errors.EnsureStack(eg.Wait())
}

func (b *base) printVersion(ctx context.Context) error {
	log.Info(ctx, "version info", log.Proto("versionInfo", version.Version))
	return nil
}
