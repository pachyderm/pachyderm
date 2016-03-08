package protofix

import(
	"testing"
)

func TestSimple(t *testing.T) {
	repairFile("example1.pb.go")
	
}
