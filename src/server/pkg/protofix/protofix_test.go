package protofix

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestSimple(t *testing.T) {
	r := repairedFileBytes("example1_pb_go")

	expected, _ := ioutil.ReadFile("example1_output_pb_go")

	if !bytes.Equal(r, expected) {
		t.Errorf("Protobuf go code not being normalized properly.\n======== Expected:\n%v\n======== Got:\n%v\n", string(expected), string(r))
	}

}
