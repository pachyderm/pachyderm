package uuid

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	uuid "github.com/satori/go.uuid"
)

// New returns a new uuid.
func New() string {
	var result string
	backoff.RetryNotify(func() error {
		uuid := uuid.NewV4()
		result = uuid.String()
		return nil
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		fmt.Printf("error from uuid.NewV4: %v", err)
		return nil
	})
	return result
}

// NewWithoutDashes returns a new uuid without no "-".
func NewWithoutDashes() string {
	return strings.Replace(New(), "-", "", -1)
}

// NewWithoutUnderscores returns a new uuid without no "_".
func NewWithoutUnderscores() string {
	return strings.Replace(New(), "_", "", -1)
}

// IsUUIDWithoutDashes checks whether a string is a UUID without dashes
func IsUUIDWithoutDashes(s string) bool {
	return uuidWithoutDashesRegexp.MatchString(s)
}

// Because we use UUIDv4, the 13th character is a '4'.
// Moreover, a UUID can only contain "hexadecimal" characters,
// lowercase here.
var uuidWithoutDashesRegexp = regexp.MustCompile("[0-9a-f]{12}4[0-9a-f]{19}")
