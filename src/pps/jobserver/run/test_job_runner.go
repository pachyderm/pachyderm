package jobserverrun

import (
	"sync"

	"github.com/pachyderm/pachyderm/src/pps/persist"
)

type testJobRunner struct {
	jobIDToPersistJobInfo map[string]*persist.JobInfo
	lock                  *sync.Mutex
}

func newTestJobRunner() *testJobRunner {
	return &testJobRunner{make(map[string]*persist.JobInfo), &sync.Mutex{}}
}

func (j *testJobRunner) Start(persistJobInfo *persist.JobInfo) error {
	j.lock.Lock()
	defer j.lock.Unlock()
	j.jobIDToPersistJobInfo[persistJobInfo.JobId] = copyPersistJobInfo(persistJobInfo)
	return nil
}

func (j *testJobRunner) GetJobIDToPersistJobInfo() map[string]*persist.JobInfo {
	j.lock.Lock()
	defer j.lock.Unlock()
	return copyJobIDToPersistJobInfo(j.jobIDToPersistJobInfo)
}

// TODO(pedge): actually copy, not an issue for now, but we do not want
// changes to the input *persist.JobInfo to affect anything in htere
func copyPersistJobInfo(persistJobInfo *persist.JobInfo) *persist.JobInfo {
	return persistJobInfo
}

// TODO(pedge): actually copy
func copyJobIDToPersistJobInfo(jobIDToPersistJobInfo map[string]*persist.JobInfo) map[string]*persist.JobInfo {
	return jobIDToPersistJobInfo
}
