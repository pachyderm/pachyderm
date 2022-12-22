package log

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
)

type amazon log.Logger

// Log implements aws.Logger.
func (a *amazon) Log(args ...any) {
	(*log.Logger)(a).Print(args...)
}

var _ aws.Logger = new(amazon)

// NewAmazonLogger returns an aws.Logger.
func NewAmazonLogger(ctx context.Context) aws.Logger {
	a := amazon(*NewStdLogAt(ctx, DebugLevel))
	return &a
}
