import usePipelines from '@dash-frontend/hooks/usePipelines';

export const PIPELINE_LIMIT = 16;
export const WORKER_LIMIT = 8;

const useCommunityEditionBanner = (expiration?: number) => {
  const {pipelines} = usePipelines({projectIds: []});

  const pipelineLimitReached =
    !expiration && pipelines && pipelines.length >= PIPELINE_LIMIT;
  const workerLimitReached =
    !expiration &&
    pipelines &&
    pipelines.some((pipeline) => {
      try {
        const spec = JSON.parse(pipeline?.jsonSpec || '{}');
        return spec?.parallelismSpec?.constant >= WORKER_LIMIT;
      } catch (e) {
        console.warn('failed to parse spec for pipeline ' + pipeline?.id);
      }
      return false;
    });

  return {pipelines, pipelineLimitReached, workerLimitReached};
};

export default useCommunityEditionBanner;
