package chunk

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/cipher"
	"encoding/hex"
	io "io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"golang.org/x/crypto/chacha20"
)

type CreateOptions struct {
	Secret      []byte
	Compression CompressionAlgo
}

func Create(ctx context.Context, opts CreateOptions, data []byte, createFunc func(ctx context.Context, data []byte) (ID, error)) (*Ref, error) {
	buf := make([]byte, len(data))
	// compression
	compressAlgo, n, err := compress(opts.Compression, buf, data)
	if err != nil {
		return nil, err
	}
	buf = buf[:n]
	// encryption
	dek := encrypt(opts.Secret, buf, buf)
	// upload
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

func Get(ctx context.Context, memCache obj.Client, ref *Ref, w io.Writer, getFunc func(ctx context.Context, id ID, w io.Writer) error) error {
	// check cache
	if err := getFromCache(ctx, memCache, ref, w); err == nil {
		return nil
	}
	buf := &bytes.Buffer{}
	if err := getFunc(ctx, ref.Id, buf); err != nil {
		return err
	}
	// validation
	actualHash := Hash(buf.Bytes())
	if !bytes.Equal(actualHash, ref.Id) {
		return errors.Errorf("bad chunk. HAVE: %x WANT: %x", actualHash, ref.Id)
	}
	// decryption
	var r io.Reader = buf
	var err error
	if r, err = decrypt(ref.Dek, r); err != nil {
		return err
	}
	// decompression
	if r, err = decompress(ref.CompressionAlgo, r); err != nil {
		return err
	}
	// write to cache
	if cw, err := cacheWriter(ctx, memCache, ref); err == nil {
		w = io.MultiWriter(w, cw)
	}
	_, err = io.Copy(w, r)
	return err
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
		err := func() error {
			gw, err := gzip.NewWriterLevel(lw, gzip.BestSpeed)
			if err != nil {
				return err
			}
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

func (r *Ref) Key() string {
	data, err := r.Marshal()
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(Hash(data)[:16])
}

func getFromCache(ctx context.Context, objs obj.Client, ref *Ref, w io.Writer) error {
	rc, err := objs.Reader(ctx, ref.Key(), 0, 0)
	if err != nil {
		return err
	}
	defer rc.Close()
	_, err = io.Copy(w, rc)
	return err
}

func cacheWriter(ctx context.Context, objs obj.Client, ref *Ref) (io.WriteCloser, error) {
	return objs.Writer(ctx, ref.Key())
}
