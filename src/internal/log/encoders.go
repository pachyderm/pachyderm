package log

import (
	"go.uber.org/zap/zapcore"
)

var (
	// This represents JSON that is equivalent to our old logrus-generated JSON.
	pachdEncoder = zapcore.EncoderConfig{
		// These are important to keep consistent throughout versions.
		TimeKey:     "time",
		EncodeTime:  zapcore.RFC3339NanoTimeEncoder,
		LevelKey:    "severity",
		EncodeLevel: zapcore.LowercaseLevelEncoder,
		MessageKey:  "message",
		// These don't matter as much.
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// This encodes to a JSON object that can be unmarshaled as a pps.LogMessage.
	workerEncoder = zapcore.EncoderConfig{
		// These are fields/formats in pps.LogMessage.
		TimeKey:    "ts",
		EncodeTime: zapcore.RFC3339NanoTimeEncoder,
		MessageKey: "message",

		// These are optional.
		LevelKey:       "severity",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// This is for pachctl in verbose mode.  The format is relatively unimportant; no other
	// systems are expected to consume these logs messages.
	pachctlEncoder = zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// This is a less chatty console encoder.
	minimalConsoleEncoder = zapcore.EncoderConfig{
		TimeKey:          zapcore.OmitKey,
		LevelKey:         "L",
		NameKey:          "N",
		CallerKey:        "C",
		FunctionKey:      zapcore.OmitKey,
		MessageKey:       "M",
		StacktraceKey:    "S",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: " ",
	}
)
