/*
Package lion provides structured logging, configurable to use serialization protocols such as Protocol Buffers.
*/
package lion // import "go.pedge.io/lion"

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

var (
	// DefaultLevel is the default Level.
	DefaultLevel = LevelInfo
	// DefaultIDAllocator is the default IDAllocator.
	DefaultIDAllocator = &idAllocator{instanceID, 0}
	// DefaultTimer is the default Timer.
	DefaultTimer = &timer{}
	// DefaultErrorHandler is the default ErrorHandler.
	DefaultErrorHandler = &errorHandler{}
	// DefaultJSONMarshalFunc is the default JSONMarshalFunc.
	DefaultJSONMarshalFunc = func(writer io.Writer, value interface{}) error {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		_, err = writer.Write(data)
		return err
	}

	// DefaultJSONMarshaller is the default JSON Marshaller.
	DefaultJSONMarshaller = NewJSONMarshaller()

	// DiscardPusher is a Pusher that discards all logs.
	DiscardPusher = discardPusherInstance
	// DiscardLogger is a Logger that discards all logs.
	DiscardLogger = NewLogger(DiscardPusher)

	// DefaultPusher is the default Pusher.
	DefaultPusher = NewTextWritePusher(os.Stderr)
	// DefaultLogger is the default Logger.
	DefaultLogger = NewLogger(DefaultPusher)

	globalLogger          = DefaultLogger
	globalLevel           = DefaultLevel
	globalJSONMarshalFunc = DefaultJSONMarshalFunc
	globalHooks           = make([]GlobalHook, 0)
	globalLock            = &sync.Mutex{}
)

// GlobalHook is a function that handles a change in the global Logger instance.
type GlobalHook func(Logger)

// GlobalLogger returns the global Logger instance.
func GlobalLogger() Logger {
	return globalLogger
}

// GlobalJSONMarshalFunc returns the global JSONMarshalFunc instance.
func GlobalJSONMarshalFunc() JSONMarshalFunc {
	return globalJSONMarshalFunc
}

// SetLogger sets the global Logger instance.
func SetLogger(logger Logger) {
	globalLock.Lock()
	defer globalLock.Unlock()
	globalLogger = logger
	globalLevel = logger.Level()
	for _, globalHook := range globalHooks {
		globalHook(globalLogger)
	}
}

// SetLevel sets the global Logger to to be at the given Level.
func SetLevel(level Level) {
	globalLock.Lock()
	defer globalLock.Unlock()
	globalLogger = globalLogger.AtLevel(level)
	globalLevel = level
	for _, globalHook := range globalHooks {
		globalHook(globalLogger)
	}
}

// SetJSONMarshalFunc sets the global JSONMarshalFunc to be used by default.
func SetJSONMarshalFunc(jsonMarshalFunc JSONMarshalFunc) {
	globalLock.Lock()
	defer globalLock.Unlock()
	globalJSONMarshalFunc = jsonMarshalFunc
	for _, globalHook := range globalHooks {
		globalHook(globalLogger)
	}
}

// AddGlobalHook adds a GlobalHook that will be called any time SetLogger or SetLevel is called.
// It will also be called when added.
func AddGlobalHook(globalHook GlobalHook) {
	globalLock.Lock()
	defer globalLock.Unlock()
	globalHooks = append(globalHooks, globalHook)
	globalHook(globalLogger)
}

// RedirectStdLogger will redirect logs to golang's standard logger to the global Logger instance.
func RedirectStdLogger() {
	AddGlobalHook(
		func(logger Logger) {
			log.SetFlags(0)
			log.SetOutput(logger.Writer())
			log.SetPrefix("")
		},
	)
}

// Flusher is an object that can be flushed to a persistent store.
type Flusher interface {
	Flush() error
}

