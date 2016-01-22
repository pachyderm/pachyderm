/*
Package protolion defines the Protocol Buffers functionality for lion.
*/
package protolion // import "go.pedge.io/lion/proto"

import (
	"io"
	"sync"

	"go.pedge.io/lion"

	"github.com/golang/protobuf/proto"
)

var (
	// Encoding is the name of the encoding.
	Encoding = "proto"

	// DelimitedMarshaller is a Marshaller that uses the protocol buffers write delimited scheme.
	DelimitedMarshaller = &delimitedMarshaller{}
	// DelimitedUnmarshaller is an Unmarshaller that uses the protocol buffers write delimited scheme.
	DelimitedUnmarshaller = &delimitedUnmarshaller{}

	globalPrimaryPackage     = "golang"
	globalSecondaryPackage   = "gogo"
	globalOnlyPrimaryPackage = true
	globalLogger             = NewLogger(lion.GlobalLogger())
	globalLock               = &sync.Mutex{}
)

func init() {
	if err := lion.RegisterEncoderDecoder(Encoding, newEncoderDecoder()); err != nil {
		panic(err.Error())
	}
	lion.AddGlobalHook(setGlobalLogger)
}

func setGlobalLogger(logger lion.Logger) {
	globalLock.Lock()
	defer globalLock.Unlock()
	globalLogger = NewLogger(logger)
}

// Logger is a lion.Logger that also has proto logging methods.
type Logger interface {
	lion.BaseLogger

	AtLevel(level lion.Level) Logger
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger

	WithContext(context proto.Message) Logger
	Debug(event proto.Message)
	Info(event proto.Message)
	Warn(event proto.Message)
	Error(event proto.Message)
	Fatal(event proto.Message)
	Panic(event proto.Message)
	Print(event proto.Message)

	LionLogger() lion.Logger
}

// GlobalLogger returns the global Logger instance.
func GlobalLogger() Logger {
	return globalLogger
}

// NewLogger returns a new Logger.
func NewLogger(delegate lion.Logger) Logger {
	return newLogger(delegate)
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
	globalLogger.Debugf(format, args...)
}

// Debugln calls Debugln on the global Logger.
func Debugln(args ...interface{}) {
	globalLogger.Debugln(args...)
}

// Infof calls Infof on the global Logger.
func Infof(format string, args ...interface{}) {
	globalLogger.Infof(format, args...)
}

// Infoln calls Infoln on the global Logger.
func Infoln(args ...interface{}) {
	globalLogger.Infoln(args...)
}

// Warnf calls Warnf on the global Logger.
func Warnf(format string, args ...interface{}) {
	globalLogger.Warnf(format, args...)
}

// Warnln calls Warnln on the global Logger.
func Warnln(args ...interface{}) {
	globalLogger.Warnln(args...)
}

// Errorf calls Errorf on the global Logger.
func Errorf(format string, args ...interface{}) {
	globalLogger.Errorf(format, args...)
}

// Errorln calls Errorln on the global Logger.
func Errorln(args ...interface{}) {
	globalLogger.Errorln(args...)
}

// Fatalf calls Fatalf on the global Logger.
func Fatalf(format string, args ...interface{}) {
	globalLogger.Fatalf(format, args...)
}

// Fatalln calls Fatalln on the global Logger.
func Fatalln(args ...interface{}) {
	globalLogger.Fatalln(args...)
}

// Panicf calls Panicf on the global Logger.
func Panicf(format string, args ...interface{}) {
	globalLogger.Panicf(format, args...)
}

// Panicln calls Panicln on the global Logger.
func Panicln(args ...interface{}) {
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
func AtLevel(level lion.Level) Logger {
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

// WithContext calls WithContext on the global Logger.
func WithContext(context proto.Message) Logger {
	return globalLogger.WithContext(context)
}

// Debug calls Debug on the global Logger.
func Debug(event proto.Message) {
	globalLogger.Debug(event)
}

// Info calls Info on the global Logger.
func Info(event proto.Message) {
	globalLogger.Info(event)
}

// Warn calls Warn on the global Logger.
func Warn(event proto.Message) {
	globalLogger.Warn(event)
}

// Error calls Error on the global Logger.
func Error(event proto.Message) {
	globalLogger.Error(event)
}

// Fatal calls Fatal on the global Logger.
func Fatal(event proto.Message) {
	globalLogger.Fatal(event)
}

// Panic calls Panic on the global Logger.
func Panic(event proto.Message) {
	globalLogger.Panic(event)
}

// Print calls Print on the global Logger.
func Print(event proto.Message) {
	globalLogger.Print(event)
}

// LionLogger calls LionLogger on the global Logger.
func LionLogger() lion.Logger {
	return globalLogger.LionLogger()
}

//// GolangFirst says to check both golang and gogo for message names and types, but golang first.
//func GolangFirst() {
//globalLock.Lock()
//defer globalLock.Unlock()
//globalPrimaryPackage = "golang"
//globalSecondaryPackage = "gogo"
//globalOnlyPrimaryPackage = false
//}

//// GolangOnly says to check only golang for message names and types, but not gogo.
//func GolangOnly() {
//globalLock.Lock()
//defer globalLock.Unlock()
//globalPrimaryPackage = "golang"
//globalSecondaryPackage = "gogo"
//globalOnlyPrimaryPackage = true
//}

//// GogoFirst says to check both gogo and golang for message names and types, but gogo first.
//func GogoFirst() {
//globalLock.Lock()
//defer globalLock.Unlock()
//globalPrimaryPackage = "gogo"
//globalSecondaryPackage = "golang"
//globalOnlyPrimaryPackage = false
//}

//// GogoOnly says to check only gogo for message names and types, but not golang.
//func GogoOnly() {
//globalLock.Lock()
//defer globalLock.Unlock()
//globalPrimaryPackage = "gogo"
//globalSecondaryPackage = "golang"
//globalOnlyPrimaryPackage = true
//}
