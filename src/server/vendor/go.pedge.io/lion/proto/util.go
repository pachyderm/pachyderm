package protolion

import (
	"fmt"
	"reflect"

	golang "github.com/golang/protobuf/proto"
)

func newMessage(name string) (golang.Message, error) {
	reflectType, err := messageType(name)
	if err != nil {
		return nil, err
	}
	return reflect.New(reflectType.Elem()).Interface().(golang.Message), nil
}

func messageType(name string) (reflect.Type, error) {
	t, err := messageTypeForPackage(globalPrimaryPackage, name)
	if err != nil {
		return nil, err
	}
	if t != nil {
		return t, nil
	}
	if !globalOnlyPrimaryPackage {
		t, err = messageTypeForPackage(globalSecondaryPackage, name)
		if err != nil {
			return nil, err
		}
		if t != nil {
			return t, nil
		}
	}
	return nil, fmt.Errorf("lion: no type for name %v", name)
}

func messageTypeForPackage(pkg string, name string) (reflect.Type, error) {
	switch pkg {
	case "golang":
		return golang.MessageType(name), nil
	//case "gogo":
	//return gogo.MessageType(name), nil
	default:
		return nil, fmt.Errorf("lion: unknown package: %s", pkg)
	}
}

func messageName(message golang.Message) (string, error) {
	name, err := messageNameForPackage(globalPrimaryPackage, message)
	if err != nil {
		return "", err
	}
	if name != "" {
		return name, nil
	}
	if !globalOnlyPrimaryPackage {
		name, err = messageNameForPackage(globalSecondaryPackage, message)
		if err != nil {
			return "", err
		}
		if name != "" {
			return name, nil
		}
	}
	return "", fmt.Errorf("lion: no name for message %v", message)
}

func messageNameForPackage(pkg string, message golang.Message) (string, error) {
	switch pkg {
	case "golang":
		return golang.MessageName(message), nil
	//case "gogo":
	//return gogo.MessageName(message), nil
	default:
		return "", fmt.Errorf("lion: unknown package: %s", pkg)
	}
}
