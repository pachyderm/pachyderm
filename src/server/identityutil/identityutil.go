package identityutil

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

const (
	NoConfigErr = "no configs are defined"
)

// PickConfig determines whether to use the config or jsonConfig.
// JsonConfig should only be used for backwards compatibility.
func PickConfig(config *structpb.Struct, jsonConfig string) ([]byte, error) {
	switch {
	case config != nil && config.Fields != nil:
		marshaler := protojson.MarshalOptions{}
		result, err := marshaler.Marshal(config)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		return result, nil
	case jsonConfig != "" && jsonConfig != "null":
		return []byte(jsonConfig), nil
	default:
		return []byte(""), errors.New(NoConfigErr)
	}
}

// IDPConnectors satisfies the interface for proto.Message so they can be deserialized using jsonpb.Unmarshal().
type IDPConnectors []*identity.IDPConnector

func (out *IDPConnectors) UnmarshalJSON(in []byte) error {
	var parts []json.RawMessage
	if err := json.Unmarshal(in, &parts); err != nil {
		return errors.Wrap(err, "unmarshal into slice of JSON messages")
	}
	var errs error
	for i, part := range parts {
		var msg identity.IDPConnector
		if err := protojson.Unmarshal(part, &msg); err != nil {
			errors.JoinInto(&errs, errors.Wrapf(err, "unmarshal connector %d/%d", i+1, len(parts)))
			continue
		}
		*out = append(*out, &msg)
	}
	return errs
}
