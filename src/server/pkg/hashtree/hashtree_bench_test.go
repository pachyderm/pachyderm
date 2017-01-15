package hashtree

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"testing"

	"github.com/golang/protobuf/proto"
)

// var hashes int = 0

func BenchmarkBigDelete(b *testing.B) {
	h := HashTree{}
	r := rand.New(rand.NewSource(0))
	// Add 50k files
	for i := 0; i < 1e4; i++ {
		h.PutFile(fmt.Sprintf("/foo/shard-%05d", i),
			br(fmt.Sprintf(`block{hash:"%x"}`, r.Uint32())))
	}

	msg, _ := proto.Marshal(&h)
	// fmt.Println("\n(Pre)  Hashes: ", hashes)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h2 := HashTree{}
		proto.Unmarshal(msg, &h2)
		h2.DeleteDir("/foo")
	}
	// fmt.Printf("\n(Post %d) Hashes: %d\n", b.N, hashes)
}

func BenchmarkSha256(b *testing.B) {
	r := rand.New(rand.NewSource(0))
	buf := new(bytes.Buffer)
	for i := 0; i < 1e6; i++ {
		buf.WriteByte(byte(r.Uint32()))
	}

	fmt.Println(buf.Len())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sha256.Sum256(buf.Bytes())
	}
}
