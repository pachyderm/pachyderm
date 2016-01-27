package lion

import (
	"fmt"
	"io"
	"os"
)

type logger struct {
	pusher       Pusher
	enableID     bool
	idAllocator  IDAllocator
	timer        Timer
	errorHandler ErrorHandler
	level        Level
	contexts     []*EntryMessage
	fields       map[string]string
}

func newLogger(pusher Pusher, options ...LoggerOption) *logger {
	logger := &logger{
		pusher,
		false,
		DefaultIDAllocator,
		DefaultTimer,
		DefaultErrorHandler,
		DefaultLevel,
		make([]*EntryMessage, 0),
		make(map[string]string, 0),
	}
	for _, option := range options {
		option(logger)
	}
	return logger
}

func (l *logger) Flush() error {
	return l.pusher.Flush()
}

func (l *logger) AtLevel(level Level) Logger {
	return &logger{
		l.pusher,
		l.enableID,
		l.idAllocator,
		l.timer,
		l.errorHandler,
		level,
		l.contexts,
		l.fields,
	}
}

func (l *logger) WithEntryMessageContext(context *EntryMessage) Logger {
	return &logger{
		l.pusher,
		l.enableID,
		l.idAllocator,
		l.timer,
		l.errorHandler,
		l.level,
		append(l.contexts, context),
		l.fields,
	}
}

func (l *logger) DebugWriter() io.Writer {
	return l.printWriter(LevelDebug)
}

func (l *logger) InfoWriter() io.Writer {
	return l.printWriter(LevelInfo)
}

func (l *logger) WarnWriter() io.Writer {
	return l.printWriter(LevelWarn)
}

func (l *logger) ErrorWriter() io.Writer {
	return l.printWriter(LevelError)
}

func (l *logger) Writer() io.Writer {
	return l.printWriter(LevelNone)
}

func (l *logger) WithField(key string, value interface{}) Logger {
	contextFields := make(map[string]string, len(l.fields)+1)
	contextFields[key] = fmt.Sprintf("%v", value)
	for key, value := range l.fields {
		contextFields[key] = value
	}
	return &logger{
		l.pusher,
		l.enableID,
		l.idAllocator,
		l.timer,
		l.errorHandler,
		l.level,
		l.contexts,
		contextFields,
	}
}

func (l *logger) WithFields(fields map[string]interface{}) Logger {
	contextFields := make(map[string]string, len(l.fields)+len(fields))
	for key, value := range fields {
		contextFields[key] = fmt.Sprintf("%v", value)
	}
	for key, value := range l.fields {
		contextFields[key] = value
	}
	return &logger{
		l.pusher,
		l.enableID,
		l.idAllocator,
		l.timer,
		l.errorHandler,
		l.level,
		l.contexts,
		contextFields,
	}
}

func (l *logger) WithKeyValues(keyValues ...interface{}) Logger {
	if len(keyValues)%2 != 0 {
		keyValues = append(keyValues, "MISSING")
	}
	fields := make(map[string]interface{}, len(keyValues)/2)
	for i := 0; i < len(keyValues); i += 2 {
		fields[fmt.Sprintf("%v", keyValues[i])] = keyValues[i+1]
	}
	return l.WithFields(fields)
}

func (l *logger) Debug(args ...interface{}) {
	l.print(LevelDebug, nil, fmt.Sprint(args...), nil)
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.print(LevelDebug, nil, fmt.Sprintf(format, args...), nil)
}

func (l *logger) Debugln(args ...interface{}) {
	l.print(LevelDebug, nil, fmt.Sprint(args...), nil)
}

func (l *logger) Info(args ...interface{}) {
	l.print(LevelInfo, nil, fmt.Sprint(args...), nil)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.print(LevelInfo, nil, fmt.Sprintf(format, args...), nil)
}

func (l *logger) Infoln(args ...interface{}) {
	l.print(LevelInfo, nil, fmt.Sprint(args...), nil)
}

