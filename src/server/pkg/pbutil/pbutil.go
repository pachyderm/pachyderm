package pbutil

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
)

func Marshal(val proto.Marshaler) ([]byte, error) {
	bytes, err := val.Marshal()
	if err != nil {
		return nil, err
	}
	uncompressedLen := len(bytes)
	bytes = snappy.Encode(nil, bytes)
	compressedLen := len(bytes)
	fmt.Printf("Uncompressed: %d, Compressed: %d (%d%)\n", uncompressedLen, compressedLen, float64(uncompressedLen-compressedLen)*100.0/float64(uncompressedLen))
	return bytes, nil
}

func Unmarshal(val proto.Unmarshaler, bytes []byte) error {
	bytes, err := snappy.Decode(nil, bytes)
	if err != nil {
		return err
	}
	return val.Unmarshal(bytes)
}
