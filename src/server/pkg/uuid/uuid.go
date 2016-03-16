package uuid

import (
	"strings"

	"github.com/satori/go.uuid"
)

// New returns a new uuid.
func New() string {
	return uuid.NewV4().String()
}

// NewWithoutDashes returns a new uuid without any -'s.
func NewWithoutDashes() string {
	return strings.Replace(New(), "-", "", -1)
}
