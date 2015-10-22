package logrus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode"

	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"go.pedge.io/proto/time"
	"go.pedge.io/protolog"
)

var (
	levelToLogrusLevel = map[protolog.Level]logrus.Level{
		protolog.Level_LEVEL_DEBUG: logrus.DebugLevel,
		protolog.Level_LEVEL_INFO:  logrus.InfoLevel,
		protolog.Level_LEVEL_WARN:  logrus.WarnLevel,
		protolog.Level_LEVEL_ERROR: logrus.ErrorLevel,
		protolog.Level_LEVEL_FATAL: logrus.FatalLevel,
		protolog.Level_LEVEL_PANIC: logrus.PanicLevel,
	}

	jsonpbMarshaller = &jsonpb.Marshaler{}
)

type pusher struct {
	logger  *logrus.Logger
	lock    *sync.Mutex
	options PusherOptions
}

func newPusher(options PusherOptions) *pusher {
	logger := logrus.New()
	if options.Out != nil {
		logger.Out = options.Out
	}
	if options.Hooks != nil && len(options.Hooks) > 0 {
		for _, hook := range options.Hooks {
			logger.Hooks.Add(hook)
		}
	}
	if options.Formatter != nil {
		logger.Formatter = options.Formatter
	}
	return &pusher{logger, &sync.Mutex{}, options}
}

func (p *pusher) Push(entry *protolog.Entry) error {
	logrusEntry, err := p.getLogrusEntry(entry)
	if err != nil {
		return err
	}
	return p.logLogrusEntry(logrusEntry)
}

func (p *pusher) Flush() error {
	if p.options.Out != nil {
		return p.options.Out.Flush()
	}
	return nil
}

func (p *pusher) getLogrusEntry(entry *protolog.Entry) (*logrus.Entry, error) {
	logrusLevel, ok := levelToLogrusLevel[entry.Level]
	if !ok {
		return nil, fmt.Errorf("protolog: no logrus Level for %v", entry.Level)
	}
	logrusEntry := logrus.NewEntry(p.logger)
	logrusEntry.Time = prototime.TimestampToTime(entry.Timestamp)
	logrusEntry.Level = logrusLevel

	if p.options.EnableID {
		logrusEntry.Data["_id"] = entry.Id
	}
	if !p.options.DisableContexts {
		contexts, err := entry.UnmarshalledContexts()
		if err != nil {
			return nil, err
		}
		for _, context := range contexts {
			name := context.ProtologName()
			switch name {
			case "protolog.Fields":
				protologFields, ok := context.(*protolog.Fields)
				if !ok {
					return nil, fmt.Errorf("protolog: expected *protolog.Fields, got %T", context)
				}
				for key, value := range protologFields.Value {
					if value != "" {
						logrusEntry.Data[key] = value
					}
				}
			default:
				if err := addProtoMessage(logrusEntry, context); err != nil {
					return nil, err
				}
			}
		}
	}
	event, err := entry.UnmarshalledEvent()
	if err != nil {
		return nil, err
	}
	name := event.ProtologName()
	switch name {
	case "protolog.Event":
		protologEvent, ok := event.(*protolog.Event)
		if !ok {
			return nil, fmt.Errorf("protolog: expected *protolog.Event, got %T", event)
		}
		logrusEntry.Message = trimRightSpace(protologEvent.Message)
	case "protolog.WriterOutput":
		writerOutput, ok := event.(*protolog.WriterOutput)
		if !ok {
			return nil, fmt.Errorf("protolog: expected *protolog.WriterOutput, got %T", event)
		}
		logrusEntry.Message = trimRightSpace(string(writerOutput.Value))
	default:
		logrusEntry.Data["_event"] = name
		if err := addProtoMessage(logrusEntry, event); err != nil {
			return nil, err
		}
	}
	return logrusEntry, nil
}

func (p *pusher) logLogrusEntry(entry *logrus.Entry) error {
	if err := entry.Logger.Hooks.Fire(entry.Level, entry); err != nil {
		return err
	}
	reader, err := entry.Reader()
	if err != nil {
		return err
	}
	p.lock.Lock()
	defer p.lock.Unlock()
	_, err = io.Copy(entry.Logger.Out, reader)
	return err
}

func addProtoMessage(logrusEntry *logrus.Entry, message proto.Message) error {
	m, err := getFieldsForProtoMessage(message)
	if err != nil {
		return err
	}
	for key, value := range m {
		logrusEntry.Data[key] = value
	}
	return nil
}

func getFieldsForProtoMessage(message proto.Message) (map[string]interface{}, error) {
	buffer := bytes.NewBuffer(nil)
	if err := jsonpbMarshaller.Marshal(buffer, message); err != nil {
		return nil, err
	}
	m := make(map[string]interface{}, 0)
	if err := json.Unmarshal(buffer.Bytes(), &m); err != nil {
		return nil, err
	}
	n := make(map[string]interface{}, len(m))
	for key, value := range m {
		switch value.(type) {
		case map[string]interface{}:
			data, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			n[key] = string(data)
		default:
			n[key] = value
		}
	}
	return n, nil
}

func trimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}
