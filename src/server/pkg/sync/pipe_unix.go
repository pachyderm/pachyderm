// +build !windows

package sync

import (
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"
)

func (p *Puller) makePipe(path string, f func(io.Writer) error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	if err := syscall.Mkfifo(path, 0666); err != nil {
		return err
	}
	func() {
		p.Lock()
		defer p.Unlock()
		p.pipes[path] = true
	}()
	// This goro will block until the user's code opens the
	// fifo.  That means we need to "abandon" this goro so that
	// the function can return and the caller can execute the
	// user's code. Waiting for this goro to return would
	// produce a deadlock. This goro will exit (if it hasn't already)
	// when CleanUp is called.
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		if err := func() (retErr error) {
			file, err := os.OpenFile(path, os.O_WRONLY, os.ModeNamedPipe)
			if err != nil {
				return err
			}
			defer func() {
				if err := file.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			// If the CleanUp routine has already run, then there's
			// no point in downloading and sending the file, so we
			// exit early.
			if func() bool {
				p.Lock()
				defer p.Unlock()
				delete(p.pipes, path)
				return p.cleaned
			}() {
				return nil
			}
			w := &sizeWriter{w: file}
			if err := f(w); err != nil {
				return err
			}
			atomic.AddInt64(&p.size, w.size)
			return nil
		}(); err != nil {
			select {
			case p.errCh <- err:
			default:
			}
		}
	}()
	return nil
}
