package protoutil

import (
	"encoding/json"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ProtoJSONDecoder decodes a concatenated stream of protocol buffer messages encoded as JSON.
type ProtoJSONDecoder struct {
	opts protojson.UnmarshalOptions
	dec  *json.Decoder
}

func NewProtoJSONDecoder(r io.Reader, opts protojson.UnmarshalOptions) *ProtoJSONDecoder {
	return &ProtoJSONDecoder{
		opts: opts,
		dec:  json.NewDecoder(r),
	}
}

func (d *ProtoJSONDecoder) More() bool {
	return d.dec.More()
}

func (d *ProtoJSONDecoder) UnmarshalNext(dst proto.Message) error {
	var msg json.RawMessage
	if err := d.dec.Decode(&msg); err != nil {
		if err == io.EOF {
			return io.EOF
		}
		return errors.Wrap(err, "decode into json.RawMessage")
	}
	if err := d.opts.Unmarshal(msg, dst); err != nil {
		return errors.Wrap(err, "unmarshal protojson")
	}
	return nil
}
