package chunk

import (
	"context"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"golang.org/x/sync/semaphore"
)

func TestTaskChain(t *testing.T) {
	ctx := context.Background()
	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		numTasks := 5
		tc := NewTaskChain(ctx, semaphore.NewWeighted(int64(numTasks)))
		doneParallel := make(chan int, numTasks)
		doneSerial := make(chan int, numTasks)
		for i := 0; i < numTasks; i++ {
			i := i
			require.NoError(t, tc.CreateTask(func(_ context.Context) (func() error, error) {
				time.Sleep(time.Duration(numTasks-i-1) * time.Second)
				doneParallel <- i
				return func() error {
					doneSerial <- i
					return nil
				}, nil
			}))
		}
		require.NoError(t, tc.Wait())
		for i := 0; i < numTasks; i++ {
			require.Equal(t, numTasks-i-1, <-doneParallel)
		}
		for i := 0; i < numTasks; i++ {
			require.Equal(t, i, <-doneSerial)
		}
	})
	t.Run("Window", func(t *testing.T) {
		t.Parallel()
		numTasks := 10
		tc := NewTaskChain(ctx, semaphore.NewWeighted(int64(numTasks/2)))
		done := make(chan int, numTasks)
		for i := 0; i < numTasks; i++ {
			i := i
			require.NoError(t, tc.CreateTask(func(_ context.Context) (func() error, error) {
				time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
				return func() error {
					done <- i
					return nil
				}, nil
			}))
		}
		require.NoError(t, tc.Wait())
		for i := 0; i < numTasks; i++ {
			require.Equal(t, i, <-done)
		}
	})
	t.Run("Error", func(t *testing.T) {
		t.Parallel()
		tc := NewTaskChain(ctx, semaphore.NewWeighted(1))
		errMsg := "task errored"
		require.NoError(t, tc.CreateTask(func(_ context.Context) (func() error, error) {
			return nil, errors.New(errMsg)
		}))
		// The error may not be acknowledged by the next create task call because of goroutine scheduling.
		// So, we ignore the error.
		tc.CreateTask(func(_ context.Context) (func() error, error) { //nolint:errcheck
			time.Sleep(time.Second)
			return func() error { return nil }, nil
		})
		// The error will always be acknowledged by the followup task because it will be forced to wait.
		err := tc.CreateTask(func(_ context.Context) (func() error, error) { return func() error { return nil }, nil })
		require.YesError(t, err)
		require.True(t, strings.Contains(err.Error(), errMsg))
		err = tc.Wait()
		require.YesError(t, err)
		require.True(t, strings.Contains(err.Error(), errMsg))
	})
}
