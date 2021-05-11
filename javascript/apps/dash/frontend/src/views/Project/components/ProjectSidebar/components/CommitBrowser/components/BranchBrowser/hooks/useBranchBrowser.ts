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
  const {dagId, projectId, repoId, branchId} = useUrlState();
  const browserHistory = useHistory();
  const handleBranchClick = useCallback(
    (branchId: string) => {
      browserHistory.push(
        repoRoute({
          branchId,
          dagId,
          projectId,
          repoId,
        }),
      );
    },
    [browserHistory, dagId, projectId, repoId],
  );

  const dropdownItems = useMemo<DropdownItem[]>(() => {
    return [
      {id: 'none', value: 'none', content: 'none', closeOnClick: true},
      ...branches
        .map((branch) => ({
          id: branch.id,
          value: branch.name,
          content: branch.name,
          closeOnClick: true,
        }))
        .sort((a, b) => {
          if (a.value === 'master' || b.value > a.value) return -1;
          if (a.value > b.value) return 1;

          return 0;
        }),
    ];
  }, [branches]);

  return {handleBranchClick, branchId, dropdownItems};
};

export default useBranchBrowser;
