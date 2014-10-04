package btrfs

import (
	"bufio"
	"io"
	"testing"
)

func TestCreateDestroy(t *testing.T) {
	fs := FS{"test", "loop2", "test", 10}
	err := fs.Format()
	if err != nil { t.Fatal(err) }
    err = Sync()
	if err != nil { t.Fatal(err) }
	err = fs.Destroy()
	if err != nil { t.Fatal(err) }
    err = Sync()
	if err != nil { t.Fatal(err) }
}

func TestCreateFile(t *testing.T) {
	fs := FS{"test", "loop2", "test", 10}
	err := fs.Format()
	if err != nil { t.Fatal(err) }
	defer func() {
		err = fs.Destroy()
		if err != nil { t.Fatal(err) }
	}()

	f, err := fs.Create("foo")
	if err != nil { t.Fatal(err) }
	f.WriteString("foo\n")
	f.Close()

    err = Sync()
	if err != nil { t.Fatal(err) }

	err = fs.Unmount()
	if err != nil { t.Fatal(err) }
	err = fs.Mount()
	if err != nil { t.Fatal(err) }

	f, err = fs.Open("foo")
	if err != nil { t.Fatal(err) }
	reader := bufio.NewReader(f)
	line, err := reader.ReadString('\n')
    if err != nil { t.Fatal(err) }
	if line != "foo\n" { t.Fatal("File contained the wrong value.") }
	f.Close()

	fs.Remove("foo")
}

func TestSendRecv(t *testing.T) {
	fs1 := FS{"test1", "loop2", "test1", 10}
	fs2 := FS{"test2", "loop3", "test2", 10}
	err := fs1.Format()
	if err != nil {
		t.Fatal(err)
	}
	err = fs2.Format()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = fs1.Destroy()
		if err != nil {
			t.Fatal(err)
		}
		err = fs2.Destroy()
		if err != nil {
			t.Fatal(err)
		}
	}()

	{
		fs1.SubvolumeCreate("vol")
		fs2.SubvolumeCreate("vol")
	}

	{
		err = fs1.Snapshot("vol", "vol/1", true)
		if err != nil {
			t.Fatal(err)
		}
		err = fs1.SendBase("vol/1",
			func(reader io.ReadCloser) error { return fs2.Recv("vol", reader) })
		if err != nil {
			t.Fatal(err)
		}
	}

	{
		f, err := fs1.Create("vol/foo")
		if err != nil {
			t.Fatal(err)
		}
		f.WriteString("foo\n")
		f.Close()
		fs1.Snapshot("vol", "vol/2", true)
		err = fs1.Send("vol/1", "vol/2",
			func(reader io.ReadCloser) error { return fs2.Recv("vol", reader) })
		if err != nil {
			t.Fatal(err)
		}
	}

    err = Sync()
	if err != nil { t.Fatal(err) }

	{
		f, err := fs2.Open("vol/2/foo")
		if err != nil {
			t.Fatal(err)
		}
		reader := bufio.NewReader(f)
		line, err := reader.ReadString('\n')
		if line != "foo\n" {
			t.Fatal(err)
		}
		f.Close()
	}
}
