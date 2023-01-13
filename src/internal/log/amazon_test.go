package log

import "testing"

func TestAmazon(t *testing.T) {
	// This is just to test that it doesn't panic.
	ctx, h := testWithCaptureParallel(t)
	l := NewAmazonLogger(ctx)
	l.Log("hello")
	l.Log()
	var any *any
	l.Log(any)
	h.HasALog(t)
}
