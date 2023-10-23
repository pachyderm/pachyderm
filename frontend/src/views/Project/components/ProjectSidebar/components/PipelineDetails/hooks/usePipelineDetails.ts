import {Permission, ResourceType} from '@graphqlTypes';

import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import {useJob} from '@dash-frontend/hooks/useJob';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useVerifiedAuthorization} from '@dash-frontend/hooks/useVerifiedAuthorization';
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
  const {job, loading: jobsLoading} = useJob(
    {
      projectId,
      pipelineName: pipelineId,
      id: searchParams.globalIdFilter || '',
    },
    {
      skip: isSpout || !pipeline || pipelineLoading,
    },
  );

  const {isAuthorizedAction: editRolesPermission} = useVerifiedAuthorization({
    permissionsList: [Permission.REPO_MODIFY_BINDINGS],
    resource: {type: ResourceType.REPO, name: `${projectId}/${pipelineId}`},
  });

  const {isAuthorizedAction: pipelineReadPermission} = useVerifiedAuthorization(
    {
      permissionsList: [Permission.REPO_READ],
      resource: {type: ResourceType.REPO, name: `${projectId}/${pipelineId}`},
    },
  );

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
    editRolesPermission,
    pipelineReadPermission,
    globalId: searchParams.globalIdFilter,
  };
};

export default usePipelineDetails;
