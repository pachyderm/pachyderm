package webhooks

import "log"

// DefaultLog contains the default logger for webhooks, and prints only info and error messages by default
// for debugs override DefaultLog or see NewLogger for creating one without debugs.
var DefaultLog Logger = new(logger)

// Logger allows for customizable logging
type Logger interface {
	// Info prints basic information.
	Info(string)
	// Error prints error information.
	Error(string)
	// Debug prints information usefull for debugging.
	Debug(string)
}

// NewLogger returns a new logger for use.
func NewLogger(debug bool) Logger {
	return &logger{PrintDebugs: debug}
}

type logger struct {
	PrintDebugs bool
}

// Info prints basic information.
func (l *logger) Info(msg string) {
	log.Println("INFO:", msg)
}

// v prints error information.
func (l *logger) Error(msg string) {
	log.Println("ERROR:", msg)
}

// Debug prints information usefull for debugging.
func (l *logger) Debug(msg string) {
	if !l.PrintDebugs {
		return
	}
	log.Println("DEBUG:", msg)
}