func (l *logger) Warn(args ...interface{}) {
	l.print(LevelWarn, nil, fmt.Sprint(args...), nil)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.print(LevelWarn, nil, fmt.Sprintf(format, args...), nil)
}

func (l *logger) Warnln(args ...interface{}) {
	l.print(LevelWarn, nil, fmt.Sprint(args...), nil)
}

func (l *logger) Error(args ...interface{}) {
	l.print(LevelError, nil, fmt.Sprint(args...), nil)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.print(LevelError, nil, fmt.Sprintf(format, args...), nil)
}

func (l *logger) Errorln(args ...interface{}) {
	l.print(LevelError, nil, fmt.Sprint(args...), nil)
}

func (l *logger) Fatal(args ...interface{}) {
	l.print(LevelFatal, nil, fmt.Sprint(args...), nil)
	os.Exit(1)
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.print(LevelFatal, nil, fmt.Sprintf(format, args...), nil)
	os.Exit(1)
}

func (l *logger) Fatalln(args ...interface{}) {
	l.print(LevelFatal, nil, fmt.Sprint(args...), nil)
	os.Exit(1)
}

func (l *logger) Panic(args ...interface{}) {
	l.print(LevelPanic, nil, fmt.Sprint(args...), nil)
	panic(fmt.Sprint(args...))
}

func (l *logger) Panicf(format string, args ...interface{}) {
	l.print(LevelPanic, nil, fmt.Sprintf(format, args...), nil)
	panic(fmt.Sprintf(format, args...))
}

func (l *logger) Panicln(args ...interface{}) {
	l.print(LevelPanic, nil, fmt.Sprint(args...), nil)
	panic(fmt.Sprint(args...))
}

func (l *logger) Print(args ...interface{}) {
	l.print(LevelNone, nil, fmt.Sprint(args...), nil)
}

func (l *logger) Printf(format string, args ...interface{}) {
	l.print(LevelNone, nil, fmt.Sprintf(format, args...), nil)
}

func (l *logger) Println(args ...interface{}) {
	l.print(LevelNone, nil, fmt.Sprint(args...), nil)
}

func (l *logger) LogEntryMessage(level Level, event *EntryMessage) {
	l.print(level, event, "", nil)
}

func (l *logger) print(level Level, event *EntryMessage, message string, writerOutput []byte) {
	if err := l.printWithError(level, event, message, writerOutput); err != nil {
		l.errorHandler.Handle(err)
	}
}

func (l *logger) printWriter(level Level) io.Writer {
	// TODO(pedge): think more about this
	//if !l.isLoggedLevel(level) {
	//return ioutil.Discard
	//}
	return newLogWriter(l, level)
}

func (l *logger) printWithError(level Level, event *EntryMessage, message string, writerOutput []byte) error {
	if !l.isLoggedLevel(level) {
		return nil
	}
	if event != nil {
		if err := checkRegisteredEncoding(event.Encoding); err != nil {
			return err
		}
	}
	entry := &Entry{
		Level: level,
		Time:  l.timer.Now(),
		// TODO(pedge): should copy this but has performance hit
		Contexts: l.contexts,
		// TODO(pedge): should copy this but has performance hit
		Fields:       l.fields,
		Event:        event,
		Message:      message,
		WriterOutput: writerOutput,
	}
	if l.enableID {
		entry.ID = l.idAllocator.Allocate()
	}
	return l.pusher.Push(entry)
}

func (l *logger) isLoggedLevel(level Level) bool {
	return level >= l.level || level == LevelNone
}

type logWriter struct {
	logger *logger
	level  Level
}

func newLogWriter(logger *logger, level Level) *logWriter {
	return &logWriter{logger, level}
}

func (w *logWriter) Write(p []byte) (int, error) {
	if err := w.logger.printWithError(w.level, nil, "", p); err != nil {
		return 0, err
	}
	return len(p), nil
}
