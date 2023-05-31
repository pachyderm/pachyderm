import {ReposWithCommitQuery, Commit} from '@graphqlTypes';
import {useHistory} from 'react-router';

import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {repoRoute} from '@dash-frontend/views/Project/utils/routes';
import {DropdownItem} from '@pachyderm/components';

const useRepositoriesList = () => {
  const {projectId} = useUrlState();
  const browserHistory = useHistory();
  const {getPathToFileBrowser} = useFileBrowserNavigation();

  const inspectCommitRedirect = (commit: Commit | null | undefined) => {
    if (commit) {
      const fileBrowserLink = getPathToFileBrowser({
        projectId,
        repoId: commit.repoName,
        commitId: commit.id,
        branchId: commit.branch?.name || 'default',
      });
      browserHistory.push(fileBrowserLink);
    }
  };

  const onOverflowMenuSelect =
    (repo: ReposWithCommitQuery['repos'][number]) => (id: string) => {
      switch (id) {
        case 'dag':
          return browserHistory.push(
            repoRoute({projectId, repoId: repo?.id || ''}),
          );
        case 'inspect-commits':
          return inspectCommitRedirect(repo?.lastCommit);
        default:
          return null;
      }
    };

  const generateIconItems = (commitId?: string): DropdownItem[] => [
    {
      id: 'inspect-commits',
      content: 'Inspect commits',
      closeOnClick: true,
      disabled: commitId === undefined,
    },
    {
      id: 'dag',
      content: 'View in DAG',
      closeOnClick: true,
    },
  ];

  return {
    generateIconItems,
    onOverflowMenuSelect,
  };
};

export default useRepositoriesList;
