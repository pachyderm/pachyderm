package protolion

import (
	"fmt"

	"go.pedge.io/lion"

	"github.com/golang/protobuf/proto"
)

type encoderDecoder struct{}

func newEncoderDecoder() *encoderDecoder {
	return &encoderDecoder{}
}

func (e *encoderDecoder) Encode(entryMessage *lion.EntryMessage) (*lion.EncodedEntryMessage, error) {
	message, ok := entryMessage.Value.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("protolion: %T not a proto.Message", entryMessage.Value)
	}
	name, err := messageName(message)
	if err != nil {
		return nil, err
	}
	value, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	return &lion.EncodedEntryMessage{
		Encoding: Encoding,
		Name:     name,
		Value:    value,
	}, nil
}

func (e *encoderDecoder) Name(entryMessage *lion.EntryMessage) (string, error) {
	message, ok := entryMessage.Value.(proto.Message)
	if !ok {
		return "", fmt.Errorf("protolion: %T not a proto.Message", entryMessage.Value)
	}
	return messageName(message)
}

func (e *encoderDecoder) Decode(encodedEntryMessage *lion.EncodedEntryMessage) (*lion.EntryMessage, error) {
	message, err := newMessage(encodedEntryMessage.Name)
	if err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(encodedEntryMessage.Value, message); err != nil {
		return nil, err
	}
	return &lion.EntryMessage{
		Encoding: encodedEntryMessage.Encoding,
		Value:    message,
	}, nil
}
