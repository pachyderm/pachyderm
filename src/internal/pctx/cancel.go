package pctx

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// WithCancel is a version of context.WithCancel that returns a cancellation function that sets the
// cancellation cause to the call frame that called the CancelFunc.
func WithCancel(root context.Context) (context.Context, context.CancelFunc) {
	_, createdFile, createdLine, createdOK := runtime.Caller(1)
	ctx, c := context.WithCancelCause(root)
	return ctx, func() {
		_, cancelledFile, cancelledLine, cancelledOK := runtime.Caller(1)
		b := new(strings.Builder)
		if createdOK {
			fmt.Fprintf(b, "context created at %v:%v", createdFile, createdLine)
		}
		if createdOK && cancelledOK {
			b.WriteString("; ")
		}
		if cancelledOK {
			fmt.Fprintf(b, "context cancelled at %v:%v", cancelledFile, cancelledLine)
		}
		err := context.Canceled
		if createdOK || cancelledOK {
			err = errors.Join(context.Canceled, errors.New(b.String()))
		}
		c(err)
	}
}
