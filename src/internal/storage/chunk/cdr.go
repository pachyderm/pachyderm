package chunk

import (
	"context"
	"time"

	"github.com/pachyderm/cdr"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
)

func (s *Storage) CDRFromDataRef(ctx context.Context, dr *DataRef) (*cdr.Ref, error) {
	id := ID(dr.Ref.Id)
	client := NewClient(s.store, s.db, s.tracker, nil).(*trackedClient)
	defer client.Close() // This shouldn't matter since we didn't pass a renewer
	paths, err := client.LookupPaths(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(paths) < 1 {
		return nil, errors.Errorf("no objects for chunk %v", id)
	}
	uSigner, ok := s.objClient.(obj.URLSigner)
	if !ok {
		return nil, errors.Errorf("object store %v does not support signed urls", s.objClient.BucketURL().Scheme)
	}
	u, err := uSigner.CreateSignedURL(ctx, paths[0], 24*time.Hour)
	if err != nil {
		return nil, err
	}
	ret := cdr.CreateHTTPRef(u, nil)
	ret = createCipherRef(ret, dr.Ref.Dek)
	ret = createCompressRef(ret, dr.Ref.CompressionAlgo)
	ret = cdr.CreateSliceRef(ret, uint64(dr.OffsetBytes), uint64(dr.OffsetBytes+dr.SizeBytes))
	return ret, nil
}

func createCipherRef(x *cdr.Ref, dek []byte) *cdr.Ref {
	return &cdr.Ref{
		Body: &cdr.Ref_Cipher{Cipher: &cdr.Cipher{
			Algo:  cdr.CipherAlgo_CHACHA20,
			Key:   dek,
			Inner: x,
		}},
	}
}

func createCompressRef(x *cdr.Ref, algo CompressionAlgo) *cdr.Ref {
	var algo2 cdr.CompressAlgo
	switch algo {
	case CompressionAlgo_GZIP_BEST_SPEED:
		algo2 = cdr.CompressAlgo_GZIP
	}
	return &cdr.Ref{
		Body: &cdr.Ref_Compress{Compress: &cdr.Compress{
			Algo:  algo2,
			Inner: x,
		}},
	}
}
