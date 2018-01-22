package pbutil

import (
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
)

// Marshal marshals val, and compresses the serialized data.
func Marshal(val proto.Marshaler) ([]byte, error) {
	bytes, err := val.Marshal()
	if err != nil {
		return nil, err
	}
	bytes = snappy.Encode(nil, bytes)
	return bytes, nil
}

// Unmarshal decompresses bytes and unmarshals it into val.
func Unmarshal(val proto.Unmarshaler, bytes []byte) error {
	bytes, err := snappy.Decode(nil, bytes)
	if err != nil {
		return err
	}
	return val.Unmarshal(bytes)
}
