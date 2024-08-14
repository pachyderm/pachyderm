import {
  Permission,
  ResourceType,
  useAuthorize,
} from '@dash-frontend/hooks/useAuthorize';
import {useCurrentPipeline} from '@dash-frontend/hooks/useCurrentPipeline';
import {useJobOrJobs} from '@dash-frontend/hooks/useJobOrJobs';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {LINEAGE_PIPELINE_PATH} from '@dash-frontend/views/Project/constants/projectPaths';

const usePipelineDetails = () => {
  const {pipelineId, projectId} = useUrlState();
  const {searchParams} = useUrlQueryState();
  const {
    loading: pipelineLoading,
    pipeline,
    isServiceOrSpout,
    isSpout,
  } = useCurrentPipeline();
  const {job, loading: jobsLoading} = useJobOrJobs(
    {
      projectId,
      pipelineName: pipelineId,
      id: searchParams.globalIdFilter || '',
    },
    !isSpout && !!pipeline && !pipelineLoading,
  );

  const {
    hasRepoModifyBindings: hasRepoEditRoles,
    hasRepoRead: hasPipelineRead,
  } = useAuthorize({
    permissions: [Permission.REPO_MODIFY_BINDINGS, Permission.REPO_READ],
    resource: {type: ResourceType.REPO, name: `${projectId}/${pipelineId}`},
  });

  const tabsBasePath = LINEAGE_PIPELINE_PATH;

  return {
    projectId,
    pipelineId,
    pipelineAndJobloading: pipelineLoading || jobsLoading,
    pipeline,
    lastJob: job,
    isServiceOrSpout,
    isSpout,
    tabsBasePath,
    hasRepoEditRoles,
    hasPipelineRead,
    globalId: searchParams.globalIdFilter,
  };
};

export default usePipelineDetails;
