import {DropdownItem} from '@pachyderm/components';
import {useCallback, useMemo} from 'react';
import {useHistory} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {repoRoute} from '@dash-frontend/views/Project/utils/routes';
import {RepoQuery} from '@graphqlTypes';

interface useBranchBrowserOpts {
  branches?: RepoQuery['repo']['branches'];
}

const useBranchBrowser = ({branches = []}: useBranchBrowserOpts = {}) => {
  const {projectId, repoId, branchId} = useUrlState();
  const browserHistory = useHistory();
  const handleBranchClick = useCallback(
    (branchId: string) => {
      browserHistory.push(
        repoRoute({
          branchId,
          projectId,
          repoId,
        }),
      );
    },
    [browserHistory, projectId, repoId],
  );
  const dropdownItems = useMemo<DropdownItem[]>(() => {
    return [
      {id: 'master', value: 'master', content: 'master', closeOnClick: true},
      ...branches
        .map((branch) => ({
          id: branch.id,
          value: branch.name,
          content: branch.name,
          closeOnClick: true,
        }))
        .filter((branch) => branch.value !== 'master')
        .sort((a, b) => (a.value > b.value ? 1 : -1)),
    ];
  }, [branches]);

  return {handleBranchClick, branchId, dropdownItems};
};

export default useBranchBrowser;
