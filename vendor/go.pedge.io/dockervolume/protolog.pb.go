package dockervolume

import proto "github.com/golang/protobuf/proto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal

type Method int32

const (
	Method_METHOD_NONE    Method = 0
	Method_METHOD_CREATE  Method = 1
	Method_METHOD_REMOVE  Method = 2
	Method_METHOD_PATH    Method = 3
	Method_METHOD_MOUNT   Method = 4
	Method_METHOD_UNMOUNT Method = 5
)

var Method_name = map[int32]string{
	0: "METHOD_NONE",
	1: "METHOD_CREATE",
	2: "METHOD_REMOVE",
	3: "METHOD_PATH",
	4: "METHOD_MOUNT",
	5: "METHOD_UNMOUNT",
}
var Method_value = map[string]int32{
	"METHOD_NONE":    0,
	"METHOD_CREATE":  1,
	"METHOD_REMOVE":  2,
	"METHOD_PATH":    3,
	"METHOD_MOUNT":   4,
	"METHOD_UNMOUNT": 5,
}

func (x Method) String() string {
	return proto.EnumName(Method_name, int32(x))
}

type MethodInvocation struct {
	Method     Method            `protobuf:"varint,1,opt,name=method,enum=dockervolume.Method" json:"method,omitempty"`
	Name       string            `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	Opts       map[string]string `protobuf:"bytes,3,rep,name=opts" json:"opts,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Mountpoint string            `protobuf:"bytes,4,opt,name=mountpoint" json:"mountpoint,omitempty"`
	Error      string            `protobuf:"bytes,5,opt,name=error" json:"error,omitempty"`
}

func (m *MethodInvocation) Reset()         { *m = MethodInvocation{} }
func (m *MethodInvocation) String() string { return proto.CompactTextString(m) }
func (*MethodInvocation) ProtoMessage()    {}

func (m *MethodInvocation) GetOpts() map[string]string {
	if m != nil {
		return m.Opts
	}
	return nil
}

func init() {
	proto.RegisterEnum("dockervolume.Method", Method_name, Method_value)
}
