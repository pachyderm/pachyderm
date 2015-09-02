package protolog

import (
	"encoding/binary"
	"fmt"
)

var (
	wireEncoderInstance = &wireEncoder{}
)

type wireEncoder struct{}

func (w *wireEncoder) Encode(data []byte) ([]byte, error) {
	lenData := len(data)
	if lenData == 0 {
		return data, nil
	}
	lenBuf := lenData + 9
	buf := make([]byte, lenBuf)
	binary.LittleEndian.PutUint64(buf, uint64(lenData))
	written := copy(buf[8:], data)
	if written != lenData {
		return nil, fmt.Errorf("protolog: tried to write %d bytes to buffer, write %d bytes", lenData, written)
	}
	// docker logs do not flush without a newline, this is the hack
	buf[lenBuf-1] = '\n'
	return buf, nil
}
