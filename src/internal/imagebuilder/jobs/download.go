package jobs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/zeebo/xxh3"
	"go.starlark.net/starlark"
	"go.uber.org/zap"
)

// DownloadedFile is a file that has been downloaded.
type DownloadedFile struct {
	NameAndPlatform
	File *File
}

func (f *DownloadedFile) FS() fs.FS {
	return f.File.FS()
}

// A job that downloads a file.
type Download struct {
	Name       string
	URL        string
	WantDigest Digest
	Platform   Platform
}

var _ Job = (*Download)(nil)

func (d Download) String() string {
	return fmt.Sprintf("<download %v={%s}#%v>", d.Name, d.URL, d.Platform)
}

func (d Download) ID() uint64 {
	return xxh3.HashString(d.Name + d.URL + d.WantDigest.String())
}

func (d Download) Inputs() []Reference { return nil }

func (d Download) Outputs() []Reference {
	return []Reference{
		NameAndPlatformAndFS{Name: "download:" + d.Name, Platform: d.Platform},
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
	if err := CheckHTTPStatus(res, http.StatusOK); err != nil {
		return nil, errors.Wrap(err, "download not done")
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
				Name: "download:" + d.Name + ":tar.zstd",
				Path: out.Path(),
				Digest: Digest{
					Algorithm: "blake3",
					Value:     hv,
				},
			},
		},
	}, nil
}

func (d *Download) Unpack(v starlark.Value) error {
	dict, ok := v.(*starlark.Dict)
	if !ok {
		return errors.New("expected download dict")
	}
	uv, ok, err := dict.Get(starlark.String("url"))
	if !ok {
		if err != nil {
			return errors.Wrap(err, "url")
		}
		return errors.New("missing required url value")
	}
	us, ok := starlark.AsString(uv)
	if !ok {
		return errors.Errorf("url value %v is not a string", uv)
	}
	d.URL = us
	dv, ok, err := dict.Get(starlark.String("digest"))
	if err != nil {
		return errors.Wrap(err, "digest")
	}
	if !ok {
		return nil
	}
	ds, ok := starlark.AsString(dv)
	var digest Digest
	if err := digest.UnmarshalText([]byte(ds)); err != nil {
		return errors.Wrap(err, "unmarshal digest")
	}
	d.WantDigest = digest
	return nil
}

type downloadSet []Download

func (x *downloadSet) Unpack(v starlark.Value) error {
	dict, ok := v.(*starlark.Dict)
	if !ok {
		return errors.New("expected dict of platform -> downloads")
	}
	for _, k := range dict.Keys() {
		v, ok, err := dict.Get(k)
		if err != nil {
			return errors.Wrapf(err, "args[%v]", k)
		}
		if !ok {
			return errors.Errorf("args[%v]: not found?", k)
		}
		platform, ok := starlark.AsString(k)
		if !ok {
			return errors.Errorf("args[%v]: not a string?", k)
		}
		var d Download
		d.Platform = Platform(platform)
		if err := d.Unpack(v); err != nil {
			return errors.Wrapf(err, "args[%v]: unpack", k)
		}
		*x = append(*x, d)
	}
	return nil
}

func (Download) NewFromStarlark(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) ([]Job, error) {
	var name string
	var byPlatform downloadSet
	if len(args) > 0 {
		return nil, errors.New("unexpected positional args")
	}
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"name", &name,
		"by_platform?", &byPlatform,
	); err != nil {
		return nil, errors.Wrap(err, "unpack")
	}
	var result []Job
	for _, j := range byPlatform {
		j.Name = name
		result = append(result, j)
	}
	return result, nil
}
