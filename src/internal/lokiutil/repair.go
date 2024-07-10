package lokiutil

// This file repairs log lines from broken Loki configurations.  Learn about the log formats here:
// https://github.com/kubernetes/design-proposals-archive/blob/main/node/kubelet-cri-logging.md

import (
	"encoding/json"
	"strings"
	"time"
)

func tryDocker(input map[string]any) (result string) {
	for k, v := range input {
		switch k {
		case "log":
			if s, ok := v.(string); ok {
				result = s
				continue
			}
			// Malformed.  "log" must be a string.
			return ""
		case "stream", "time":
			continue
		default:
			// Docker logs contain only "log", "stream", and "time" keys.  time
			// and stream are treated as optional.  If we see a different key,
			// then this isn't Docker log JSON.
			return ""
		}
	}
	return
}

func tryCRI(input string) (result string) {
	parts := strings.SplitN(input, " ", 4)
	if len(parts) != 4 {
		return input
	}
	ts, stream, flag, msg := parts[0], parts[1], parts[2], parts[3]
	switch flag {
	case "P", "F":
	default:
		return input
	}
	switch stream {
	case "stdout", "stderr":
	default:
		return input
	}
	if _, err := time.Parse(time.RFC3339Nano, ts); err != nil {
		return input
	}
	return msg
}

// RepairLine takes a log line from a potentially-misconfigured Loki instance and returns the actual
// logged text.  If the logged text is JSON, the parsed text is returned as well.  Downstream
// consumers expecting JSON should not attempt an additional parse of the returned line.
//
// The misconfiguration is where promtail is told the lines are Docker JSON, but they're actually
// CRI text, or promtail is told the lines are CRI plain text, but are actually Docker JSON.  In
// these cases, promtail just sends whatever is in the file that it's tailing instead of just the
// message.  (It's likely that stdout/stderr information is lost in this case as well, and we don't
// attempt to recover that, since we don't typically care.)
func RepairLine(line string) (resultLine string, resultJSON map[string]any) {
	if line == "" {
		return "", nil
	}
	resultLine = line

	switch line[0] {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		// This will be '2' for the next 980 years ;)
		resultLine = tryCRI(line)

	case '{':
		if err := json.Unmarshal([]byte(line), &resultJSON); err != nil {
			// The log is not JSON, and not in Docker format, so pass through the
			// original line and return nothing for the JSON object.
			return line, nil
		}
		l := tryDocker(resultJSON)
		if l == "" {
			// The log is not Docker JSON, which means it's just a normal JSON log line.
			// Return the original line and the parsed JSON.
			return line, resultJSON
		}
		resultLine, resultJSON = l, nil
	}
	if err := json.Unmarshal([]byte(resultLine), &resultJSON); err != nil {
		return resultLine, nil
	}
	return resultLine, resultJSON
}
