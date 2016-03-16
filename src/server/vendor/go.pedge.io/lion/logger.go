package lion

import (
	"fmt"
	"io"
	"os"
)

var (
	discardLevelLoggerInstance = newDiscardLevelLogger()
)

type logger struct {
	pusher       Pusher
	enableID     bool
	idAllocator  IDAllocator
	timer        Timer
	errorHandler ErrorHandler
	l            Level
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

func (l *logger) Level() Level {
	return l.l
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
	if context == nil {
		return l
	}
	return &logger{
		l.pusher,
		l.enableID,
		l.idAllocator,
		l.timer,
		l.errorHandler,
		l.l,
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
		l.l,
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
		l.l,
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

func (l *logger) Debugf(format string, args ...interface{}) {
	if LevelDebug < l.l {
		return
	}
	l.print(LevelDebug, nil, fmt.Sprintf(format, args...), nil)
}

func (l *logger) Debugln(args ...interface{}) {
	if LevelDebug < l.l {
		return
	}
	l.print(LevelDebug, nil, fmt.Sprint(args...), nil)
}

func (l *logger) Infof(format string, args ...interface{}) {
	if LevelInfo < l.l {
		return
	}
	l.print(LevelInfo, nil, fmt.Sprintf(format, args...), nil)
}

func (l *logger) Infoln(args ...interface{}) {
	if LevelInfo < l.l {
		return
	}
	l.print(LevelInfo, nil, fmt.Sprint(args...), nil)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	if LevelWarn < l.l {
		return
	}
	l.print(LevelWarn, nil, fmt.Sprintf(format, args...), nil)
}

func (l *logger) Warnln(args ...interface{}) {
	if LevelWarn < l.l {
		return
	}
	l.print(LevelWarn, nil, fmt.Sprint(args...), nil)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	if LevelError < l.l {
		return
	}
	l.print(LevelError, nil, fmt.Sprintf(format, args...), nil)
}

func (l *logger) Errorln(args ...interface{}) {
	if LevelError < l.l {
		return
	}
	l.print(LevelError, nil, fmt.Sprint(args...), nil)
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	if LevelFatal < l.l {
		return
	}
	l.print(LevelFatal, nil, fmt.Sprintf(format, args...), nil)
	os.Exit(1)
}

func (l *logger) Fatalln(args ...interface{}) {
	if LevelFatal < l.l {
		return
	}
	l.print(LevelFatal, nil, fmt.Sprint(args...), nil)
	os.Exit(1)
}

func (l *logger) Panicf(format string, args ...interface{}) {
	if LevelPanic < l.l {
		return
	}
	l.print(LevelPanic, nil, fmt.Sprintf(format, args...), nil)
	panic(fmt.Sprintf(format, args...))
}

func (l *logger) Panicln(args ...interface{}) {
	if LevelPanic < l.l {
		return
	}
	l.print(LevelPanic, nil, fmt.Sprint(args...), nil)
	panic(fmt.Sprint(args...))
}

func (l *logger) Printf(format string, args ...interface{}) {
	l.print(LevelNone, nil, fmt.Sprintf(format, args...), nil)
}

func (l *logger) Println(args ...interface{}) {
	l.print(LevelNone, nil, fmt.Sprint(args...), nil)
}

func (l *logger) LogEntryMessage(level Level, event *EntryMessage) {
	if level < l.l {
		return
	}
	if event == nil {
		return
	}
	l.print(level, event, "", nil)
}

func (l *logger) LogDebug() LevelLogger {
	if LevelDebug < l.l {
		return discardLevelLoggerInstance
	}
	return newLevelLogger(l, LevelDebug)
}

func (l *logger) LogInfo() LevelLogger {
	if LevelInfo < l.l {
		return discardLevelLoggerInstance
	}
	return newLevelLogger(l, LevelInfo)
}

func (l *logger) print(level Level, event *EntryMessage, message string, writerOutput []byte) {
	if err := l.printWithError(level, event, message, writerOutput); err != nil {
		l.errorHandler.Handle(err)
	}
}

func (l *logger) printWriter(level Level) io.Writer {
	// TODO(pedge): think more about this
	//if level < l.l {
	//return ioutil.Discard
	//}
	return newLogWriter(l, level)
}

func (l *logger) printWithError(level Level, event *EntryMessage, message string, writerOutput []byte) error {
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

type levelLogger struct {
	*logger
	level Level
}

func newLevelLogger(logger *logger, level Level) *levelLogger {
	return &levelLogger{logger, level}
}

func (l *levelLogger) Printf(format string, args ...interface{}) {
	l.print(l.level, nil, fmt.Sprintf(format, args...), nil)
}

func (l *levelLogger) Println(args ...interface{}) {
	l.print(l.level, nil, fmt.Sprint(args...), nil)
}

func (l *levelLogger) WithField(key string, value interface{}) LevelLogger {
	return &levelLogger{l.logger.WithField(key, value).(*logger), l.level}
}

func (l *levelLogger) WithFields(fields map[string]interface{}) LevelLogger {
	return &levelLogger{l.logger.WithFields(fields).(*logger), l.level}
}

func (l *levelLogger) WithKeyValues(keyValues ...interface{}) LevelLogger {
	return &levelLogger{l.logger.WithKeyValues(keyValues...).(*logger), l.level}
}

func (l *levelLogger) WithEntryMessageContext(context *EntryMessage) LevelLogger {
	return &levelLogger{l.logger.WithEntryMessageContext(context).(*logger), l.level}
}

type discardLevelLogger struct{}

func newDiscardLevelLogger() *discardLevelLogger {
	return &discardLevelLogger{}
}

func (d *discardLevelLogger) Printf(format string, args ...interface{})                 {}
func (d *discardLevelLogger) Println(args ...interface{})                               {}
func (d *discardLevelLogger) WithField(key string, value interface{}) LevelLogger       { return d }
func (d *discardLevelLogger) WithFields(fields map[string]interface{}) LevelLogger      { return d }
func (d *discardLevelLogger) WithKeyValues(keyvalues ...interface{}) LevelLogger        { return d }
func (d *discardLevelLogger) WithEntryMessageContext(context *EntryMessage) LevelLogger { return d }
func (d *discardLevelLogger) LogEntryMessage(level Level, event *EntryMessage)          {}
