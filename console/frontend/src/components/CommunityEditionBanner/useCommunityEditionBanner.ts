import {useAllPipelines} from '@dash-frontend/hooks/useAllPipelines';

export const PIPELINE_LIMIT = 16;
export const WORKER_LIMIT = 8;

const useCommunityEditionBanner = (expiration?: number) => {
  const {pipelines} = useAllPipelines();

  const pipelineLimitReached =
    !expiration && pipelines && pipelines.length >= PIPELINE_LIMIT;
  const workerLimitReached =
    !expiration &&
    pipelines &&
    pipelines.some((pipeline) => {
      try {
        return Number(pipeline?.parallelism ?? 0) >= WORKER_LIMIT;
      } catch (e) {
        console.warn(
          'failed to parse spec for pipeline ' + pipeline?.pipeline?.name,
        );
      }
      return false;
    });

  return {pipelines, pipelineLimitReached, workerLimitReached};
};

export default useCommunityEditionBanner;
