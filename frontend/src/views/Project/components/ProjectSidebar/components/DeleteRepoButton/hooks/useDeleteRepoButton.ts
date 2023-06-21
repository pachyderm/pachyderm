import {useState} from 'react';

import {useGetDagQuery} from '@dash-frontend/generated/hooks';
import useCurrentRepo from '@dash-frontend/hooks/useCurrentRepo';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {hasAtLeastRole} from '@dash-frontend/lib/rbac';

const useDeleteRepoButton = () => {
  const [modalOpen, setModalOpen] = useState(false);
  const {repoId, projectId} = useUrlState();
  const {repo, loading: repoLoading} = useCurrentRepo();
  const {data: dagData, loading: dagLoading} = useGetDagQuery({
    variables: {args: {projectId}},
  });

  const hasAuthDeleteRepo = hasAtLeastRole(
    'repoOwner',
    repo?.authInfo?.rolesList,
  );

  const canDelete =
    dagData &&
    !dagData?.dag?.some(({parents}) =>
      parents.some((parent) => parent === `${projectId}_${repoId}`),
    );

  const disableButton =
    repoLoading || dagLoading || !hasAuthDeleteRepo || !canDelete;

  const tooltipText = () => {
    if (!hasAuthDeleteRepo)
      return 'You need at least repoOwner to delete this.';
    else if (!canDelete)
      return "This repo can't be deleted while it has associated pipelines.";
    return 'Delete Repo';
  };

  return {
    modalOpen,
    setModalOpen,
    disableButton,
    tooltipText: tooltipText(),
  };
};

export default useDeleteRepoButton;
