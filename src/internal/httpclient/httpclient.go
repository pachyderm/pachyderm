package httpclient

import (
	"io"
	"net/http"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"
)

type bodyCloseReader struct {
	io.ReadCloser
	didRead bool
	read    func()
	done    func()
}

func (r *bodyCloseReader) Read(b []byte) (int, error) {
	if !r.didRead {
		r.read()
		r.didRead = true
	}
	return r.ReadCloser.Read(b)
}

func (r *bodyCloseReader) Close() error {
	r.done()
	return r.ReadCloser.Close()
}

type RoundTripper struct {
	http.RoundTripper
}

func (rt *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, done := log.SpanContext(req.Context(), "http client", zap.String("uri", req.URL.String()))
	doneCh := make(chan string, 1)
	go func() {
		msg := "request ongoing"
		for {
			select {
			case newMsg, ok := <-doneCh:
				if !ok {
					return
				}
				msg = newMsg
			case <-time.After(5 * time.Second):
			}
			log.Debug(ctx, msg)

		}
	}()
	res, err := rt.RoundTripper.RoundTrip(req)
	if err != nil {
		done(zap.Error(err))
		return res, err
	}
	log.Debug(ctx, "http request finished", zap.String("status", res.Status), zap.Int("code", res.StatusCode))
	res.Body = &bodyCloseReader{
		ReadCloser: res.Body,
		read: func() {
			doneCh <- "started reading body"
		},
		done: func() {
			done()
			close(doneCh)
		},
	}
	return res, err
}
