package lion

import (
	"bytes"
	"time"
)

type jsonMarshaller struct {
	disableNewlines bool
}

func newJSONMarshaller(options ...JSONMarshallerOption) *jsonMarshaller {
	jsonMarshaller := &jsonMarshaller{false}
	for _, option := range options {
		option(jsonMarshaller)
	}
	return jsonMarshaller
}

func (t *jsonMarshaller) Marshal(entry *Entry) ([]byte, error) {
	return jsonMarshalEntry(entry, t.disableNewlines)
}

func jsonMarshalEntry(entry *Entry, disableNewlines bool) ([]byte, error) {
	jsonEntry, err := entryToJSONEntry(entry)
	if err != nil {
		return nil, err
	}
	if jsonEntry == nil {
		return nil, nil
	}
	buffer := bytes.NewBuffer(nil)
	if err := globalJSONMarshalFunc(buffer, jsonEntry); err != nil {
		return nil, err
	}
	if !disableNewlines {
		_ = buffer.WriteByte('\n')
	}
	return buffer.Bytes(), nil
}

type jsonEntry struct {
	ID           string            `json:"id,omitempty"`
	Level        string            `json:"level,omitempty"`
	Time         string            `json:"time,omitempty"`
	Contexts     []interface{}     `json:"contexts,omitempty"`
	Fields       map[string]string `json:"fields,omitempty"`
	Event        interface{}       `json:"event,omitempty"`
	Message      string            `json:"message,omitempty"`
	WriterOutput string            `json:"writer_output,omitempty"`
}

type jsonEntryMessage struct {
	Name  string      `json:"name,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

func entryToJSONEntry(entry *Entry) (*jsonEntry, error) {
	if entry == nil {
		return nil, nil
	}
	jsonEntry := &jsonEntry{
		ID:           entry.ID,
		Level:        entry.Level.String(),
		Time:         entry.Time.Format(time.RFC3339),
		Fields:       entry.Fields,
		Message:      entry.Message,
		WriterOutput: string(entry.WriterOutput),
	}
	if len(entry.Contexts) > 0 {
		jsonEntry.Contexts = make([]interface{}, 0)
		for _, context := range entry.Contexts {
			jsonContext, err := entryMessageToJSONEntryMessage(context)
			if err != nil {
				return nil, err
			}
			if jsonContext != nil {
				jsonEntry.Contexts = append(jsonEntry.Contexts, jsonContext)
			}
		}
	}
	if entry.Event != nil {
		jsonEvent, err := entryMessageToJSONEntryMessage(entry.Event)
		if err != nil {
			return nil, err
		}
		jsonEntry.Event = jsonEvent
	}
	return jsonEntry, nil
}

func entryMessageToJSONEntryMessage(entryMessage *EntryMessage) (*jsonEntryMessage, error) {
	if entryMessage == nil {
		return nil, nil
	}
	name, err := entryMessage.Name()
	if err != nil {
		return nil, nil
	}
	return &jsonEntryMessage{
		Name:  name,
		Value: entryMessage.Value,
	}, nil
}
