import {NodeType, Permission, ResourceType} from '@graphqlTypes';
import {useMemo, useState} from 'react';

import {useGetDagQuery} from '@dash-frontend/generated/hooks';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useVerifiedAuthorization} from '@dash-frontend/hooks/useVerifiedAuthorization';

const useDeletePipelineButton = () => {
  const {projectId, pipelineId} = useUrlState();
  const [modalOpen, setModalOpen] = useState(false);
  const {data: dagData, loading: dagLoading} = useGetDagQuery({
    variables: {args: {projectId}},
  });
  const {isAuthorizedAction: hasAuthDeleteRepo} = useVerifiedAuthorization({
    permissionsList: [Permission.REPO_DELETE],
    resource: {type: ResourceType.REPO, name: `${projectId}/${pipelineId}`},
  });

  const canDelete = useMemo(() => {
    return (
      dagData &&
      !dagData?.dag?.some(
        ({parents, type}) =>
          type === NodeType.PIPELINE &&
          parents.some(
            ({project, name}) => project === projectId && name === pipelineId,
          ),
      )
    );
  }, [pipelineId, dagData, projectId]);

  const disableButton = dagLoading || !hasAuthDeleteRepo || !canDelete;

  const tooltipText = () => {
    if (!hasAuthDeleteRepo)
      return 'You need at least repoOwner to delete this.';
    else if (!canDelete)
      return "This pipeline can't be deleted while it has downstream pipelines.";
    return 'Delete Pipeline';
  };

  return {
    modalOpen,
    setModalOpen,
    disableButton,
    tooltipText: tooltipText(),
  };
};

export default useDeletePipelineButton;
