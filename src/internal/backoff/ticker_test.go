package backoff_test

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
)

func TestTicker(t *testing.T) {
	ctx := pctx.TestContext(t)
	const successOn = 3
	var i = 0

	// This function is successful on "successOn" calls.
	f := func() error {
		i++
		log.Info(ctx, "function is called", zap.Int("i", i))

		if i == successOn {
			log.Info(ctx, "OK")
			return nil
		}

		log.Info(ctx, "error")
		return errors.New("error")
	}

	b := backoff.NewExponentialBackOff()
	ticker := backoff.NewTicker(b)

	var err error
	for range ticker.C {
		if err = f(); err != nil {
			t.Log(err)
			continue
		}

		break
	}
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if i != successOn {
		t.Errorf("invalid number of retries: %d", i)
	}
}
