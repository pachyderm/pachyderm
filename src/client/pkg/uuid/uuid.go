package uuid

import (
	"strings"

	"github.com/satori/go.uuid"
)

// New returns a new uuid.
func New() string {
	return uuid.NewV4().String()
}

// NewWithoutDashes returns a new uuid without no "-".
func NewWithoutDashes() string {
	return strings.Replace(New(), "-", "", -1)
}

// NewWithoutUnderscores returns a new uuid without no "_".
func NewWithoutUnderscores() string {
	return strings.Replace(New(), "_", "", -1)
}
