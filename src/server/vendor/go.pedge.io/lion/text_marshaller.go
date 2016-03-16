package lion

import (
	"bytes"
	"time"
	"unicode"

	"github.com/fatih/color"
)

var (
	levelToColorString = map[Level]string{
		LevelNone:  color.BlueString(LevelNone.String()),
		LevelDebug: color.WhiteString(LevelDebug.String()),
		LevelInfo:  color.BlueString(LevelInfo.String()),
		LevelWarn:  color.YellowString(LevelWarn.String()),
		LevelError: color.RedString(LevelError.String()),
		LevelFatal: color.RedString(LevelFatal.String()),
		LevelPanic: color.RedString(LevelPanic.String()),
	}
	// TODO(pedge): clean up
	fourLetterLevels = map[Level]bool{
		LevelNone: true,
		LevelInfo: true,
		LevelWarn: true,
	}
)

type textMarshaller struct {
	disableTime     bool
	disableLevel    bool
	disableContexts bool
	disableNewlines bool
	colorize        bool
}

func newTextMarshaller(options ...TextMarshallerOption) *textMarshaller {
	textMarshaller := &textMarshaller{
		false,
		false,
		false,
		false,
		false,
	}
	for _, option := range options {
		option(textMarshaller)
	}
	return textMarshaller
}

func (t *textMarshaller) WithColors() TextMarshaller {
	return &textMarshaller{
		t.disableTime,
		t.disableLevel,
		t.disableContexts,
		t.disableNewlines,
		true,
	}
}

func (t *textMarshaller) WithoutColors() TextMarshaller {
	return &textMarshaller{
		t.disableTime,
		t.disableLevel,
		t.disableContexts,
		t.disableNewlines,
		false,
	}
}

func (t *textMarshaller) Marshal(entry *Entry) ([]byte, error) {
	return textMarshalEntry(
		entry,
		t.disableTime,
		t.disableLevel,
		t.disableContexts,
		t.disableNewlines,
		t.colorize,
	)
}

func textMarshalEntry(
	entry *Entry,
	disableTime bool,
	disableLevel bool,
	disableContexts bool,
	disableNewlines bool,
	colorize bool,
) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	if entry.ID != "" {
		_, _ = buffer.WriteString(entry.ID)
		_ = buffer.WriteByte(' ')
	}
	if !disableTime {
		_, _ = buffer.WriteString(entry.Time.Format(time.RFC3339))
		_ = buffer.WriteByte(' ')
	}
	if !disableLevel {
		var levelString string
		if colorize {
			levelString = levelToColorString[entry.Level]
		} else {
			levelString = entry.Level.String()
		}
		_, _ = buffer.WriteString(levelString)
		_ = buffer.WriteByte(' ')
		if _, ok := fourLetterLevels[entry.Level]; ok {
			_ = buffer.WriteByte(' ')
		}
	}
	eventSeen := false
	// TODO(pedge): verify only one of Event, Message, WriterOutput?
	if entry.Event != nil {
		eventSeen = true
		if err := textMarshalMessage(buffer, entry.Event); err != nil {
			return nil, err
		}
	}
	if entry.Message != "" {
		eventSeen = true
		_, _ = buffer.WriteString(entry.Message)
	}
	if entry.WriterOutput != nil {
		eventSeen = true
		_, _ = buffer.Write(trimRightSpaceBytes(entry.WriterOutput))
	}
	if len(entry.Contexts) > 0 && !disableContexts {
		if eventSeen {
			_ = buffer.WriteByte(' ')
		}
		eventSeen = true
		lenContexts := len(entry.Contexts)
		for i, context := range entry.Contexts {
			if err := textMarshalMessage(buffer, context); err != nil {
				return nil, err
			}
			if i != lenContexts-1 {
				_ = buffer.WriteByte(' ')
			}
		}
	}
	if len(entry.Fields) > 0 && !disableContexts {
		if eventSeen {
			_ = buffer.WriteByte(' ')
		}
		if err := globalJSONMarshalFunc(buffer, entry.Fields); err != nil {
			return nil, err
		}
	}
	data := trimRightSpaceBytes(buffer.Bytes())
	if !disableNewlines {
		buffer = bytes.NewBuffer(data)
		_ = buffer.WriteByte('\n')
		return buffer.Bytes(), nil
	}
	return data, nil
}

func textMarshalMessage(buffer *bytes.Buffer, message *EntryMessage) error {
	if message == nil {
		return nil
	}
	name, err := message.Name()
	if err != nil {
		return err
	}
	_, _ = buffer.WriteString(name)
	_ = buffer.WriteByte(' ')
	return globalJSONMarshalFunc(buffer, message.Value)
}

func trimRightSpaceBytes(b []byte) []byte {
	return bytes.TrimRightFunc(b, unicode.IsSpace)
}
