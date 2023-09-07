package starlark

import (
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"go.starlark.net/starlark"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func marshalStack(stack starlark.CallStack) func(zapcore.ArrayEncoder) error {
	return func(ae zapcore.ArrayEncoder) error {
		if len(stack) == 0 {
			return nil
		}
		ae.AppendString("Traceback (most recent call last):")
		for _, fr := range stack {
			ae.AppendString(fmt.Sprintf("  %s: in %s", fr.Pos, fr.Name))
		}
		return nil
	}
}

func Stack(name string, stack starlark.CallStack) zap.Field {
	if len(stack) == 0 {
		return zap.Skip()
	}
	return zap.Array(name, zapcore.ArrayMarshalerFunc(marshalStack(stack)))
}

func Error(name string, err error) []zap.Field {
	result := []zap.Field{zap.Error(err)}
	eErr := new(starlark.EvalError)
	if errors.As(err, &eErr) {
		result = append(result, zap.Object("starlarkError", zapcore.ObjectMarshalerFunc(func(oe zapcore.ObjectEncoder) (retErr error) {
			if eErr.Msg != err.Error() {
				oe.AddString("msg", eErr.Msg)
			}
			errors.JoinInto(&retErr,
				oe.AddArray("traceback", zapcore.ArrayMarshalerFunc(marshalStack(eErr.CallStack))))
			if ue := errors.Unwrap(eErr); err != nil {
				if ue.Error() != err.Error() {
					oe.AddString("cause", ue.Error())
				}
			}
			return
		})))
	}
	return result
}
