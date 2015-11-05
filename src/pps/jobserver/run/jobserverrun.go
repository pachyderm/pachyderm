package jobserverrun

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/container"
	"github.com/pachyderm/pachyderm/src/pps/persist"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

const (
	InputMountDir  = "/pfs/in"
	OutputMountDir = "/pfs/out"
)

type JobRunner interface {
	Start(*persist.JobInfo) error
}

type TestJobRunner interface {
	JobRunner
	GetJobIDToPersistJobInfo() map[string]*persist.JobInfo
}

type JobRunnerOptions struct {
	RemoveContainers bool
}

func NewJobRunner(
	pfsAPIClient pfs.APIClient,
	persistAPIClient persist.APIClient,
	containerClient container.Client,
	pfsMountDir string,
	options JobRunnerOptions,
) JobRunner {
	return newJobRunner(
		pfsAPIClient,
		persistAPIClient,
		containerClient,
		pfsMountDir,
		options,
	)
}

func NewTestJobRunner() TestJobRunner {
	return newTestJobRunner()
}

func StartJob(client *kube.Client, jobInfo *persist.JobInfo) error {
	_, err := client.Jobs(api.NamespaceDefault).Create(job(jobInfo))
	return err
}

func job(jobInfo *persist.JobInfo) *extensions.Job {
	app := jobInfo.JobId
	return &extensions.Job{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Job",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: jobInfo.JobId,
			Labels: map[string]string{
				"app": app,
			},
		},
		Spec: extensions.JobSpec{
			Selector: &extensions.PodSelector{
				MatchLabels: map[string]string{
					"app": app,
				},
			},
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name: jobInfo.JobId,
					Labels: map[string]string{
						"app": app,
					},
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:    "user",
							Image:   jobInfo.GetTransform().Image,
							Command: jobInfo.GetTransform().Cmd,
						},
					},
				},
			},
		},
	}
}
