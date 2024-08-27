package chunk

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pachyderm/pachyderm/v2/src/cdr"
	"gocloud.dev/blob"
	"golang.org/x/crypto/chacha20"
)

func (s *Storage) CDRFromDataRef(ctx context.Context, dataRef *DataRef, cache *lru.Cache[string, string]) (_ *cdr.Ref, retErr error) {
	client := NewClient(s.store, s.db, s.tracker, nil, s.pool).(*trackedClient)
	defer errors.Close(&retErr, client, "close cdr client")
	key := string(dataRef.Ref.Id)
	url, isCached := cache.Get(key)
	if !isCached {
		path, err := client.GetPath(ctx, ID(dataRef.Ref.Id))
		if err != nil {
			return nil, err
		}
		u, err := s.bucket.SignedURL(ctx, path, &blob.SignedURLOptions{Expiry: 24 * time.Hour})
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		url = u
		cache.Add(key, url)
	}
	ref := createHTTPRef(url, nil)
	ref = createContentHashRef(ref, dataRef.Ref.Id)
	ref = createCipherRef(ref, dataRef.Ref.EncryptionAlgo, dataRef.Ref.Dek)
	ref = createCompressRef(ref, dataRef.Ref.CompressionAlgo)
	ref = createSliceRef(ref, uint64(dataRef.OffsetBytes), uint64(dataRef.OffsetBytes+dataRef.SizeBytes))
	return ref, nil
}

func createHTTPRef(url string, headers map[string]string) *cdr.Ref {
	return &cdr.Ref{
		Body: &cdr.Ref_Http{Http: &cdr.HTTP{
			Url:     url,
			Headers: headers,
		}},
	}
}

func createContentHashRef(ref *cdr.Ref, hash []byte) *cdr.Ref {
	return &cdr.Ref{
		Body: &cdr.Ref_ContentHash{
			ContentHash: &cdr.ContentHash{
				Inner: ref,
				Algo:  cdr.HashAlgo_BLAKE2b_256,
				Hash:  hash,
			},
		},
	}
}

func createCipherRef(ref *cdr.Ref, algo EncryptionAlgo, dek []byte) *cdr.Ref {
	cipher := &cdr.Cipher{
		Inner: ref,
		Key:   dek,
		Nonce: make([]byte, chacha20.NonceSize),
	}
	switch algo {
	case EncryptionAlgo_ENCRYPTION_ALGO_UNKNOWN:
		return ref
	case EncryptionAlgo_CHACHA20:
		cipher.Algo = cdr.CipherAlgo_CHACHA20
	}
	return &cdr.Ref{Body: &cdr.Ref_Cipher{Cipher: cipher}}
}

func createCompressRef(ref *cdr.Ref, algo CompressionAlgo) *cdr.Ref {
	compress := &cdr.Compress{Inner: ref}
	switch algo {
	case CompressionAlgo_NONE:
		return ref
	case CompressionAlgo_GZIP_BEST_SPEED:
		compress.Algo = cdr.CompressAlgo_GZIP
	}
	return &cdr.Ref{Body: &cdr.Ref_Compress{Compress: compress}}
}

func createSliceRef(ref *cdr.Ref, start, end uint64) *cdr.Ref {
	if end <= start {
		panic("end must be > start")
	}
	return &cdr.Ref{
		Body: &cdr.Ref_Slice{Slice: &cdr.Slice{
			Inner: ref,
			Start: start,
			End:   end,
		}},
	}
}

func CreateConcatRef(refs []*cdr.Ref) *cdr.Ref {
	return &cdr.Ref{
		Body: &cdr.Ref_Concat{Concat: &cdr.Concat{Refs: refs}},
	}
}
