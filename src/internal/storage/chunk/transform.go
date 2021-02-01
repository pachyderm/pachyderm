package chunk

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/cipher"
	io "io"
	"io/ioutil"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/chacha20"
)

// CreateOptions affect how chunks are created.
type CreateOptions struct {
	Secret      []byte
	Compression CompressionAlgo
}

// Create calls createFunc to create a new chunk, but first compresses, and encrypts ptext.
// ptext will not be modified.
func Create(ctx context.Context, opts CreateOptions, ptext []byte, createFunc func(ctx context.Context, data []byte) (ID, error)) (*Ref, error) {
	buf := make([]byte, len(ptext))
	compressAlgo, n, err := compress(opts.Compression, buf, ptext)
	if err != nil {
		return nil, err
	}
	buf = buf[:n]
	// encrypt in place; compress will always make a copy of the data.
	dek := encrypt(opts.Secret, buf, buf)
	id, err := createFunc(ctx, buf)
	if err != nil {
		return nil, err
	}
	return &Ref{
		Id:              id,
		SizeBytes:       int64(len(buf)),
		Dek:             dek,
		CompressionAlgo: compressAlgo,
	}, nil
}

// Get calls getfunc to retrieve a chunk, then verifies, decrypts, and decompresses the data.
// Uncompressed plaintext is written to w.
func Get(ctx context.Context, cache kv.GetPut, ref *Ref, w io.Writer, getFunc func(ctx context.Context, id ID, cb kv.ValueCallback) error) error {
	if err := getFromCache(ctx, cache, ref, w); err == nil {
		return nil
	}
	return getFunc(ctx, ref.Id, func(ctext []byte) error {
		if err := verifyData(ref.Id, ctext); err != nil {
			return err
		}
		var r io.Reader = bytes.NewReader(ctext)
		var err error
		if r, err = decrypt(ref.Dek, r); err != nil {
			return err
		}
		if r, err = decompress(ref.CompressionAlgo, r); err != nil {
			return err
		}
		rawData, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		if err := putInCache(ctx, cache, ref, rawData); err != nil {
			logrus.Error(err)
		}
		_, err = w.Write(rawData)
		return err
	})
}

// compress attempts to compress src using algo. If the compressed data is bigger
// then no compression is used.
// compress returns the compression algorithm used (algo or NONE), the number of bytes written to dst
// or an error
func compress(algo CompressionAlgo, dst, src []byte) (CompressionAlgo, int, error) {
	switch algo {
	case CompressionAlgo_NONE:
		copy(dst, src)
		return CompressionAlgo_NONE, len(src), nil
	case CompressionAlgo_GZIP_BEST_SPEED:
		lw := newLimitWriter(dst)
		err := func() (retErr error) {
			gw, err := gzip.NewWriterLevel(lw, gzip.BestSpeed)
			if err != nil {
				return err
			}
			defer func() {
				if err := gw.Close(); retErr == nil {
					retErr = err
				}
			}()
			_, err = gw.Write(src)
			if err != nil {
				return err
			}
			return gw.Close()
		}()
		if err == io.ErrShortWrite {
			return compress(CompressionAlgo_NONE, dst, src)
		}
		return CompressionAlgo_GZIP_BEST_SPEED, lw.pos, err
	default:
		return 0, 0, errors.Errorf("unrecognized compression: %v", algo)
	}
}

func decompress(algo CompressionAlgo, r io.Reader) (io.Reader, error) {
	switch algo {
	case CompressionAlgo_NONE:
		return r, nil
	case CompressionAlgo_GZIP_BEST_SPEED:
		gr, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		return gr, nil
	default:
		return nil, errors.Errorf("unrecognized compression: %v", algo)
	}
}

type limitWriter struct {
	buf []byte
	pos int
}

func newLimitWriter(buf []byte) *limitWriter {
	return &limitWriter{buf: buf}
}

func (w *limitWriter) Write(p []byte) (int, error) {
	if w.pos >= len(w.buf) {
		return 0, io.ErrShortWrite
	}
	n := copy(w.buf[w.pos:], p)
	if n < len(p) {
		return n, io.ErrShortWrite
	}
	w.pos += n
	return n, nil
}

// encrypt generates a key using secret and src, and encrypts src, writing the output to dst
func encrypt(secret []byte, dst, src []byte) (dek []byte) {
	dek = deriveKey(secret, src)
	cryptoXOR(dek[:32], dst, src)
	return dek
}

// decrypt returns an io.Reader containing r decrypted using dek
func decrypt(dek []byte, r io.Reader) (io.Reader, error) {
	if len(dek) != 32 {
		return nil, errors.Errorf("data encryption key is wrong length")
	}
	nonce := [chacha20.NonceSize]byte{}
	ciph, err := chacha20.NewUnauthenticatedCipher(dek, nonce[:])
	if err != nil {
		return nil, err
	}
	return cipher.StreamReader{S: ciph, R: r}, nil
}

// deriveKey returns Hash(secret + Hash(ptext))
func deriveKey(secret, ptext []byte) []byte {
	var x []byte
	x = append(x, secret...)
	x = append(x, Hash(ptext)...)
	return Hash(x)[:32]
}

// cryptoXOR setups up a stream cipher using key, and writes (src XOR keystream) to dst
func cryptoXOR(key, dst, src []byte) {
	nonce := [chacha20.NonceSize]byte{}
	ciph, err := chacha20.NewUnauthenticatedCipher(key, nonce[:])
	if err != nil {
		panic(err) // this only happens if you pass in an invalid key/nonce size, which we shouldn't do
	}
	ciph.XORKeyStream(dst, src)
}

func verifyData(id ID, x []byte) error {
	actualHash := Hash(x)
	if !bytes.Equal(actualHash[:], id) {
		return errors.Errorf("bad chunk. HAVE: %x WANT: %x", actualHash, id)
	}
	return nil
}

// Key returns a unique key for the Ref suitable for use in hash tables
func (r *Ref) Key() pachhash.Output {
	data, err := r.Marshal()
	if err != nil {
		panic(err)
	}
	return pachhash.Sum(data)
}

func getFromCache(ctx context.Context, cache kv.GetPut, ref *Ref, w io.Writer) error {
	key := ref.Key()
	return cache.GetF(ctx, key[:], func(value []byte) error {
		_, err := w.Write(value)
		return err
	})
}

func putInCache(ctx context.Context, cache kv.GetPut, ref *Ref, data []byte) error {
	key := ref.Key()
	return cache.Put(ctx, key[:], data)
}
