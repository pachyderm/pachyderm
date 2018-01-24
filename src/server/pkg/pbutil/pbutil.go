package pbutil

import (
	"github.com/gogo/protobuf/proto"
)

// Convert converts on proto to another by marhsaling and unmarshaling it.
func Convert(in proto.Marshaler, out proto.Unmarshaler) error {
	bytes, err := in.Marshal()
	if err != nil {
		return err
	}
	return out.Unmarshal(bytes)
}
