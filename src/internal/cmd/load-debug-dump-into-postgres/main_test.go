package main

import (
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

type field struct {
	Time   time.Time
	Fields map[string]any
	Ok     bool
}

func TestPgBouncerParsing(t *testing.T) {
	input := []string{
		"2024-03-22 21:12:39.301 UTC [1] LOG C-0x561debb8dd30: pachyderm/pachyderm@127.0.0.1:37214 login attempt: db=pachyderm user=pachyderm tls=no",
		"2024-03-22 21:12:39.302 UTC [1] LOG C-0x561debb8dd30: pachyderm/pachyderm@127.0.0.1:37214 closing because: client close request (age=0s)",
		"2024-03-22 21:12:48.834 UTC [1] LOG stats: 1 xacts/s, 6 queries/s, in 3419 B/s, out 4077 B/s, xact 2775 us, query 408 us, wait 3 us",
		"2024-03-22 21:13:06.460 UTC [1] LOG S-0x561debb91720: dex/pachyderm@10.96.150.195:5432 closing because: server lifetime over (age=3600s)",
		"2024-03-22 21:13:06.460 UTC",
		"2024-03-22 21:13:06.460 UT",
	}
	want := []field{
		{Time: time.Date(2024, 3, 22, 21, 12, 39, 301000000, time.UTC), Fields: map[string]any{
			"message": "pachyderm/pachyderm@127.0.0.1:37214 login attempt: db=pachyderm user=pachyderm tls=no",
			"peer":    "client",
		}, Ok: true},
		{Time: time.Date(2024, 3, 22, 21, 12, 39, 302000000, time.UTC), Fields: map[string]any{
			"message": "pachyderm/pachyderm@127.0.0.1:37214 closing because: client close request (age=0s)",
			"peer":    "client",
		}, Ok: true},
		{Time: time.Date(2024, 3, 22, 21, 12, 48, 834000000, time.UTC), Fields: map[string]any{
			"stats": "1 xacts/s, 6 queries/s, in 3419 B/s, out 4077 B/s, xact 2775 us, query 408 us, wait 3 us",
		}, Ok: true},
		{Time: time.Date(2024, 3, 22, 21, 13, 06, 460000000, time.UTC), Fields: map[string]any{
			"message": "dex/pachyderm@10.96.150.195:5432 closing because: server lifetime over (age=3600s)",
			"peer":    "server",
		}, Ok: true},
		{Time: time.Date(2024, 3, 22, 21, 13, 06, 460000000, time.UTC), Fields: map[string]any{
			"parts": []string{""},
		}, Ok: true},
		{Ok: false},
	}
	var got []field
	for _, line := range input {
		ts, fields, ok := parsePgBouncerLine([]byte(line))
		got = append(got, field{ts, fields, ok})
	}
	require.NoDiff(t, want, got, nil)
}
