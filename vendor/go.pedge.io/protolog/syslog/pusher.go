package syslog

import (
	"fmt"
	"log/syslog"

	"go.pedge.io/protolog"
)

type pusher struct {
	writer     *syslog.Writer
	marshaller protolog.Marshaller
}

func newPusher(writer *syslog.Writer, options PusherOptions) *pusher {
	marshaller := options.Marshaller
	if marshaller == nil {
		marshaller = globalMarshaller
	}
	return &pusher{writer, marshaller}
}

func (p *pusher) Flush() error {
	return nil
}

func (p *pusher) Push(entry *protolog.Entry) error {
	dataBytes, err := p.marshaller.Marshal(entry)
	if err != nil {
		return err
	}
	data := string(dataBytes)
	switch entry.Level {
	case protolog.Level_LEVEL_DEBUG:
		return p.writer.Debug(data)
	case protolog.Level_LEVEL_INFO:
		return p.writer.Info(data)
	case protolog.Level_LEVEL_WARN:
		return p.writer.Warning(data)
	case protolog.Level_LEVEL_ERROR:
		return p.writer.Err(data)
	case protolog.Level_LEVEL_FATAL:
		return p.writer.Crit(data)
	case protolog.Level_LEVEL_PANIC:
		return p.writer.Alert(data)
	default:
		return fmt.Errorf("protolog: unknown level: %v", entry.Level)
	}
}
