package protolog

import (
	"encoding/binary"
	"fmt"
	"io"
)

var (
	wireDecoderInstance = &wireDecoder{}
)

type wireDecoder struct{}

func (w *wireDecoder) Decode(reader io.Reader) ([]byte, error) {
	sizeBuf, err := w.read(reader, 8)
	if err != nil {
		return nil, err
	}
	size := binary.LittleEndian.Uint64(sizeBuf)
	data, err := w.read(reader, int(size))
	if err != nil {
		return nil, err
	}
	// docker logs do not flush without a newline, this is the hack
	separatorBuf, err := w.read(reader, 1)
	if err != nil {
		return nil, err
	}
	if separatorBuf[0] != '\n' {
		return nil, fmt.Errorf("protolog: tried to read newline, but was not newline, size was %d", size)
	}
	return data, nil
}

func (w *wireDecoder) read(reader io.Reader, size int) ([]byte, error) {
	buffer := make([]byte, size)
	readSoFar := 0
	for readSoFar < size {
		data := make([]byte, size-readSoFar)
		n, err := reader.Read(data)
		if err != nil {
			return nil, err
		}
		if n > 0 {
			m := copy(buffer[readSoFar:readSoFar+n], data[:n])
			if m != n {
				return nil, fmt.Errorf("protolog: tried to copy %d bytes, only copied %d bytes", n, m)
			}
			readSoFar += n
		}
	}
	return buffer, nil
}
