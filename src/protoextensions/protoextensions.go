package protoextensions

import "go.uber.org/zap/zapcore"

func AddBytes(enc zapcore.ObjectEncoder, key string, buf []byte) {
	enc.AddString(key, "<elided bytes>")
}

func AppendBytes(enc zapcore.ArrayEncoder, buf []byte) {
	enc.AppendString("<elided bytes>")
}
