package jobs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/zeebo/xxh3"
	"go.uber.org/zap"
)

// A job that downloads a file.
type Download struct {
	Name       string
	URL        string
	WantDigest Digest
	Platform   Platform
}

var _ Job = (*Download)(nil)

func (d Download) String() string {
	return fmt.Sprintf("<download %v=%q#%v>", d.Name, d.URL, d.Platform)
}

func (d Download) ID() uint64 {
	return xxh3.HashString(d.Name + d.URL + d.WantDigest.String())
}

func (d Download) Inputs() []Reference { return nil }

func (d Download) Outputs() []Reference {
	return []Reference{
		NameAndPlatform{Name: "download:" + d.Name, Platform: d.Platform},
	}
}

func (d Download) Run(ctx context.Context, jc *JobContext, inputs []Artifact) (_ []Artifact, retErr error) {
	ctx, done := log.SpanContext(ctx, "download", zap.String("url", d.URL))
	defer done(log.Errorp(&retErr))

	h, err := d.WantDigest.Hash()
	if err != nil {
		return nil, errors.Wrap(err, "get hasher")
	}
	req, err := http.NewRequestWithContext(ctx, "GET", d.URL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "new HTTP request")
	}
	res, err := jc.HTTPClient.Do(req)
	if err != nil {
		return nil, WrapRetryable(errors.Wrap(err, "do HTTP request"))
	}
	defer errors.Close(&retErr, res.Body, "close body")
	if got, want := res.StatusCode, http.StatusOK; got != want {
		return nil, WrapRetryable(errors.Errorf("unexpected HTTP status %s", res.Status))
	}
	out, err := jc.Cache.NewCacheableFile("download-" + url.PathEscape(d.URL))
	if err != nil {
		return nil, errors.Wrap(err, "new cacheable file")
	}

	var tee io.Writer = out
	if d.WantDigest.Algorithm != "blake3" {
		// CacheableFile already calculates a blake3 hash.
		tee = io.MultiWriter(out, h)
	}
	if _, err := io.Copy(tee, res.Body); err != nil {
		return nil, WrapRetryable(errors.Wrap(err, "read body"))
	}
	if err := out.Close(); err != nil {
		return nil, errors.Wrap(err, "close output file")
	}
	hv := out.Sum(nil)
	if d.WantDigest.Algorithm != "blake3" {
		hv = h.Sum(nil)
	}
	if !bytes.Equal(hv, d.WantDigest.Value) {
		return nil, errors.Errorf("Downloaded file failed integrity check:\n   got: %s:%x\n  want: %s:%x", d.WantDigest.Algorithm, hv, d.WantDigest.Algorithm, d.WantDigest.Value)
	}
	return []Artifact{
		&DownloadedFile{
			NameAndPlatform: NameAndPlatform{
				Name:     "download:" + d.Name,
				Platform: d.Platform,
			},
			File: &File{
				Path: out.Path(),
				Digest: Digest{
					Algorithm: "blake3",
					Value:     hv,
				},
			},
		},
	}, nil
}

type DownloadedFile struct {
	NameAndPlatform
	File *File
}
