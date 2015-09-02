package protolog

import "github.com/golang/protobuf/proto"

func messageToEntryMessage(message Message, marshalFunc func(proto.Message) ([]byte, error)) (*Entry_Message, error) {
	if marshalFunc == nil {
		marshalFunc = defaultMarshalFunc
	}
	value, err := marshalFunc(message)
	if err != nil {
		return nil, err
	}
	return &Entry_Message{
		Name:  message.ProtologName(),
		Value: value,
	}, nil
}

func entryMessageToMessage(entryMessage *Entry_Message, unmarshalFunc func([]byte, proto.Message) error) (Message, error) {
	if unmarshalFunc == nil {
		unmarshalFunc = defaultUnmarshalFunc
	}
	message, err := newMessage(entryMessage.Name)
	if err != nil {
		return nil, err
	}
	if err := unmarshalFunc(entryMessage.Value, message); err != nil {
		return nil, err
	}
	return message, nil
}

func messagesToEntryMessages(messages []Message, marshalFunc func(proto.Message) ([]byte, error)) ([]*Entry_Message, error) {
	entryMessages := make([]*Entry_Message, len(messages))
	for i, message := range messages {
		entryMessage, err := messageToEntryMessage(message, marshalFunc)
		if err != nil {
			return nil, err
		}
		entryMessages[i] = entryMessage
	}
	return entryMessages, nil
}

func entryMessagesToMessages(entryMessages []*Entry_Message, unmarshalFunc func([]byte, proto.Message) error) ([]Message, error) {
	messages := make([]Message, len(entryMessages))
	for i, entryMessage := range entryMessages {
		message, err := entryMessageToMessage(entryMessage, unmarshalFunc)
		if err != nil {
			return nil, err
		}
		messages[i] = message
	}
	return messages, nil
}
