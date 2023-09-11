import {NodeType, Permission, ResourceType} from '@graphqlTypes';
import {useState, useMemo} from 'react';

import {useGetDagQuery} from '@dash-frontend/generated/hooks';
import useCurrentRepo from '@dash-frontend/hooks/useCurrentRepo';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {useVerifiedAuthorization} from '@dash-frontend/hooks/useVerifiedAuthorization';

const useDeleteRepoButton = () => {
  const [modalOpen, setModalOpen] = useState(false);
  const {repoId, projectId} = useUrlState();
  const {repo, loading: repoLoading} = useCurrentRepo();
  const {data: dagData, loading: dagLoading} = useGetDagQuery({
    variables: {args: {projectId}},
  });

  const {isAuthorizedAction: hasAuthDeleteRepo} = useVerifiedAuthorization({
    permissionsList: [Permission.REPO_DELETE],
    resource: {type: ResourceType.REPO, name: `${projectId}/${repo?.id}`},
  });

  const canDelete = useMemo(
    () =>
      !dagData?.dag?.some(({parents, project, name, type}) => {
        const isDescendant = parents.some(
          ({project: parentProject, name: parentName}) =>
            parentProject === projectId && parentName === repoId,
        );
        const isOutputRepo =
          project === projectId &&
          name === repoId &&
          type === NodeType.PIPELINE;

        return isDescendant || isOutputRepo;
      }),
    [projectId, repoId, dagData?.dag],
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
