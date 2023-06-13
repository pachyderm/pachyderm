package server

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func (kd *kubeDriver) getEgressSecretEnvVars(pipelineInfo *pps.PipelineInfo) []v1.EnvVar {
	result := []v1.EnvVar{}
	egress := pipelineInfo.Details.Egress
	if egress != nil && egress.GetSqlDatabase() != nil && egress.GetSqlDatabase().GetSecret() != nil {
		secret := egress.GetSqlDatabase().GetSecret()
		result = append(result, v1.EnvVar{
			Name: "PACHYDERM_SQL_PASSWORD", // TODO avoid hardcoding this
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{Name: secret.Name},
					Key:                  secret.Key,
				},
			},
		})
	}
	return result
}

func (kd *kubeDriver) createWorkerSvcAndRc(ctx context.Context, pipelineInfo *pps.PipelineInfo) (retErr error) {
	log.Info(ctx, "upserting workers for pipeline")
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/pps.Master/CreateWorkerRC", // ctx never used, but we want the right one in scope for future uses
		"project", pipelineInfo.Pipeline.Project.GetName(),
		"pipeline", pipelineInfo.Pipeline.Name)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()

	specs, err := ppsutil.SpecsFromPipelineInfo(ctx, kd, pipelineInfo)
	if err != nil {
		return errors.Wrap(err, "could not generate pod spec from pipeline info")
	}

	if specs.Secret != nil {
		if _, err := kd.kubeClient.CoreV1().Secrets(kd.namespace).Create(ctx, specs.Secret, metav1.CreateOptions{}); err != nil {
			if !errutil.IsAlreadyExistError(err) {
				return errors.EnsureStack(err)
			}
		}
	}

	if _, err := kd.kubeClient.CoreV1().ReplicationControllers(kd.namespace).Create(ctx, specs.ReplicationController, metav1.CreateOptions{}); err != nil {
		if !errutil.IsAlreadyExistError(err) {
			return errors.EnsureStack(err)
		}
	}

	for _, service := range specs.Services {
		if _, err := kd.kubeClient.CoreV1().Services(kd.namespace).Create(ctx, service, metav1.CreateOptions{}); err != nil {
			if !errutil.IsAlreadyExistError(err) {
				return errors.EnsureStack(err)
			}
		}
	}

	return nil
}
