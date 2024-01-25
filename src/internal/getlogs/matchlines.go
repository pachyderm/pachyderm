package getlogs

import (
	"bufio"
	"encoding/json"
	"io"
	"regexp"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
)

// MatchLines implements LineMatcher interface
type MatchLines struct {
	matchers [](func(line []byte, fields *LogFields) bool)
}

type LogFields struct {
	Ts           time.Time
	Level        string
	ProjectName  string
	PipelineName string
	JobId        string
	WorkerId     string
	DatumId      string
	Master       bool
	User         bool
}

func (m *MatchLines) Match(sourceLines io.Reader, returnMatch func([]byte) error) error {

	line := bufio.NewScanner(sourceLines)
	for line.Scan() {
		logBytes := line.Bytes()

		logFields := new(LogFields)
		json.Unmarshal(logBytes, logFields)

		for _, matcher := range m.matchers {
			if !matcher(logBytes, logFields) {
				continue
			}
			if err := returnMatch(logBytes); err != nil {
				if errors.Is(err, errutil.ErrBreak) {
					return nil
				}
				return err
			}
		}
	}
	return nil
}

func (m *MatchLines) MatchTime(fromTime time.Time, untilTime time.Time) error {
	if fromTime.After(untilTime) {
		return errors.Errorf("invalid parameters, start time %t is after end time %t", fromTime, untilTime)
	}
	m.matchers = append(m.matchers, func(line []byte, fields *LogFields) bool {
		if fields == nil {
			return false
		}
		if fields.Ts.Before(fromTime) || fields.Ts.After(untilTime) {
			return false
		}
		return true
	})
	return nil
}

func (m *MatchLines) MatchRegex(regexStr string, negate bool) error {
	regex, err := regexp.Compile(regexStr)
	if err != nil {
		return errors.Errorf("compile error for regex %q, %q", regexStr, err.Error())
	}
	m.matchers = append(m.matchers, func(line []byte, _ *LogFields) bool {
		if regex.MatchString(string(line)) {
			if negate {
				return false
			}
		} else if !negate {
			return false
		}
		return true
	})
	return nil
}

func (m *MatchLines) MatchPipelineJob(project string, pipeline string, job string) {
	if project == "" || pipeline == "" || job == "" {
		return
	}
	m.matchers = append(m.matchers, func(line []byte, fields *LogFields) bool {
		if fields == nil {
			return false
		}
		return fields.ProjectName == project && fields.PipelineName == pipeline && fields.JobId == job
	})
}

func (m *MatchLines) MatchProject(project string) {
	if project == "" {
		return
	}
	m.matchers = append(m.matchers, func(line []byte, fields *LogFields) bool {
		if fields == nil {
			return false
		}
		return fields.ProjectName == project
	})
}

func (m *MatchLines) MatchPipeline(project string, pipeline string) {
	if project == "" || pipeline == "" {
		return
	}
	m.matchers = append(m.matchers, func(line []byte, fields *LogFields) bool {
		if fields == nil {
			return false
		}
		return fields.ProjectName == project && fields.PipelineName == pipeline
	})
}

func (m *MatchLines) MatchJob(job string) {
	if job == "" {
		return
	}
	m.matchers = append(m.matchers, func(line []byte, fields *LogFields) bool {
		if fields == nil {
			return false
		}
		return fields.JobId == job
	})
}

func (m *MatchLines) MatchDatum(datum string) {
	if datum == "" {
		return
	}
	m.matchers = append(m.matchers, func(line []byte, fields *LogFields) bool {
		if fields == nil {
			return false
		}
		return fields.DatumId == datum
	})
}

func (m *MatchLines) MatchMaster(project string, pipeline string) {
	if project == "" || pipeline == "" {
		return
	}
	m.matchers = append(m.matchers, func(line []byte, fields *LogFields) bool {
		if fields == nil {
			return false
		}
		return fields.Master && fields.ProjectName == project && fields.PipelineName == pipeline
	})
}

func (m *MatchLines) MatchStorage(project string, pipeline string) {
	if project == "" || pipeline == "" {
		return
	}
	m.matchers = append(m.matchers, func(line []byte, fields *LogFields) bool {
		if fields == nil {
			return false
		}
		return fields.ProjectName == project && fields.PipelineName == pipeline
	})
}

func (m *MatchLines) MatchUser() {
	m.matchers = append(m.matchers, func(line []byte, fields *LogFields) bool {
		if fields == nil {
			return false
		}
		return fields.User
	})
}

func (m *MatchLines) MatchLogLevel(level uint32) {
	if level > 2 {
		return
	}
	m.matchers = append(m.matchers, func(line []byte, fields *LogFields) bool {
		if fields == nil {
			return false
		}
		switch level {
		case 0:
			return fields.Level == "debug"
		case 1:
			return fields.Level == "info"
		case 2:
			return fields.Level == "error"
		}
		return false
	})
}
