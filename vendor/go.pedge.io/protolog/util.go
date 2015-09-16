package protolog

import "github.com/golang/protobuf/proto"

func messageToEntryMessage(message Message) (*Entry_Message, error) {
	value, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	return &Entry_Message{
		Name:  message.ProtologName(),
		Value: value,
	}, nil
}

func entryMessageToMessage(entryMessage *Entry_Message) (Message, error) {
	message, err := newMessage(entryMessage.Name)
	if err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(entryMessage.Value, message); err != nil {
		return nil, err
	}
	return message, nil
}

func messagesToEntryMessages(messages []Message) ([]*Entry_Message, error) {
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

func entryMessagesToMessages(entryMessages []*Entry_Message) ([]Message, error) {
	messages := make([]Message, len(entryMessages))
	for i, entryMessage := range entryMessages {
		message, err := entryMessageToMessage(entryMessage)
		if err != nil {
			return nil, err
		}
		messages[i] = message
	}
	return messages, nil
}
