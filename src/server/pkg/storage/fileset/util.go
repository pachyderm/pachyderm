package fileset

import "github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"

// CopyN copies n bytes from a file set reader to a file set writer.
// The bytes copied correspond to the current file being read / written respectively.
func CopyN(w *Writer, r *Reader, n int64) error {
	// Lazily setup reader for underlying file.
	if r.tr == nil {
		if err := r.setupReader(); err != nil {
			return err
		}
	}
	if err := chunk.CopyN(w.cw, r.cr, n); err != nil {
		return err
	}
	w.hdr.Idx.SizeBytes += n
	w.hdr.Idx.DataOp.Tags[len(w.hdr.Idx.DataOp.Tags)-1].SizeBytes += n
	if err := r.tr.Skip(n); err != nil {
		return err
	}
	return w.tw.Skip(n)
}
