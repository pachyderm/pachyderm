package minikubetestenv

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"google.golang.org/protobuf/types/known/durationpb"
)

func getCoverageFolder() string {
	if folder := os.Getenv("TEST_RESULTS"); folder != "" {
		return folder
	}
	return "/tmp/test-results"
}

func saveCoverProfile(pachClient *client.APIClient, covFile *os.File) error {
	return pachClient.Profile(&debug.Profile{
		Name:     "cover",
		Duration: durationpb.New(30 * time.Minute), // no test should be longer than 30 minutes
	}, &debug.Filter{}, covFile)
}

func collectMinikubeCodeCoverage(t testing.TB, pachClient *client.APIClient, valueOverrides map[string]string) {
	coverageFolder := getCoverageFolder()
	coverageFilePath := filepath.Join(coverageFolder,
		fmt.Sprintf("%s-%s-cover.tgz",
			strings.ReplaceAll(filepath.ToSlash(t.Name()), "/", "-"), // ToSlash is for hypothetical different os
			uuid.NewWithoutDashes()),
	)

	if _, err := os.Stat(coverageFilePath); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(coverageFolder, os.ModePerm); err != nil {
			t.Logf("Error creating code coverage folder: %v", err)
			return
		}
	} else { // Since we are using a uuid in the name it should never exists
		t.Logf("Error checking coverage file: %v", err)
		return
	}

	covFile, err := os.Create(coverageFilePath)
	if err != nil {
		t.Logf("Error creating code coverage file: %v", err)
		return
	}
	defer func() {
		if err := covFile.Close(); err != nil {
			t.Logf("Error closing code coverage file: %v", err)
			return
		}
	}()

	err = saveCoverProfile(pachClient, covFile)
	if errors.Is(err, auth.ErrNotSignedIn) {
		// If auth is enabled at any point in the test, we need to log in first
		token, ok := valueOverrides["pachd.rootToken"]
		if !ok {
			token = testutil.RootToken
		}
		pachClient.SetAuthToken(token)
		err = saveCoverProfile(pachClient, covFile)
	}
	if err != nil {
		t.Logf("Failed pachctl debug call attempting to retrieve code coverage: %v", err)
	} else {
		t.Logf("Successfully output minikube code coverage to file: %v", coverageFilePath)
	}
}
