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
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("modifications", zapcore.ArrayMarshalerFunc(modificationsArrMarshaller))
	file_sourcesArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.FileSources {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("file_sources", zapcore.ArrayMarshalerFunc(file_sourcesArrMarshaller))
	enc.AddObject("validator", x.Validator)
	return nil
}

func (x *ModificationSpec) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddInt64("count", x.Count)
	enc.AddObject("put_file", x.PutFile)
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
	enc.AddObject("file_source", x.FileSource)
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
	enc.AddObject("random", x.Random)
	return nil
}

func (x *RandomFileSourceSpec) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("directory", x.Directory)
	sizesArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Sizes {
			enc.AppendObject(v)
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
	enc.AddObject("depth", x.Depth)
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
	enc.AddObject("frequency", x.Frequency)
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
			enc.AppendObject(v)
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
	enc.AddObject("commit", x.Commit)
	protoextensions.AddBytes(enc, "hash", x.Hash)
	return nil
}
