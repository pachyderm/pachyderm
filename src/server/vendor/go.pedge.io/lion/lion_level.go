package lion

import (
	"fmt"
	"strconv"
)

const (
	// LevelDebug is the debug Level.
	LevelDebug Level = 0
	// LevelInfo is the info Level.
	LevelInfo Level = 1
	// LevelWarn is the warn Level.
	LevelWarn Level = 2
	// LevelError is the error Level.
	LevelError Level = 3
	// LevelFatal is the fatal Level.
	LevelFatal Level = 4
	// LevelPanic is the panic Level.
	LevelPanic Level = 5
	// LevelNone represents no Level.
	// It is always logged.
	LevelNone Level = 6
)

var (
	levelToName = map[Level]string{
		LevelDebug: "DEBUG",
		LevelInfo:  "INFO",
		LevelWarn:  "WARN",
		LevelError: "ERROR",
		LevelFatal: "FATAL",
		LevelPanic: "PANIC",
		LevelNone:  "NONE",
	}
	nameToLevel = map[string]Level{
		"DEBUG": LevelDebug,
		"INFO":  LevelInfo,
		"WARN":  LevelWarn,
		"ERROR": LevelError,
		"FATAL": LevelFatal,
		"PANIC": LevelPanic,
		"NONE":  LevelNone,
	}
)

// Level is a logging level.
type Level int32

// String returns the name of a Level or the numerical value if the Level is unknown.
func (l Level) String() string {
	name, ok := levelToName[l]
	if !ok {
		return strconv.Itoa(int(l))
	}
	return name
}

// NameToLevel returns the Level for the given name.
func NameToLevel(name string) (Level, error) {
	level, ok := nameToLevel[name]
	if !ok {
		return LevelNone, fmt.Errorf("lion: no level for name: %s", name)
	}
	return level, nil
}
