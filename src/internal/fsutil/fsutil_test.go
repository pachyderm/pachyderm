package fsutil

import (
	"os"
	"testing"
)

func TestFind(t *testing.T) {
	infos, err := Find(os.DirFS("."))
	if err != nil {
		t.Fatal(err)
	}
	for _, info := range infos {
		t.Logf("%s", info)
	}
}
