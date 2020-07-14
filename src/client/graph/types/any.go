package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gogo/protobuf/jsonpb"
	proto "github.com/gogo/protobuf/proto"
)

var AnyTypeRegistry = NewTypesRegistry()

// UnmarshalAny unmarshals an Any object based on its TypeUrl type hint.
func UnmarshalAny(any *Any) (interface{}, error) {
	class := any.TypeUrl
	bytes := any.Value

	if class, ok := AnyTypeRegistry.Get(class); !ok {
		return nil, fmt.Errorf("Couldn't find type in registry")
	} else {
		instance := reflect.New(class).Interface()
		err := proto.Unmarshal(bytes, instance.(proto.Message))
		if err != nil {
			return nil, err
		}
		return instance, nil
	}
}

// MarshalAny uses reflection to marshal an interface{} into an Any object and
// sets up its TypeUrl type hint.

func MarshalAny(i interface{}) (*Any, error) {
	msg, ok := i.(proto.Message)
	if !ok {
		err := fmt.Errorf("Unable to convert to proto.Message: %v", i)
		return nil, err
	}
	bytes, err := proto.Marshal(msg)

	if err != nil {
		return nil, err
	}

	return &Any{
		TypeUrl: reflect.ValueOf(i).Elem().Type().Name(),
		Value:   bytes,
	}, nil
}

// marshal any to json
func (a *Any) MarshalJSON() ([]byte, error) {
	obj, err := UnmarshalAny(a)
	if err != nil {
		return []byte{}, err
	}

	bytes, err := json.Marshal(obj)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

func (a *Any) Scan(src interface{}) error {
	var source []byte
	switch src.(type) {
	case []byte:
		source = src.([]byte)
	default:
		return fmt.Errorf("Received unsupported type to Scan(): %v", src)
	}

	err := jsonpb.Unmarshal(bytes.NewBuffer(source), a)
	if err != nil {
		return err
	}

	return nil
}
