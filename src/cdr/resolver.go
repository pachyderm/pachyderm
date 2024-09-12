package cdr

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/cipher"
	"hash"
	"io"
	"net/http"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/chacha20"
	"google.golang.org/protobuf/proto"
)

type Option func(r *Resolver)

func WithHTTPClient(c *http.Client) Option {
	return func(r *Resolver) {
		r.httpClient = c
	}
}

type Resolver struct {
	httpClient *http.Client
	cache      cache
}

func NewResolver(opts ...Option) *Resolver {
	r := &Resolver{
		httpClient: http.DefaultClient,
		cache:      newCache(),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Resolver) Deref(ctx context.Context, ref *Ref) (io.ReadCloser, error) {
	return r.deref(ctx, ref, true)
}

func (r *Resolver) deref(ctx context.Context, ref *Ref, cache bool) (io.ReadCloser, error) {
	if data := r.cache.get(ref); data != nil {
		return io.NopCloser(bytes.NewReader(data)), nil
	}
	cacheRC := func(rc io.ReadCloser) (io.ReadCloser, error) {
		data, err := io.ReadAll(rc)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		r.cache.put(ref, data)
		return io.NopCloser(bytes.NewReader(data)), nil
	}
	var rc io.ReadCloser
	var err error
	switch body := ref.Body.(type) {
	case *Ref_Http:
		rc, err = r.derefHTTP(ctx, body.Http)
		if err != nil {
			return nil, err
		}
	case *Ref_ContentHash:
		rc, err = r.derefContentHash(ctx, body.ContentHash, false)
		if cache && IsImmutable(ref) {
			rc, err = cacheRC(rc)
		}
	case *Ref_Cipher:
		rc, err = r.derefCipher(ctx, body.Cipher, false)
		if err != nil {
			return nil, err
		}
		if cache && IsImmutable(ref) {
			rc, err = cacheRC(rc)
		}
	case *Ref_Compress:
		rc, err = r.derefCompress(ctx, body.Compress, false)
		if err != nil {
			return nil, err
		}
		if cache && IsImmutable(ref) {
			rc, err = cacheRC(rc)
		}
	case *Ref_Slice:
		rc, err = r.derefSlice(ctx, body.Slice, cache)
	case *Ref_Concat:
		rc, err = r.derefConcat(ctx, body.Concat, cache)
	default:
		return nil, errors.Errorf("unsupported Ref variant %v: %T", ref.Body, ref.Body)
	}
	if err != nil {
		return nil, err
	}
	return rc, nil
}

func (r *Resolver) derefHTTP(ctx context.Context, ref *HTTP) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ref.Url, nil)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	for k, v := range ref.Headers {
		req.Header.Add(k, v)
	}
	// TODO: Check response code.
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return resp.Body, nil
}

func (r *Resolver) derefContentHash(ctx context.Context, ref *ContentHash, cache bool) (io.ReadCloser, error) {
	innerRc, err := r.deref(ctx, ref.Inner, cache)
	if err != nil {
		return nil, err
	}
	// TODO: add a maximum limit here
	data, err := io.ReadAll(innerRc)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	var h hash.Hash
	switch ref.Algo {
	case HashAlgo_BLAKE2b_256:
		h, err = blake2b.New256(nil)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
	default:
		return nil, errors.Errorf("unrecognized hash algo %v", ref.Algo)
	}
	h.Write(data)
	sum := h.Sum(nil)
	if !bytes.Equal(sum, ref.Hash) {
		return nil, errors.Errorf("content failed hash check HAVE: %x WANT: %x", sum, ref.Hash)
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (r *Resolver) derefCipher(ctx context.Context, ref *Cipher, cache bool) (io.ReadCloser, error) {
	innerRc, err := r.deref(ctx, ref.Inner, cache)
	if err != nil {
		return nil, err
	}
	var rd io.Reader
	switch ref.Algo {
	case CipherAlgo_CHACHA20:
		ciph, err := chacha20.NewUnauthenticatedCipher(ref.Key, ref.Nonce)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		rd = &cipher.StreamReader{R: innerRc, S: ciph}
	default:
		return nil, errors.Errorf("unrecognized cipher algo %v", ref.Algo)
	}
	rc := readCloser{r: rd, closes: []func() error{
		innerRc.Close,
	}}
	return rc, nil
}

func (r *Resolver) derefCompress(ctx context.Context, ref *Compress, cache bool) (io.ReadCloser, error) {
	innerRc, err := r.deref(ctx, ref.Inner, cache)
	if err != nil {
		return nil, err
	}
	switch ref.Algo {
	case CompressAlgo_GZIP:
		gr, err := gzip.NewReader(innerRc)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		return readCloser{r: gr, closes: []func() error{
			innerRc.Close,
			gr.Close,
		}}, nil
	default:
		return nil, errors.Errorf("unrecognized compress algo %v", ref.Algo)
	}
}

func (r *Resolver) derefSlice(ctx context.Context, ref *Slice, cache bool) (io.ReadCloser, error) {
	innerRc, err := r.deref(ctx, ref.Inner, cache)
	if err != nil {
		return nil, err
	}
	if _, err := io.CopyN(io.Discard, innerRc, int64(ref.Start)); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return readCloser{
		r: io.LimitReader(innerRc, int64(ref.End-ref.Start)),
		closes: []func() error{
			innerRc.Close,
		},
	}, nil
}

func (r *Resolver) derefConcat(ctx context.Context, ref *Concat, cache bool) (io.ReadCloser, error) {
	rcs := make([]io.Reader, len(ref.Refs))
	closes := make([]func() error, len(ref.Refs))
	for i, ref := range ref.Refs {
		rc, err := r.deref(ctx, ref, cache)
		if err != nil {
			return nil, err
		}
		rcs[i] = rc
		closes[i] = rc.Close
	}
	return readCloser{r: io.MultiReader(rcs...), closes: closes}, nil
}

type readCloser struct {
	r      io.Reader
	closes []func() error
}

func (c readCloser) Read(buf []byte) (int, error) {
	n, err := c.r.Read(buf)
	return n, errors.EnsureStack(err)
}

func (c readCloser) Close() error {
	var retErr error
	for i := range c.closes {
		if err := c.closes[i](); err != nil {
			errors.JoinInto(&retErr, err)
		}
	}
	return retErr
}

type cache struct {
	cache *lru.Cache[[32]byte, []byte]
}

func newCache() cache {
	c, err := lru.New[[32]byte, []byte](8)
	if err != nil {
		panic(err)
	}
	return cache{
		cache: c,
	}
}

func (c cache) put(r *Ref, data []byte) {
	key := c.makeKey(r)
	data = append([]byte{}, data...) // clone
	c.cache.Add(key, data)
}

func (c cache) get(r *Ref) []byte {
	key := c.makeKey(r)
	v, ok := c.cache.Get(key)
	if !ok {
		return nil
	}
	return v
}

func (c cache) makeKey(r *Ref) [32]byte {
	data, err := proto.Marshal(r)
	if err != nil {
		panic(err)
	}
	return blake2b.Sum256(data)
}
