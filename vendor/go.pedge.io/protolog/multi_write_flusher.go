package protolog

import "io"

type multiWriteFlusher struct {
	writeFlushers []WriteFlusher
}

func newMultiWriteFlusher(writeFlushers []WriteFlusher) *multiWriteFlusher {
	w := make([]WriteFlusher, len(writeFlushers))
	copy(w, writeFlushers)
	return &multiWriteFlusher{writeFlushers}
}

func (m *multiWriteFlusher) Write(p []byte) (int, error) {
	for _, writeFlusher := range m.writeFlushers {
		n, err := writeFlusher.Write(p)
		if err != nil {
			return n, err
		}
		if n != len(p) {
			return n, io.ErrShortWrite
		}
	}
	return len(p), nil
}

func (m *multiWriteFlusher) Flush() error {
	var retErr error
	for _, writeFlusher := range m.writeFlushers {
		if err := writeFlusher.Flush(); err != nil {
			retErr = err
		}
	}
	return retErr
}
