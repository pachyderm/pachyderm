//go:build unit_test

package server

import (
	"bytes"
	"context"
	"os"
	"path"
	"testing"
	"time"

	"gocloud.dev/blob"

	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj/integrationtests"
	"github.com/pachyderm/pachyderm/v2/src/internal/randutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestSharding(t *testing.T) {
	integrationtests.LoadGoogleParameters(t)
	credFile := path.Join(t.TempDir(), "tmp-google-cred")
	require.NoError(t, os.WriteFile(credFile, []byte(os.Getenv("GOOGLE_CLIENT_CREDS")), 0666))
	require.NoError(t, os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credFile))
	bucketName := os.Getenv("GOOGLE_CLIENT_BUCKET")
	url, err := obj.ParseURL("gs://" + bucketName)
	require.NoError(t, err, "should be able to parse url")
	files := make(map[string]string)
	testDir := "./testing/testdata/urlCoordination"
	dir, err := os.ReadDir(testDir)
	objStoreDir := randutil.UniqueString("url-coord-")
	require.NoError(t, err, "should be able to read dir")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*300000)
	defer cancel()
	bucket, err := openBucket(ctx, url)
	require.NoError(t, err, "should be able to open bucket")
	defer func() {
		require.NoError(t, bucket.Close())
	}()
	for _, file := range dir {
		fileData, err := os.ReadFile(path.Join(testDir, file.Name()))
		require.NoError(t, err, "should be able to read file")
		files[file.Name()] = string(fileData)
		writeToObjStorage(ctx, t, bucket, path.Join(objStoreDir, file.Name()), string(fileData))
	}
	defer func() {
		for file := range files {
			require.NoError(t, bucket.Delete(ctx, path.Join(objStoreDir, file)), "should be able to delete file")
		}
	}()
	var tasks []*PutFileURLTask
	defaultSizeThreshold, defaultNumObjectsThreshold = 3, 3
	require.NoError(t, shardObjects(ctx, url.BucketString()+"/"+objStoreDir,
		func(paths []string, startOffset, endOffset int64) error {
			task := &PutFileURLTask{
				Paths:       paths,
				StartOffset: startOffset,
				EndOffset:   endOffset,
			}
			tasks = append(tasks, task)
			return nil
		}))
	processedFiles := processTasks(ctx, t, tasks, bucket)
	for file, data := range files {
		require.Equal(t, data, processedFiles[file], "files should match")
	}
	for file := range processedFiles {
		if _, ok := files[file]; !ok {
			t.Fatalf("unknown file %s was added by tasks", file)
		}
	}
}

func processTasks(ctx context.Context, t *testing.T, tasks []*PutFileURLTask, bucket *blob.Bucket) map[string]string {
	verifiedFiles := make(map[string]string)
	for _, task := range tasks {
		startOffset := task.StartOffset
		length := int64(-1) // -1 means to read until end of file.
		for i, filePath := range task.Paths {
			func() {
				if i != 0 {
					startOffset = 0
				}
				if i == len(task.Paths)-1 && task.EndOffset != int64(-1) {
					length = task.EndOffset - startOffset
				}
				r, err := bucket.NewRangeReader(ctx, filePath, startOffset, length, nil)
				require.NoError(t, err, "should be able to create range reader")
				defer func() {
					require.NoError(t, r.Close(), "should be able to close file")
				}()
				buf := &bytes.Buffer{}
				_, err = r.WriteTo(buf)
				require.NoError(t, err, "should be able to write to buffer")
				verifiedFiles[path.Base(filePath)] += buf.String()
			}()
		}
	}
	return verifiedFiles
}
