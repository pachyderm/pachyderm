package lion

import (
	"fmt"
	"strconv"
)

const (
	// LevelNone represents no Level.
	LevelNone Level = 0
	// LevelDebug is the debug Level.
	LevelDebug Level = 1
	// LevelInfo is the info Level.
	LevelInfo Level = 2
	// LevelWarn is the warn Level.
	LevelWarn Level = 3
	// LevelError is the error Level.
	LevelError Level = 4
	// LevelFatal is the fatal Level.
	LevelFatal Level = 5
	// LevelPanic is the panic Level.
	LevelPanic Level = 6
)

var (
	levelToName = map[Level]string{
		LevelNone:  "NONE",
		LevelDebug: "DEBUG",
		LevelInfo:  "INFO",
		LevelWarn:  "WARN",
		LevelError: "ERROR",
		LevelFatal: "FATAL",
		LevelPanic: "PANIC",
	}
	nameToLevel = map[string]Level{
		"NONE":  LevelNone,
		"DEBUG": LevelDebug,
		"INFO":  LevelInfo,
		"WARN":  LevelWarn,
		"ERROR": LevelError,
		"FATAL": LevelFatal,
		"PANIC": LevelPanic,
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
