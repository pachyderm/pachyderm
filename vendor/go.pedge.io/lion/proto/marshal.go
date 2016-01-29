package protolion

import (
	"bytes"
	"fmt"
	"io"

	"go.pedge.io/lion"
	"go.pedge.io/pb/go/google/protobuf"
)

var (
	levelToProto = map[lion.Level]Level{
		lion.LevelDebug: Level_LEVEL_DEBUG,
		lion.LevelInfo:  Level_LEVEL_INFO,
		lion.LevelWarn:  Level_LEVEL_WARN,
		lion.LevelError: Level_LEVEL_ERROR,
		lion.LevelFatal: Level_LEVEL_FATAL,
		lion.LevelPanic: Level_LEVEL_PANIC,
		lion.LevelNone:  Level_LEVEL_NONE,
	}
	protoToLevel = map[Level]lion.Level{
		Level_LEVEL_DEBUG: lion.LevelDebug,
		Level_LEVEL_INFO:  lion.LevelInfo,
		Level_LEVEL_WARN:  lion.LevelWarn,
		Level_LEVEL_ERROR: lion.LevelError,
		Level_LEVEL_FATAL: lion.LevelFatal,
		Level_LEVEL_PANIC: lion.LevelPanic,
		Level_LEVEL_NONE:  lion.LevelNone,
	}
)

type delimitedMarshaller struct {
	base64Encode bool
	newline      bool
}

func newDelimitedMarshaller(base64Encode bool, newline bool) *delimitedMarshaller {
	return &delimitedMarshaller{base64Encode, newline}
}

func (m *delimitedMarshaller) Marshal(entry *lion.Entry) ([]byte, error) {
	encodedEntry, err := entry.Encode()
	if err != nil {
		return nil, err
	}
	protoEntry, err := encodedEntryToProtoEntry(encodedEntry)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(nil)
	if _, err := writeDelimited(buffer, protoEntry, m.base64Encode, m.newline); err != nil {
		return nil, err
	}
	data := buffer.Bytes()
	return data, nil
}

type delimitedUnmarshaller struct {
	base64Decode bool
	newline      bool
}

func newDelimitedUnmarshaller(base64Decode bool, newline bool) *delimitedUnmarshaller {
	return &delimitedUnmarshaller{base64Decode, newline}
}

func (u *delimitedUnmarshaller) Unmarshal(reader io.Reader, encodedEntry *lion.EncodedEntry) error {
	protoEntry := &Entry{}
	if _, err := readDelimited(reader, protoEntry, u.base64Decode, u.newline); err != nil {
		return err
	}
	iEntry, err := protoEntryToEncodedEntry(protoEntry)
	if err != nil {
		return err
	}
	*encodedEntry = *iEntry
	return nil
}

func encodedEntryToProtoEntry(encodedEntry *lion.EncodedEntry) (*Entry, error) {
	contexts, err := messagesToEntryMessages(encodedEntry.Contexts)
	if err != nil {
		return nil, err
	}
	event, err := messageToEntryMessage(encodedEntry.Event)
	if err != nil {
		return nil, err
	}
	protoLevel, ok := levelToProto[encodedEntry.Level]
	if !ok {
		return nil, fmt.Errorf("lion: unknown level: %v", encodedEntry.Level)
	}
	return &Entry{
		Id:           encodedEntry.ID,
		Level:        protoLevel,
		Timestamp:    google_protobuf.TimeToProto(encodedEntry.Time),
		Context:      contexts,
		Fields:       encodedEntry.Fields,
		Event:        event,
		Message:      encodedEntry.Message,
		WriterOutput: encodedEntry.WriterOutput,
	}, nil
}

func protoEntryToEncodedEntry(protoEntry *Entry) (*lion.EncodedEntry, error) {
	contexts, err := entryMessagesToMessages(protoEntry.Context)
	if err != nil {
		return nil, err
	}
	event, err := entryMessageToMessage(protoEntry.Event)
	if err != nil {
		return nil, err
	}
	level, ok := protoToLevel[protoEntry.Level]
	if !ok {
		return nil, fmt.Errorf("lion: unknown level: %v", protoEntry.Level)
	}
	return &lion.EncodedEntry{
		ID:           protoEntry.Id,
		Level:        level,
		Time:         protoEntry.Timestamp.GoTime(),
		Contexts:     contexts,
		Fields:       protoEntry.Fields,
		Event:        event,
		Message:      protoEntry.Message,
		WriterOutput: protoEntry.WriterOutput,
	}, nil
}

// NOTE: the jsonpb.Marshaler was EPICALLY SLOW in benchmarks
// When using the stdlib json.Marshal function instead for the text Marshaller,
// a speedup of 6X was observed!

func messagesToEntryMessages(messages []*lion.EncodedEntryMessage) ([]*Entry_Message, error) {
	if messages == nil {
		return nil, nil
	}
	entryMessages := make([]*Entry_Message, len(messages))
	for i, message := range messages {
		entryMessage, err := messageToEntryMessage(message)
		if err != nil {
			return nil, err
		}
		entryMessages[i] = entryMessage
	}
	return entryMessages, nil
}

func entryMessagesToMessages(entryMessages []*Entry_Message) ([]*lion.EncodedEntryMessage, error) {
	if entryMessages == nil {
		return nil, nil
	}
	messages := make([]*lion.EncodedEntryMessage, len(entryMessages))
	for i, entryMessage := range entryMessages {
		message, err := entryMessageToMessage(entryMessage)
		if err != nil {
			return nil, err
		}
		messages[i] = message
	}
	return messages, nil
}

func messageToEntryMessage(message *lion.EncodedEntryMessage) (*Entry_Message, error) {
	if message == nil {
		return nil, nil
	}
	return &Entry_Message{
		Encoding: message.Encoding,
		Name:     message.Name,
		Value:    message.Value,
	}, nil
}

func entryMessageToMessage(entryMessage *Entry_Message) (*lion.EncodedEntryMessage, error) {
	if entryMessage == nil {
		return nil, nil
	}
	return &lion.EncodedEntryMessage{
		Encoding: entryMessage.Encoding,
		Name:     entryMessage.Name,
		Value:    entryMessage.Value,
	}, nil
}
