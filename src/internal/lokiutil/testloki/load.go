package testloki

import (
	"bufio"
	"context"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

func AddLogFile(ctx context.Context, r io.Reader, l *TestLoki) error {
	s := bufio.NewScanner(r)
	labels := map[string]string{}
	var i int
	var logs []*Log
	for s.Scan() {
		i++
		line := s.Text()
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "&map[") && strings.HasSuffix(line, "]"):
			labels = parseLabels(line)
		default:
			log := parseLog(i, line)
			if log.Time.IsZero() {
				return errors.Errorf("line %d (%q): no time", i, line)
			}
			log.Labels = make(map[string]string)
			for k, v := range labels {
				log.Labels[k] = v
			}
			logs = append(logs, log)
		}
	}
	if err := s.Err(); err != nil {
		return errors.Wrap(err, "scan")
	}
	if err := l.AddLog(ctx, logs...); err != nil {
		return errors.Wrapf(err, "line %d: AddLog", i)
	}
	return nil
}

// parseLabels parses a &map[key:value key2:value2] line.  These appear in the loki-logs.txt files
// from a debug dump.  The line must have the prefix "&map[" and suffix "]".
func parseLabels(line string) map[string]string {
	result := map[string]string{}
	var key, value strings.Builder
	var state bool // false = accumulate key, true = accumulate value
	for _, c := range line[len("&map[") : len(line)-1] {
		switch {
		case c == ' ':
			if key.String() != "" && value.String() != "" {
				result[key.String()] = value.String()
			}
			key.Reset()
			value.Reset()
			state = false
		case c == ':':
			state = true
		default:
			if !state {
				key.WriteRune(c)
			} else {
				value.WriteRune(c)
			}
		}
	}
	if key.String() != "" && value.String() != "" {
		result[key.String()] = value.String()
	}
	return result
}

var (
	findRFC3339 = regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[.]\d+Z)`)
	findDate    = regexp.MustCompile(`(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}.\d{3})`)
	findUnix    = regexp.MustCompile(`"ts":([^,]+),`)
)

func parseLog(i int, line string) *Log {
	var ts time.Time
	if matches := findRFC3339.FindStringSubmatch(line); len(matches) == 2 {
		ts, _ = time.Parse(time.RFC3339Nano, matches[1])
	} else if matches := findDate.FindStringSubmatch(line); len(matches) == 2 {
		ts, _ = time.Parse("2006-01-02 15:04:05.999", matches[1])
	} else if matches := findUnix.FindStringSubmatch(line); len(matches) == 2 {
		if unix, err := strconv.ParseFloat(matches[1], 64); err == nil {
			ts = time.Unix(0, int64(float64(1_000_000_000)*unix))
		}
	} else {
		// Some logs don't have timestamps; console and envoy starting up, mostly.
		ts = time.Date(2024, 04, 25, 17, 0, 0, i, time.UTC)
	}
	return &Log{
		Time:    ts,
		Message: line,
	}
}
