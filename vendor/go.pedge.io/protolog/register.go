package protolog

import "fmt"

var (
	nameToMessageConstructor = make(map[string]func() Message)
)

// Register registers a Message constructor funcation to a message name.
// This should only be called by generated code.
func Register(name string, messageConstructor func() Message) {
	if _, ok := nameToMessageConstructor[name]; ok {
		panic(fmt.Sprintf("protolog: duplicate Message name registered: %s", name))
	}
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
