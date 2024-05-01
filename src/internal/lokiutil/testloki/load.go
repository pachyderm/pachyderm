package testloki

import (
	"bufio"
	"context"
	"io"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

func AddLogFile(ctx context.Context, r io.Reader, l *TestLoki) error {
	s := bufio.NewScanner(r)
	labels := map[string]string{}
	var i int
	for s.Scan() {
		i++
		line := s.Text()
		switch {
		case strings.HasPrefix(line, "&map[") && strings.HasSuffix(line, "]"):
			labels = parseLabels(line)
		default:
			log := parseLog(line)
			log.Labels = labels
			if err := l.AddLog(ctx, log); err != nil {
				return errors.Wrapf(err, "line %d: AddLog", i)

			}
		}
	}
	if err := s.Err(); err != nil {
		return errors.Wrap(err, "scan")
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

func parseLog(line string) *Log {
	return &Log{
		Message: line,
	}
}
