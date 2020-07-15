package testutil

import (
	"os"
)

// GetTestEnterpriseCode Pulls the enterprise code out of the env var stored in travis
func GetTestEnterpriseCode() string {
	acode, exists := os.LookupEnv("ENT_ACT_CODE")
	if !exists {
		panic("Enterpise activation code not found in env vars")
	}
	return acode
}
