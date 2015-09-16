package protolog

import (
	"bytes"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	"github.com/satori/go.uuid"
)

var (
	defaultIDAllocator  = &idAllocator{instanceID, 0}
	defaultTimer        = &timer{}
	defaultErrorHandler = &errorHandler{}
	defaultMarshaller   = &marshaller{}
	defaultUnmarshaller = &unmarshaller{}

	// go.uuid calls rand.Read, which gets down to a mutex
	// we just need ids to be unique across logging processes
	// so we use a base ID and then add an atomic int
	// it's not great but it's better than blocking on the mutex
	instanceID = uuid.NewV4().String()
)

type idAllocator struct {
	id    string
	value uint64
}

func (i *idAllocator) Allocate() string {
	return fmt.Sprintf("%s-%d", i.id, atomic.AddUint64(&i.value, 1))
}

type timer struct{}

func (t *timer) Now() time.Time {
	return time.Now().UTC()
}

type errorHandler struct{}

func (e *errorHandler) Handle(err error) {
	panic(err.Error())
}

type marshaller struct{}

func (m *marshaller) Marshal(entry *Entry) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	if _, err := pbutil.WriteDelimited(buffer, entry); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

type unmarshaller struct{}

func (u *unmarshaller) Unmarshal(reader io.Reader, entry *Entry) error {
	_, err := pbutil.ReadDelimited(reader, entry)
	return err
}
