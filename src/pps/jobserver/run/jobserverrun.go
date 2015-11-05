package jobserverrun

import (
	"github.com/pachyderm/pachyderm/src/pps/persist"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

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
