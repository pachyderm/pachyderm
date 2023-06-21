import {Permission, ResourceType} from '@graphqlTypes';

import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import {useJobs} from '@dash-frontend/hooks/useJobs';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useVerifiedAuthorization} from '@dash-frontend/hooks/useVerifiedAuthorization';
import {LINEAGE_PIPELINE_PATH} from '@dash-frontend/views/Project/constants/projectPaths';

const usePipelineDetails = () => {
  const {pipelineId, projectId} = useUrlState();
  const {
    loading: pipelineLoading,
    pipeline,
    isServiceOrSpout,
  } = useCurrentPipeline();
  const {jobs, loading: jobsLoading} = useJobs(
    {
      projectId,
      pipelineId,
      limit: 1,
    },
    {
      skip: !pipeline || pipelineLoading,
    },
  );

  const {isAuthorizedAction: editRolesPermission} = useVerifiedAuthorization({
    permissionsList: [Permission.REPO_MODIFY_BINDINGS],
    resource: {type: ResourceType.REPO, name: `${projectId}/${pipelineId}`},
  });

  const tabsBasePath = LINEAGE_PIPELINE_PATH;

  return {
    projectId,
    pipelineId,
    loading: pipelineLoading || jobsLoading,
    pipeline,
    lastJob: jobs[0],
    isServiceOrSpout,
    tabsBasePath,
    editRolesPermission,
  };
};

export default usePipelineDetails;
