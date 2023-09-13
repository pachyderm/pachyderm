import {NodeType, Permission, ResourceType} from '@graphqlTypes';
import {useMemo, useState} from 'react';

import {useGetVerticesQuery} from '@dash-frontend/generated/hooks';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useVerifiedAuthorization} from '@dash-frontend/hooks/useVerifiedAuthorization';

const useDeletePipelineButton = () => {
  const {projectId, pipelineId} = useUrlState();
  const [modalOpen, setModalOpen] = useState(false);
  const {data: verticesData, loading: verticesLoading} = useGetVerticesQuery({
    variables: {args: {projectId}},
  });
  const {isAuthorizedAction: hasAuthDeleteRepo} = useVerifiedAuthorization({
    permissionsList: [Permission.REPO_DELETE],
    resource: {type: ResourceType.REPO, name: `${projectId}/${pipelineId}`},
  });

  const canDelete = useMemo(() => {
    return !verticesData?.vertices?.some(
      ({parents, type}) =>
        type === NodeType.PIPELINE &&
        parents.some(
          ({project, name}) => project === projectId && name === pipelineId,
        ),
    );
  }, [pipelineId, verticesData?.vertices, projectId]);

  const disableButton = verticesLoading || !hasAuthDeleteRepo || !canDelete;

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
