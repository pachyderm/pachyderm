package protolog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"

	"go.pedge.io/proto/time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

var (
	jsonPBMarshaller = &jsonpb.Marshaler{}
)

type textMarshaller struct {
	options MarshallerOptions
}

func newTextMarshaller(options MarshallerOptions) *textMarshaller {
	return &textMarshaller{options}
}

func (t *textMarshaller) Marshal(entry *Entry) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	if t.options.EnableID {
		_, _ = buffer.WriteString(entry.Id)
		_ = buffer.WriteByte(' ')
	}
	if !t.options.DisableTimestamp {
		stdTime := prototime.TimestampToTime(entry.Timestamp)
		_, _ = buffer.WriteString(stdTime.Format(time.RFC3339))
		_ = buffer.WriteByte(' ')
	}
	if !t.options.DisableLevel {
		levelString := strings.Replace(entry.Level.String(), "LEVEL_", "", -1)
		_, _ = buffer.WriteString(levelString)
		if len(levelString) == 4 {
			_, _ = buffer.WriteString("  ")
		} else {
			_ = buffer.WriteByte(' ')
		}
	}
	event, err := entry.UnmarshalledEvent()
	if err != nil {
		return nil, err
	}
	name := messageName(event)
	switch name {
	case "protolog.Event":
		protologEvent, ok := event.(*Event)
		if !ok {
			return nil, fmt.Errorf("protolog: expected *protolog.Event, got %T", event)
		}
		_, _ = buffer.WriteString(protologEvent.Message)
	case "protolog.WriterOutput":
		writerOutput, ok := event.(*WriterOutput)
		if !ok {
			return nil, fmt.Errorf("protolog: expected *protolog.WriterOutput, got %T", event)
		}
		_, _ = buffer.Write(trimRightSpaceBytes(writerOutput.Value))
	default:
		if err := t.marshalMessage(buffer, event); err != nil {
			return nil, err
		}
	}
	if entry.Context != nil && len(entry.Context) > 0 && !t.options.DisableContexts {
		_, _ = buffer.WriteString(" contexts=[")
		contexts, err := entry.UnmarshalledContexts()
		if err != nil {
			return nil, err
		}
		lenContexts := len(contexts)
		for i, context := range contexts {
			name := messageName(context)
			switch name {
			case "protolog.Fields":
				protologFields, ok := context.(*Fields)
				if !ok {
					return nil, fmt.Errorf("protolog: expected *protolog.Fields, got %T", context)
				}
				data, err := json.Marshal(protologFields.Value)
				if err != nil {
					return nil, err
				}
				_, _ = buffer.Write(data)
			default:
				if err := t.marshalMessage(buffer, context); err != nil {
					return nil, err
				}
			}
			if i != lenContexts-1 {
				_, _ = buffer.WriteString(", ")
			}
		}
		_ = buffer.WriteByte(']')
	}
	return trimRightSpaceBytes(buffer.Bytes()), nil
}

func (t *textMarshaller) marshalMessage(buffer *bytes.Buffer, message proto.Message) error {
	s, err := jsonPBMarshaller.MarshalToString(message)
	if err != nil {
		return err
	}
	_, _ = buffer.WriteString(messageName(message))
	_ = buffer.WriteByte(' ')
	_, _ = buffer.WriteString(s)
	return nil
}

func trimRightSpaceBytes(b []byte) []byte {
	return bytes.TrimRightFunc(b, unicode.IsSpace)
}
