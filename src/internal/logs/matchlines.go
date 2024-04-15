package logs

import (
	"encoding/json"
	"regexp"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
)

type LineMatcher func(line []byte, fields *LogFields) bool

// MatchLines implements LineMatcher interface
type MatchLines struct {
	matchers [](func(line []byte, fields *LogFields) bool)
	Matchers [](func(line []byte, fields *LogFields) bool)
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

func (m *MatchLines) Match(line string, returnMatch func([]byte) error) error {

	logFields := new(LogFields)
	json.Unmarshal([]byte(line), logFields)

	for _, matcher := range m.Matchers {
		if !matcher([]byte(line), logFields) {
			continue
		}
		if err := returnMatch([]byte(line)); err != nil {
			if errors.Is(err, errutil.ErrBreak) {
				return nil
			}
			return err
		}
	}
	return nil
}

func TrueMatcher() (func(line []byte, fields *LogFields) bool, error) {
	return func(line []byte, fields *LogFields) bool {
		return true
	}, nil
}

func FalseMatcher() (func(line []byte, fields *LogFields) bool, error) {
	return func(line []byte, fields *LogFields) bool {
		return false
	}, nil
}

func TimeMatcher(fromTime time.Time, untilTime time.Time) (func(line []byte, fields *LogFields) bool, error) {
	if fromTime.After(untilTime) {
		return nil, errors.Errorf("invalid parameters, start time %t is after end time %t", fromTime, untilTime)
	}
	return func(line []byte, fields *LogFields) bool {
		if fields == nil {
			return false
		}
		if fields.Ts.Before(fromTime) || fields.Ts.After(untilTime) {
			return false
		}
		return true
	}, nil
}

func RegexMatcher(regexStr string, negate bool) (func(line []byte, fields *LogFields) bool, error) {
	regex, err := regexp.Compile(regexStr)
	if err != nil {
		return nil, errors.Errorf("compile error for regex %q, %q", regexStr, err.Error())
	}
	return func(line []byte, _ *LogFields) bool {
		if regex.MatchString(string(line)) {
			if negate {
				return false
			}
		} else if !negate {
			return false
		}
		return true
	}, nil
}

func PipelineJobMatcher(project string, pipeline string, job string) (func(line []byte, fields *LogFields) bool, error) {
	if project == "" || pipeline == "" || job == "" {
		return nil, errors.Errorf("invalid parameters, cannot be empty")
	}
	return func(line []byte, fields *LogFields) bool {
		if fields == nil {
			return false
		}
		return fields.ProjectName == project && fields.PipelineName == pipeline && fields.JobId == job
	}, nil
}

func ProjectMatcher(project string) (func(line []byte, fields *LogFields) bool, error) {
	if project == "" {
		return nil, errors.Errorf("invalid parameters, cannot be empty")
	}
	return func(line []byte, fields *LogFields) bool {
		if fields == nil {
			return false
		}
		return fields.ProjectName == project
	}, nil
}

func PipelineMatcher(project string, pipeline string) (func(line []byte, fields *LogFields) bool, error) {
	if project == "" || pipeline == "" {
		return nil, errors.Errorf("invalid parameters, cannot be empty")
	}
	return func(line []byte, fields *LogFields) bool {
		if fields == nil {
			return false
		}
		return fields.ProjectName == project && fields.PipelineName == pipeline
	}, nil
}

func JobMatcher(job string) (func(line []byte, fields *LogFields) bool, error) {
	if job == "" {
		return nil, errors.Errorf("invalid parameters, cannot be empty")
	}
	return func(line []byte, fields *LogFields) bool {
		if fields == nil {
			return false
		}
		return fields.JobId == job
	}, nil
}

func DatumMatcher(datum string) (func(line []byte, fields *LogFields) bool, error) {
	if datum == "" {
		return nil, errors.Errorf("invalid parameters, cannot be empty")
	}
	return func(line []byte, fields *LogFields) bool {
		if fields == nil {
			return false
		}
		return fields.DatumId == datum
	}, nil
}

func MasterMatcher(project string, pipeline string) (func(line []byte, fields *LogFields) bool, error) {
	if project == "" || pipeline == "" {
		return nil, errors.Errorf("invalid parameters, cannot be empty")
	}
	return func(line []byte, fields *LogFields) bool {
		if fields == nil {
			return false
		}
		return fields.Master && fields.ProjectName == project && fields.PipelineName == pipeline
	}, nil
}

func StorageMatcher(project string, pipeline string) (func(line []byte, fields *LogFields) bool, error) {
	if project == "" || pipeline == "" {
		return nil, errors.Errorf("invalid parameters, cannot be empty")
	}
	return func(line []byte, fields *LogFields) bool {
		if fields == nil {
			return false
		}
		return fields.ProjectName == project && fields.PipelineName == pipeline
	}, nil
}

func UserMatcher() (func(line []byte, fields *LogFields) bool, error) {
	return func(line []byte, fields *LogFields) bool {
		if fields == nil {
			return false
		}
		return fields.User
	}, nil
}

func LogLevelMatcher(level uint32) (func(line []byte, fields *LogFields) bool, error) {
	if level > 2 {
		return nil, errors.Errorf("invalid parameters, level %d cannot be greater than 2", level)
	}
	return func(line []byte, fields *LogFields) bool {
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
	}, nil
}
