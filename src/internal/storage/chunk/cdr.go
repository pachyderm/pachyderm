package chunk

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/cdr"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"gocloud.dev/blob"
	"golang.org/x/crypto/chacha20"
)

func (s *Storage) CDRFromDataRef(ctx context.Context, dataRef *DataRef) (*cdr.Ref, error) {
	client := NewClient(s.store, s.db, s.tracker, nil, s.pool).(*trackedClient)
	defer client.Close()
	path, err := client.GetPath(ctx, ID(dataRef.Ref.Id))
	if err != nil {
		return nil, err
	}
	url, err := s.bucket.SignedURL(ctx, path, &blob.SignedURLOptions{Expiry: 24 * time.Hour})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	ref := createHTTPRef(url, nil)
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
