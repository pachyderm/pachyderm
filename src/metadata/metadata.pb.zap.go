// Code generated by protoc-gen-zap (etc/proto/protoc-gen-zap). DO NOT EDIT.
//
// source: metadata/metadata.proto

package metadata

import (
	fmt "fmt"
	zapcore "go.uber.org/zap/zapcore"
)

func (x *Edit) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("project", x.GetProject())
	if obj, ok := interface{}(x.GetReplace()).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("replace", obj)
	} else {
		enc.AddReflected("replace", x.GetReplace())
	}
	return nil
}

func (x *Edit_Replace) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("replacement", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		for k, v := range x.Replacement {
			enc.AddString(fmt.Sprintf("%v", k), v)
		}
		return nil
	}))
	return nil
}

func (x *EditMetadataRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	editsArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Edits {
			if obj, ok := interface{}(v).(zapcore.ObjectMarshaler); ok {
				enc.AppendObject(obj)
			} else {
				enc.AppendReflected(v)
			}
		}
		return nil
	}
	enc.AddArray("edits", zapcore.ArrayMarshalerFunc(editsArrMarshaller))
	return nil
}

func (x *EditMetadataResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	return nil
}
