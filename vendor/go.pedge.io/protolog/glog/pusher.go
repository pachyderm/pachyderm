package glog

import (
	"github.com/golang/glog"
	"go.pedge.io/protolog"
)

type pusher struct {
	marshaller protolog.Marshaller
	logDebug   bool
}

func newPusher(marshaller protolog.Marshaller, logDebug bool) *pusher {
	return &pusher{marshaller, logDebug}
}

func (p *pusher) Flush() error {
	glog.Flush()
	return nil
}

func (p *pusher) Push(entry *protolog.Entry) error {
	if entry.Level == protolog.Level_LEVEL_DEBUG && !p.logDebug {
		return nil
	}
	dataBytes, err := p.marshaller.Marshal(entry)
	if err != nil {
		return err
	}
	data := string(dataBytes)
	switch entry.Level {
	case protolog.Level_LEVEL_DEBUG, protolog.Level_LEVEL_INFO:
		glog.Infoln(data)
	case protolog.Level_LEVEL_WARN:
		glog.Warningln(data)
	case protolog.Level_LEVEL_ERROR:
		glog.Errorln(data)
	case protolog.Level_LEVEL_FATAL:
		// cannot use fatal since this will exit before logging completes,
		// which is particularly important for a multi-pusher
		glog.Errorln(data)
	case protolog.Level_LEVEL_PANIC:
		// cannot use panic since this will panic before logging completes,
		// which is particularly important for a multi-pusher
		glog.Errorln(data)
	}
	return nil
}
