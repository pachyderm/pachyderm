package identityutil

import (
	"bytes"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"

	proto "github.com/gogo/protobuf/proto"
)

const (
	NoConfigErr = "no configs are defined"
)

// PickConfig determines whether to use the config or jsonConfig.
// JsonConfig should only be used for backwards compatibility.
func PickConfig(config *types.Struct, jsonConfig string) ([]byte, error) {
	switch {
	case config != nil && config.Fields != nil:
		marshaler := jsonpb.Marshaler{}
		buf := bytes.Buffer{}
		err := marshaler.Marshal(&buf, config)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		return buf.Bytes(), nil
	case jsonConfig != "" && jsonConfig != "null":
		return []byte(jsonConfig), nil
	default:
		return []byte(""), errors.New(NoConfigErr)
	}
}

// IDPConnectors satisfies the interface for proto.Message so they can be deserialized using jsonpb.Unmarshal().
type IDPConnectors []identity.IDPConnector

func (m *IDPConnectors) Reset()         { *m = IDPConnectors{} }
func (m *IDPConnectors) String() string { return proto.CompactTextString(m) }
func (*IDPConnectors) ProtoMessage()    {}
