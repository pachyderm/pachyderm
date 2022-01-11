package secrets

import "fmt"

// Secret is the type of secret data.
// It prevents the actual contents of the secret from being logged
type Secret string

func (s Secret) String() string {
	return fmt.Sprintf("Secret{len=%d}", len(s))
}
