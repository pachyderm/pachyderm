import {Commit} from '@graphqlTypes';
import {useHistory} from 'react-router';

import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {lineageRoute} from '@dash-frontend/views/Project/utils/routes';
import {DropdownItem} from '@pachyderm/components';

const useCommitsList = () => {
  const {projectId} = useUrlState();
  const {getNewViewState} = useUrlQueryState();
  const {getPathToFileBrowser} = useFileBrowserNavigation();
  const browserHistory = useHistory();

  const globalIdRedirect = (runId: string) => {
    getNewViewState({
      globalIdFilter: runId,
    });
    browserHistory.push(
      lineageRoute({
        projectId,
      }),
    );
  };

  const inspectCommitRedirect = (commit: Commit) => {
    const fileBrowserLink = getPathToFileBrowser({
      projectId,
      repoId: commit.repoName,
      commitId: commit.id,
      branchId: commit.branch?.name || 'default',
    });
    browserHistory.push(fileBrowserLink);
  };

  const onOverflowMenuSelect = (commit: Commit) => (id: string) => {
    switch (id) {
      case 'apply-run':
        return globalIdRedirect(commit.id);
      case 'inspect-commit':
        return inspectCommitRedirect(commit);
      default:
        return null;
    }
  };

  const iconItems: DropdownItem[] = [
    {
      id: 'inspect-commit',
      content: 'Inspect commit',
      closeOnClick: true,
    },
    {
      id: 'apply-run',
      content: 'Apply Global ID and view in DAG',
      closeOnClick: true,
    },
  ];

  return {
    iconItems,
    onOverflowMenuSelect,
  };
};

export default useCommitsList;
