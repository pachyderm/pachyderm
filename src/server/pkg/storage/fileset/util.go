package fileset

import "github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"

// CopyN copies n bytes from a file set reader to a file set writer.
// The bytes copied correspond to the current file being read / written respectively.
func CopyN(w *Writer, r *Reader, n int64) error {
	if err := chunk.CopyN(w.cw, r.cr, n); err != nil {
		return err
	}
	if err := r.tr.Skip(n); err != nil {
		return err
	}
	return w.tw.Skip(n)
}
