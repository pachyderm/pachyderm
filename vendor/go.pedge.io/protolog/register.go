package protolog

import "fmt"

var (
	nameToMessageConstructor = make(map[string]func() Message)
	registeredMessages       = make(map[registeredMessage]bool)
)

type registeredMessage struct {
	name        string
	messageType MessageType
}

// Register registers a Message constructor funcation to a message name.
// This should only be called by generated code.
func Register(name string, messageType MessageType, messageConstructor func() Message) {
	registeredMessage := registeredMessage{name: name, messageType: messageType}
	if _, ok := registeredMessages[registeredMessage]; ok {
		panic(fmt.Sprintf("protolog: duplicate Message registered: %s", name))
	}
	registeredMessages[registeredMessage] = true
	nameToMessageConstructor[name] = messageConstructor
}

func isNameRegistered(name string) bool {
	_, ok := nameToMessageConstructor[name]
	return ok
}

func checkNameRegistered(name string) error {
	if !isNameRegistered(name) {
		return fmt.Errorf("protolog: no Message registered for name: %s", name)
	}
	return nil
}

func newMessage(name string) (Message, error) {
	messageConstructor, ok := nameToMessageConstructor[name]
	if !ok {
		return nil, fmt.Errorf("protolog: no Message registered for name: %s", name)
	}
	return messageConstructor(), nil
}
