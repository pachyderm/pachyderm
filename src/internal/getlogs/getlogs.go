package getlogs

import (
	"context"
  "io"
  "sync"

	"golang.org/x/sync/errgroup"

)

type GetLogs struct {
  Streamers []LogSource
  Matchers LineMatcher
  LockStreams bool
}

type LogSource interface {
  Stream(context.Context) (io.ReadCloser, error)
}

// LineMatcher Match() selects lines that match its criteria and returns them through sendMatch
type LineMatcher interface {
  Match(inputStream io.Reader, sendMatch func([]byte) error) error
}

func (gl *GetLogs) StartMatching(ctx context.Context, eg errgroup.Group, sendMatch func([]byte) error) error {
	var mtx sync.Mutex
  eg.Go(func() error {
    for _, streamer := range gl.Streamers {
      streamer := streamer
      if gl.LockStreams {
        mtx.Lock()
      }
      eg.Go(func() (retErr error) {
        if gl.LockStreams {
          defer mtx.Unlock()
        }
        stream, err := streamer.Stream(ctx)
        if err != nil {
          return err
        }
        defer func() {
          if err := stream.Close(); err != nil && retErr == nil {
            retErr = err
          }
        }()
        gl.Matchers.Match(stream, sendMatch)
        return retErr
      })
    }
    return nil
  })

  return nil
}

