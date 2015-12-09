package protolog

import (
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
)

func messageToEntryMessage(message proto.Message) (*Entry_Message, error) {
	value, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	return &Entry_Message{
		Name:  messageName(message),
		Value: value,
	}, nil
}

func entryMessageToMessage(entryMessage *Entry_Message) (proto.Message, error) {
	message, err := newMessage(entryMessage.Name)
	if err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(entryMessage.Value, message); err != nil {
		return nil, err
	}
	return message, nil
}

func messagesToEntryMessages(messages []proto.Message) ([]*Entry_Message, error) {
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

func entryMessagesToMessages(entryMessages []*Entry_Message) ([]proto.Message, error) {
	messages := make([]proto.Message, len(entryMessages))
	for i, entryMessage := range entryMessages {
		message, err := entryMessageToMessage(entryMessage)
		if err != nil {
			return nil, err
		}
		messages[i] = message
	}
	return messages, nil
}

func newMessage(name string) (proto.Message, error) {
	reflectType := proto.MessageType(name)
	if reflectType == nil {
		return nil, fmt.Errorf("protolog: no Message registered for name: %s", name)
	}

	return reflect.New(reflectType.Elem()).Interface().(proto.Message), nil
}

func messageName(message proto.Message) string {
	if message == nil {
		return ""
	}
	return proto.MessageName(message)
}
