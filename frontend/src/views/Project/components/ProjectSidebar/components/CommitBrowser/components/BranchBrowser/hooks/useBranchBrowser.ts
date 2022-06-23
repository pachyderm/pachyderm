import {RepoQuery} from '@graphqlTypes';
import {DropdownItem} from '@pachyderm/components';
import {useCallback, useMemo} from 'react';
import {useHistory} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {repoRoute} from '@dash-frontend/views/Project/utils/routes';

interface useBranchBrowserOpts {
  branches: RepoQuery['repo']['branches'];
}

const useBranchBrowser = ({branches}: useBranchBrowserOpts) => {
  const {projectId, repoId, branchId} = useUrlState();
  const browserHistory = useHistory();

  const selectedBranch = useMemo(() => {
    return branchId === 'default' ? branches[0].name : branchId;
  }, [branchId, branches]);

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
      {
        id: selectedBranch,
        value: selectedBranch,
        content: selectedBranch,
        closeOnClick: true,
      },
      ...branches
        .map((branch) => ({
          id: branch.name,
          value: branch.name,
          content: branch.name,
          closeOnClick: true,
        }))
        .filter((branch) => branch.value !== selectedBranch)
        .sort((a, b) => (a.value === 'master' || a.value > b.value ? 1 : -1)),
    ];
  }, [branches, selectedBranch]);

  return {handleBranchClick, dropdownItems, selectedBranch};
};

export default useBranchBrowser;
