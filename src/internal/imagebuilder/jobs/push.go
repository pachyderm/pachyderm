package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strconv"
	"strings"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/zeebo/xxh3"
	"go.starlark.net/starlark"
	"go.uber.org/zap"
)

type PushResult struct{}

type Tag struct {
	Scheme, Registry, Repository, Image, Tag string
}

func (t Tag) String() string {
	return path.Join(t.Registry, t.Repository, t.Image) + ":" + t.Tag
}

// PushRequest is a request to push some blobs in some order.
type PushRequest struct {
	NameAndPlatform
	Sequence [][]Blob
	Manifest *v1.Manifest
}

// PushOutput is the result of a push.
type PushOutput struct {
	NameAndPlatform
	Pushed      []string
	ManifestURL string
}

func (o PushOutput) String() string {
	return fmt.Sprintf("<pushed manifest %v (%v) %v", o.ManifestURL, o.NameAndPlatform, o.Pushed)
}

// DoPush is a job that executes PushRequests.
type DoPush struct {
	NameAndPlatform
	Tag
	Objects []Reference
}

var _ Job = (*DoPush)(nil)

func (p DoPush) ID() uint64 {
	return xxh3.HashString(fmt.Sprint(p))
}

func (p DoPush) String() string {
	return fmt.Sprintf("<push [%v] to %v #%v>", p.Objects, p.Tag, p.Platform)
}

func (p DoPush) Inputs() []Reference {
	return p.Objects
}

func (p DoPush) Outputs() []Reference {
	return []Reference{
		NameAndPlatform{
			Name:     "pushed:" + p.Name,
			Platform: p.Platform,
		},
	}
}

func (p DoPush) Run(ctx context.Context, jc *JobContext, inputs []Artifact) ([]Artifact, error) {
	image := fmt.Sprintf("%v%v/%v/%v", p.Scheme, p.Registry, p.Repository, p.Image)
	if !jc.PushAllowed(image) {
		return nil, errors.Errorf("push denied because %v does not match any allowed push prefix", image)
	}

	var reqs []PushRequest
	for _, in := range inputs {
		req, ok := in.(PushRequest)
		if !ok {
			return nil, errors.Errorf("%v is not a PushRequest", in)
		}
		reqs = append(reqs, req)
	}

	var result []string
	var pushErrs error
	var manifestURL string
	for _, preq := range reqs {
		for _, seq := range preq.Sequence {
			for _, blob := range seq {
				if err := pushBlob(ctx, jc, p.Tag, blob); err != nil {
					errors.JoinInto(&pushErrs, fmt.Errorf("push blob %v: %w", blob.Digest(), err))
					continue
				}
				result = append(result, string(blob.Digest()))
			}
		}
		if m := preq.Manifest; m != nil {
			var err error
			manifestURL, err = pushManifest(ctx, jc, p.Tag, *m)
			if err != nil {
				errors.JoinInto(&pushErrs, fmt.Errorf("push manifest: %w", err))
			}
		}
	}
	if pushErrs != nil {
		return nil, errors.Wrap(pushErrs, "push blobs")
	}
	return []Artifact{
		PushOutput{
			NameAndPlatform: NameAndPlatform{
				Name:     "pushed:" + p.Name,
				Platform: p.Platform,
			},
			Pushed:      result,
			ManifestURL: manifestURL,
		},
	}, nil
}

func startPushSession(ctx context.Context, jc *JobContext, tag Tag) (_ *url.URL, retErr error) {
	ctx, done := log.SpanContext(ctx, "session")
	defer done(log.Errorp(&retErr))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%v%v/v2/%v/%v/blobs/uploads/", tag.Scheme, tag.Registry, tag.Repository, tag.Image), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "create http request")
	}
	res, err := jc.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "do request")
	}
	defer errors.Close(&retErr, res.Body, "close response body")
	if err := CheckHTTPStatus(res, http.StatusAccepted); err != nil {
		if dump, err := httputil.DumpResponse(res, true); err == nil {
			log.Debug(ctx, "response body", zap.ByteString("dump", dump))
		}
		return nil, errors.Wrap(err, "session not created")
	}
	location := res.Header.Get("location")
	if location == "" {
		return nil, errors.New("empty redirect location")
	}
	u, err := url.Parse(location)
	if err != nil {
		return nil, errors.Wrapf(err, "parse redirect location %v", u)
	}
	if !jc.PushAllowed(u.String()) {
		// Nice try, scumbag.
		return nil, errors.Errorf("registry redirected us to a forbidden location: %v (not in allowed push prefixes)", location)
	}
	return u, nil
}