// BaseLogger is a Logger without the methods that are self-returning.
//
// This is so sub-packages can implement these.
type BaseLogger interface {
	Flusher
	Level() Level
	DebugWriter() io.Writer
	InfoWriter() io.Writer
	WarnWriter() io.Writer
	ErrorWriter() io.Writer
	Writer() io.Writer
	Debugf(format string, args ...interface{})
	Debugln(args ...interface{})
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Warnf(format string, args ...interface{})
	Warnln(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
	Panicf(format string, args ...interface{})
	Panicln(args ...interface{})
	Printf(format string, args ...interface{})
	Println(args ...interface{})
}

// BaseLevelLogger is a LevelLogger without the methods that are self-returning.
//
// This is so sub-packages can implement these.
type BaseLevelLogger interface {
	Printf(format string, args ...interface{})
	Println(args ...interface{})
}

// LevelLogger is a logger tied to a specific Level.
//
// It is returned from a Logger only.
//
// If the requested Level is less than the Logger's level
// when LevelLogger is constructed, a discard logger will be
// returned. This is to help with performance concerns of doing
// lots of WithField/WithFields/etc calls, and then at the end
// finding the level is discarded.
//
// If log calls are ignored, this has better performance than the standard Logger.
// If log calls are not ignored, this has slightly worse performance than the standard logger.
//
// Main use of this is for debug calls.
type LevelLogger interface {
	BaseLevelLogger

	WithField(key string, value interface{}) LevelLogger
	WithFields(fields map[string]interface{}) LevelLogger
	WithKeyValues(keyvalues ...interface{}) LevelLogger

	// This generally should only be used internally or by sub-loggers such as the protobuf Logger.
	WithEntryMessageContext(context *EntryMessage) LevelLogger
	// This generally should only be used internally or by sub-loggers such as the protobuf Logger.
	LogEntryMessage(level Level, event *EntryMessage)
}

// Logger is the main logging interface. All methods are also replicated
// on the package and attached to a global Logger.
type Logger interface {
	BaseLogger

	AtLevel(level Level) Logger
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithKeyValues(keyvalues ...interface{}) Logger

	// This generally should only be used internally or by sub-loggers such as the protobuf Logger.
	WithEntryMessageContext(context *EntryMessage) Logger
	// This generally should only be used internally or by sub-loggers such as the protobuf Logger.
	LogEntryMessage(level Level, event *EntryMessage)

	// NOTE: this function name may change, this is experimental
	LogDebug() LevelLogger
	// NOTE: this function name may change, this is experimental
	LogInfo() LevelLogger
}

// EntryMessage is a context or event in an Entry.
// It has different meanings depending on what state it is in.
type EntryMessage struct {
	// Encoding specifies the encoding to use, if converting to an EncodedEntryMessage, such as "json", "protobuf", "thrift".
	Encoding string `json:"encoding,omitempty"`
	// Value is the unmarshalled value to use. It is expected that Value can be marshalled to JSON.
	Value interface{} `json:"value,omitempty"`
}

// Entry is a log entry.
type Entry struct {
	// ID may not be set depending on LoggerOptions.
	// it is up to the user to determine if ID is required.
	ID    string    `json:"id,omitempty"`
	Level Level     `json:"level,omitempty"`
	Time  time.Time `json:"time,omitempty"`
	// both Contexts and Fields can be set
	Contexts []*EntryMessage   `json:"contexts,omitempty"`
	Fields   map[string]string `json:"fields,omitempty"`
	// zero or one of Event, Message, WriterOutput will be set
	Event        *EntryMessage `json:"event,omitempty"`
	Message      string        `json:"message,omitempty"`
	WriterOutput []byte        `json:"writer_output,omitempty"`
}

// EncodedEntryMessage is an encoded EntryMessage.
type EncodedEntryMessage struct {
	// Encoding specifies the encoding, such as "json", "protobuf", "thrift".
	Encoding string `json:"encoding,omitempty"`
	// Name specifies the globally-unique name of the message type, such as "google.protobuf.Timestamp".
	Name string `json:"name,omitempty"`
	// Value is the encoded value.
	Value []byte `json:"value,omitempty"`
}

// EncodedEntry is an encoded log entry.
type EncodedEntry struct {
	// ID may not be set depending on LoggerOptions.
	// it is up to the user to determine if ID is required.
	ID    string    `json:"id,omitempty"`
	Level Level     `json:"level,omitempty"`
	Time  time.Time `json:"time,omitempty"`
	// both Contexts and Fields can be set
	Contexts []*EncodedEntryMessage `json:"contexts,omitempty"`
	Fields   map[string]string      `json:"fields,omitempty"`
	// zero or one of Event, Message, WriterOutput will be set
	Event        *EncodedEntryMessage `json:"event,omitempty"`
	Message      string               `json:"message,omitempty"`
	WriterOutput []byte               `json:"writer_output,omitempty"`
}

// Encoder encodes EntryMessages.
type Encoder interface {
	Encode(entryMessage *EntryMessage) (*EncodedEntryMessage, error)
	// Name just gets the globally-unique name for an EntryMessage.
	Name(entryMessage *EntryMessage) (string, error)
}

// Decoder decodes EntryMessages.
type Decoder interface {
	Decode(encodedEntryMessage *EncodedEntryMessage) (*EntryMessage, error)
}

// EncoderDecoder is an Encoder and Decoarder.
type EncoderDecoder interface {
	Encoder
	Decoder
}

// RegisterEncoderDecoder registers an EncoderDecoder by encoding.
func RegisterEncoderDecoder(encoding string, encoderDecoder EncoderDecoder) error {
	return registerEncoderDecoder(encoding, encoderDecoder)
}

// Encode encodes an EntryMessage.
func (e *EntryMessage) Encode() (*EncodedEntryMessage, error) {
	return encodeEntryMessage(e)
}

// Decode decodes an encoded EntryMessage.
func (e *EncodedEntryMessage) Decode() (*EntryMessage, error) {
	return decodeEncodedEntryMessage(e)
}

// Name gets the globally-unique name for an EntryMessge.
func (e *EntryMessage) Name() (string, error) {
	return entryMessageName(e)
}

// Encode encodes an Entry.
func (e *Entry) Encode() (*EncodedEntry, error) {
	return encodeEntry(e)
}

// Decode decodes an encoded Entry.
func (e *EncodedEntry) Decode() (*Entry, error) {
	return decodeEncodedEntry(e)
}

// Pusher is the interface used to push Entry objects to a persistent store.
type Pusher interface {
	Flusher
	Push(entry *Entry) error
}

// EncodedPusher pushes EncodedEntry objects.
type EncodedPusher interface {
	Flusher
	Push(encodedEntry *EncodedEntry) error
}

// EncodedPusherToPusher returns a Pusher for the EncodedPusher
func EncodedPusherToPusher(encodedPusher EncodedPusher) Pusher {
	return newEncodedPusherToPusherWrapper(encodedPusher)
}

// IDAllocator allocates unique IDs for Entry objects. The default
// behavior is to allocate a new UUID for the process, then add an
// incremented integer to the end.
type IDAllocator interface {
	Allocate() string
}

// Timer returns the current time. The default behavior is to
// call time.Now().UTC().
type Timer interface {
	Now() time.Time
}

// ErrorHandler handles errors when logging. The default behavior
// is to panic.
type ErrorHandler interface {
	Handle(err error)
}

// LoggerOption is an option for the Logger constructor.
type LoggerOption func(*logger)

// LoggerEnableID enables IDs for the Logger.
func LoggerEnableID() LoggerOption {
	return func(logger *logger) {
		logger.enableID = true
	}
}

// LoggerWithIDAllocator uses the IDAllocator for the Logger.
func LoggerWithIDAllocator(idAllocator IDAllocator) LoggerOption {
	return func(logger *logger) {
		logger.idAllocator = idAllocator
	}
}

// LoggerWithTimer uses the Timer for the Logger.
func LoggerWithTimer(timer Timer) LoggerOption {
	return func(logger *logger) {
		logger.timer = timer
	}
}

// LoggerWithErrorHandler uses the ErrorHandler for the Logger.
func LoggerWithErrorHandler(errorHandler ErrorHandler) LoggerOption {
	return func(logger *logger) {
		logger.errorHandler = errorHandler
	}
}

// NewLogger constructs a new Logger using the given Pusher.
func NewLogger(pusher Pusher, options ...LoggerOption) Logger {
	return newLogger(pusher, options...)
}

// Marshaller marshals Entry objects to be written.
type Marshaller interface {
	Marshal(entry *Entry) ([]byte, error)
}

// NewWritePusher constructs a new Pusher that writes to the given io.Writer.
func NewWritePusher(writer io.Writer, marshaller Marshaller) Pusher {
	return newWritePusher(writer, marshaller)
}

// NewTextWritePusher constructs a new Pusher using a TextMarshaller.
func NewTextWritePusher(writer io.Writer, textMarshallerOptions ...TextMarshallerOption) Pusher {
	return NewWritePusher(
		writer,
		NewTextMarshaller(textMarshallerOptions...),
	)
}

// NewJSONWritePusher constructs a new Pusher using a JSON Marshaller.
func NewJSONWritePusher(writer io.Writer, jsonMarshallerOptions ...JSONMarshallerOption) Pusher {
	return NewWritePusher(
		writer,
		NewJSONMarshaller(jsonMarshallerOptions...),
	)
}

// Puller pulls EncodedEntry objects from a persistent store.
type Puller interface {
	Pull() (*EncodedEntry, error)
}

// Unmarshaller unmarshalls a marshalled EncodedEntry object. At the end
// of a stream, Unmarshaller will return io.EOF.
type Unmarshaller interface {
	Unmarshal(reader io.Reader, encodedtry *EncodedEntry) error
}

// NewReadPuller constructs a new Puller that reads from the given Reader.
func NewReadPuller(reader io.Reader, unmarshaller Unmarshaller) Puller {
	return newReadPuller(reader, unmarshaller)
}

// JSONMarshalFunc marshals JSON for a TextMarshaller.
//
// It is used internally in a TextMarshaller, and is not a Marshaller itself.
type JSONMarshalFunc func(writer io.Writer, value interface{}) error

// TextMarshaller is a Marshaller used for text.
type TextMarshaller interface {
	Marshaller
	WithColors() TextMarshaller
	WithoutColors() TextMarshaller
}

// TextMarshallerOption is an option for creating Marshallers.
type TextMarshallerOption func(*textMarshaller)

// TextMarshallerDisableTime will suppress the printing of Entry Timestamps.
func TextMarshallerDisableTime() TextMarshallerOption {
	return func(textMarshaller *textMarshaller) {
		textMarshaller.disableTime = true
	}
}

// TextMarshallerDisableLevel will suppress the printing of Entry Levels.
func TextMarshallerDisableLevel() TextMarshallerOption {
	return func(textMarshaller *textMarshaller) {
		textMarshaller.disableLevel = true
	}
}

// TextMarshallerDisableContexts will suppress the printing of Entry contexts.
func TextMarshallerDisableContexts() TextMarshallerOption {
	return func(textMarshaller *textMarshaller) {
		textMarshaller.disableContexts = true
	}
}

// TextMarshallerDisableNewlines disables newlines after each marshalled Entry.
func TextMarshallerDisableNewlines() TextMarshallerOption {
	return func(textMarshaller *textMarshaller) {
		textMarshaller.disableNewlines = true
	}
}

// NewTextMarshaller constructs a new Marshaller that produces human-readable
// marshalled Entry objects. This Marshaller is currently inefficient.
func NewTextMarshaller(options ...TextMarshallerOption) TextMarshaller {
	return newTextMarshaller(options...)
}

// JSONMarshallerDisableNewlines disables newlines after each marshalled Entry.
func JSONMarshallerDisableNewlines() JSONMarshallerOption {
	return func(jsonMarshaller *jsonMarshaller) {
		jsonMarshaller.disableNewlines = true
	}
}

// JSONMarshallerOption is an option for creating Marshallers.
type JSONMarshallerOption func(*jsonMarshaller)

// NewJSONMarshaller constructs a new Marshaller for JSON.
func NewJSONMarshaller(options ...JSONMarshallerOption) Marshaller {
	return newJSONMarshaller(options...)
}

// NewMultiPusher constructs a new Pusher that calls all the given Pushers.
func NewMultiPusher(pushers ...Pusher) Pusher {
	return newMultiPusher(pushers)
}

// Flush calls Flush on the global Logger.
func Flush() error {
	return globalLogger.Flush()
}

// DebugWriter calls DebugWriter on the global Logger.
func DebugWriter() io.Writer {
	return globalLogger.DebugWriter()
}

// InfoWriter calls InfoWriter on the global Logger.
func InfoWriter() io.Writer {
	return globalLogger.InfoWriter()
}

// WarnWriter calls WarnWriter on the global Logger.
func WarnWriter() io.Writer {
	return globalLogger.WarnWriter()
}

// ErrorWriter calls ErrorWriter on the global Logger.
func ErrorWriter() io.Writer {
	return globalLogger.ErrorWriter()
}

// Writer calls Writer on the global Logger.
func Writer() io.Writer {
	return globalLogger.Writer()
}

// Debugf calls Debugf on the global Logger.
func Debugf(format string, args ...interface{}) {
	if LevelDebug < globalLevel {
		return
	}
	globalLogger.Debugf(format, args...)
}

// Debugln calls Debugln on the global Logger.
func Debugln(args ...interface{}) {
	if LevelDebug < globalLevel {
		return
	}
	globalLogger.Debugln(args...)
}

// Infof calls Infof on the global Logger.
func Infof(format string, args ...interface{}) {
	if LevelInfo < globalLevel {
		return
	}
	globalLogger.Infof(format, args...)
}

// Infoln calls Infoln on the global Logger.
func Infoln(args ...interface{}) {
	if LevelInfo < globalLevel {
		return
	}
	globalLogger.Infoln(args...)
}

// Warnf calls Warnf on the global Logger.
func Warnf(format string, args ...interface{}) {
	if LevelWarn < globalLevel {
		return
	}
	globalLogger.Warnf(format, args...)
}

// Warnln calls Warnln on the global Logger.
func Warnln(args ...interface{}) {
	if LevelWarn < globalLevel {
		return
	}
	globalLogger.Warnln(args...)
}

// Errorf calls Errorf on the global Logger.
func Errorf(format string, args ...interface{}) {
	if LevelError < globalLevel {
		return
	}
	globalLogger.Errorf(format, args...)
}

// Errorln calls Errorln on the global Logger.
func Errorln(args ...interface{}) {
	if LevelError < globalLevel {
		return
	}
	globalLogger.Errorln(args...)
}

// Fatalf calls Fatalf on the global Logger.
func Fatalf(format string, args ...interface{}) {
	if LevelFatal < globalLevel {
		return
	}
	globalLogger.Fatalf(format, args...)
}

// Fatalln calls Fatalln on the global Logger.
func Fatalln(args ...interface{}) {
	if LevelFatal < globalLevel {
		return
	}
	globalLogger.Fatalln(args...)
}

// Panicf calls Panicf on the global Logger.
func Panicf(format string, args ...interface{}) {
	if LevelPanic < globalLevel {
		return
	}
	globalLogger.Panicf(format, args...)
}

// Panicln calls Panicln on the global Logger.
func Panicln(args ...interface{}) {
	if LevelPanic < globalLevel {
		return
	}
	globalLogger.Panicln(args...)
}

// Printf calls Printf on the global Logger.
func Printf(format string, args ...interface{}) {
	globalLogger.Printf(format, args...)
}

// Println calls Println on the global Logger.
func Println(args ...interface{}) {
	globalLogger.Println(args...)
}

// AtLevel calls AtLevel on the global Logger.
func AtLevel(level Level) Logger {
	return globalLogger.AtLevel(level)
}

// WithField calls WithField on the global Logger.
func WithField(key string, value interface{}) Logger {
	return globalLogger.WithField(key, value)
}

// WithFields calls WithFields on the global Logger.
func WithFields(fields map[string]interface{}) Logger {
	return globalLogger.WithFields(fields)
}

// WithKeyValues calls WithKeyValues on the global Logger.
func WithKeyValues(keyValues ...interface{}) Logger {
	return globalLogger.WithKeyValues(keyValues...)
}

// LogDebug calls LogDebug on the global Logger.
func LogDebug() LevelLogger {
	return globalLogger.LogDebug()
}

// LogInfo calls LogInfo on the global Logger.
func LogInfo() LevelLogger {
	return globalLogger.LogInfo()
}
