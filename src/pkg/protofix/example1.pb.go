package protofix

type Commit struct {
	Id   string `protobuf:"bytes,2,opt,name=id" json:"id,omitempty"`
}