func uploadBlob(ctx context.Context, jc *JobContext, location string, blob Blob) (retErr error) {
	ctx, done := log.SpanContext(ctx, "upload")
	defer done(log.Errorp(&retErr))

	r, err := blob.Open()
	if err != nil {
		return errors.Wrapf(err, "open blob %v", blob)
	}
	// Note: the HTTP client closes this reader for us.

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, location, r)
	if err != nil {
		return errors.Wrapf(err, "create upload http request")
	}
	req.Header.Set("content-length", strconv.FormatInt(blob.Size, 10))
	req.Header.Set("content-type", "application/octet-stream")
	res, err := jc.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "do upload request")
	}
	defer errors.Close(&retErr, res.Body, "close upload body")

	if err := CheckHTTPStatus(res, http.StatusCreated); err != nil {
		return errors.Join(registryError(res), errors.Wrap(err, "blob upload not accepted"))
	}
	return nil
}

func registryError(res *http.Response) error {
	if !strings.Contains(res.Header.Get("content-type"), "application/json") {
		return nil
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "read error body")
	}
	var info struct {
		Errors []struct {
			Code, Message string
			Detail        int
		}
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return errors.Wrap(err, "unmarshal json body")
	}
	var errs error
	for _, e := range info.Errors {
		errors.JoinInto(&errs, fmt.Errorf("error response: %v (%v, %v)", e.Message, e.Code, e.Detail))
	}
	return errs
}

func pushBlob(ctx context.Context, jc *JobContext, tag Tag, blob Blob) (retErr error) {
	ctx, done := log.SpanContext(ctx, "pushblob", zap.String("underlying", blob.Underlying.Path), zap.String("blob", string(blob.Digest())))
	defer done(log.Errorp(&retErr))

	redirect, err := startPushSession(ctx, jc, tag)
	if err != nil {
		return errors.Wrap(err, "start push session")
	}
	q := redirect.Query()
	q.Set("digest", string(blob.Digest()))
	redirect.RawQuery = q.Encode()
	if err := uploadBlob(ctx, jc, redirect.String(), blob); err != nil {
		return errors.Wrap(err, "put blob")
	}
	return nil
}

func pushManifest(ctx context.Context, jc *JobContext, tag Tag, m v1.Manifest) (_ string, retErr error) {
	ctx, done := log.SpanContext(ctx, "pushmanifest")
	defer done(log.Errorp(&retErr))

	js, err := json.Marshal(m)
	if err != nil {
		return "", errors.Wrap(err, "marshal manifest")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%v%v/v2/%v/%v/manifests/%v", tag.Scheme, tag.Registry, tag.Repository, tag.Image, tag.Tag), bytes.NewReader(js))
	if err != nil {
		return "", errors.Wrap(err, "create manifest request")
	}
	req.Header.Set("content-length", strconv.Itoa(len(js)))
	req.Header.Set("content-type", v1.MediaTypeImageManifest)
	res, err := jc.HTTPClient.Do(req)
	if err != nil {
		return "", errors.Wrapf(err, "do upload request")
	}
	defer errors.Close(&retErr, res.Body, "manifest upload body")

	if err := CheckHTTPStatus(res, http.StatusCreated); err != nil {
		return "", errors.Join(registryError(res), errors.Wrap(err, "manifest not accepted"))
	}
	return res.Header.Get("location"), nil
}

func (DoPush) NewFromStarlark(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) ([]Job, error) {
	reg := GetRegistry(thread)
	var outputs []Job
	for _, j := range reg.Jobs {
		for _, o := range j.Outputs() {
			if pr, ok := o.(WithName); ok && strings.HasPrefix(pr.GetName(), "pushable:") {
				platform := o.(WithPlatform).GetPlatform()
				outputs = append(outputs, DoPush{
					NameAndPlatform: NameAndPlatform{
						Name:     pr.GetName(),
						Platform: platform,
					},
					Tag: Tag{
						Scheme:     "http://",
						Registry:   "localhost:5001",
						Repository: "test",
						Image:      "test",
						Tag:        "test",
					},
					Objects: []Reference{o},
				})
			}
		}
	}
	return outputs, nil
}
