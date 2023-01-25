// Code generated by protoc-gen-zap (etc/proto/protoc-gen-zap). DO NOT EDIT.
//
// source: internal/pfsload/pfsload.proto

package pfsload

import (
	protoextensions "github.com/pachyderm/pachyderm/v2/src/protoextensions"
	zapcore "go.uber.org/zap/zapcore"
)

func (x *CommitSpec) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddInt64("count", x.Count)
	modificationsArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Modifications {
			if obj, ok := interface{}(v).(zapcore.ObjectMarshaler); ok {
				enc.AppendObject(obj)
			} else {
				enc.AppendReflected(v)
			}
		}
		return nil
	}
	enc.AddArray("modifications", zapcore.ArrayMarshalerFunc(modificationsArrMarshaller))
	file_sourcesArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.FileSources {
			if obj, ok := interface{}(v).(zapcore.ObjectMarshaler); ok {
				enc.AppendObject(obj)
			} else {
				enc.AppendReflected(v)
			}
		}
		return nil
	}
	enc.AddArray("file_sources", zapcore.ArrayMarshalerFunc(file_sourcesArrMarshaller))
	if obj, ok := interface{}(x.Validator).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("validator", obj)
	} else {
		enc.AddReflected("validator", x.Validator)
	}
	return nil
}

func (x *ModificationSpec) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddInt64("count", x.Count)
	if obj, ok := interface{}(x.PutFile).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("put_file", obj)
	} else {
		enc.AddReflected("put_file", x.PutFile)
	}
	return nil
}

func (x *PutFileSpec) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddInt64("count", x.Count)
	enc.AddString("source", x.Source)
	return nil
}

func (x *PutFileTask) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddInt64("count", x.Count)
	if obj, ok := interface{}(x.FileSource).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("file_source", obj)
	} else {
		enc.AddReflected("file_source", x.FileSource)
	}
	enc.AddInt64("seed", x.Seed)
	enc.AddString("auth_token", x.AuthToken)
	return nil
}

func (x *PutFileTaskResult) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("file_set_id", x.FileSetId)
	protoextensions.AddBytes(enc, "hash", x.Hash)
	return nil
}

func (x *FileSourceSpec) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("name", x.Name)
	if obj, ok := interface{}(x.Random).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("random", obj)
	} else {
		enc.AddReflected("random", x.Random)
	}
	return nil
}

func (x *RandomFileSourceSpec) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	if obj, ok := interface{}(x.Directory).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("directory", obj)
	} else {
		enc.AddReflected("directory", x.Directory)
	}
	sizesArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Sizes {
			if obj, ok := interface{}(v).(zapcore.ObjectMarshaler); ok {
				enc.AppendObject(obj)
			} else {
				enc.AppendReflected(v)
			}
		}
		return nil
	}
	enc.AddArray("sizes", zapcore.ArrayMarshalerFunc(sizesArrMarshaller))
	enc.AddBool("increment_path", x.IncrementPath)
	return nil
}

func (x *RandomDirectorySpec) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	if obj, ok := interface{}(x.Depth).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("depth", obj)
	} else {
		enc.AddReflected("depth", x.Depth)
	}
	enc.AddInt64("run", x.Run)
	return nil
}

func (x *SizeSpec) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddInt64("min_size", x.MinSize)
	enc.AddInt64("max_size", x.MaxSize)
	enc.AddInt64("prob", x.Prob)
	return nil
}

func (x *ValidatorSpec) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	if obj, ok := interface{}(x.Frequency).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("frequency", obj)
	} else {
		enc.AddReflected("frequency", x.Frequency)
	}
	return nil
}

func (x *FrequencySpec) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddInt64("count", x.Count)
	enc.AddInt64("prob", x.Prob)
	return nil
}

func (x *State) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	commitsArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Commits {
			if obj, ok := interface{}(v).(zapcore.ObjectMarshaler); ok {
				enc.AppendObject(obj)
			} else {
				enc.AppendReflected(v)
			}
		}
		return nil
	}
	enc.AddArray("commits", zapcore.ArrayMarshalerFunc(commitsArrMarshaller))
	return nil
}

func (x *State_Commit) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	if obj, ok := interface{}(x.Commit).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("commit", obj)
	} else {
		enc.AddReflected("commit", x.Commit)
	}
	protoextensions.AddBytes(enc, "hash", x.Hash)
	return nil
}
