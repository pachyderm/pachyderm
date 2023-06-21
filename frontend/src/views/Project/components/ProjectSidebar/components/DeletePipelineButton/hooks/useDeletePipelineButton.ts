import {useMemo, useState} from 'react';

import {useGetDagQuery} from '@dash-frontend/generated/hooks';
import useCurrentOuptutRepoOfPipeline from '@dash-frontend/hooks/useCurrentOuptutRepoOfPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {hasAtLeastRole} from '@dash-frontend/lib/rbac';

const useDeletePipelineButton = () => {
  const {projectId, pipelineId} = useUrlState();
  const [modalOpen, setModalOpen] = useState(false);
  const {data: dagData, loading: dagLoading} = useGetDagQuery({
    variables: {args: {projectId}},
  });
  const {repo, loading: repoLoading} = useCurrentOuptutRepoOfPipeline();

  const hasAuthDeleteRepo = hasAtLeastRole(
    'repoOwner',
    repo?.authInfo?.rolesList,
  );

  const canDelete = useMemo(() => {
    return (
      dagData &&
      !dagData?.dag?.some(
        ({parents, type}) =>
          type === 'PIPELINE' &&
          parents.some((parent) => parent === `${projectId}_${pipelineId}`),
      )
    );
  }, [pipelineId, dagData, projectId]);

  const disableButton =
    repoLoading || dagLoading || !hasAuthDeleteRepo || !canDelete;

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
