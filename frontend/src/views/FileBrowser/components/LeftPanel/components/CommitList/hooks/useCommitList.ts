import {useState} from 'react';
import {useHistory} from 'react-router';

import {UUID_WITHOUT_DASHES_REGEX} from '@dash-frontend/constants/pachCore';
import {CommitInfo} from '@dash-frontend/generated/proto/pfs/pfs.pb';
import {useCommitSearch} from '@dash-frontend/hooks/useCommitSearch';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {fileBrowserRoute} from '@dash-frontend/views/Project/utils/routes';
const useCommitList = (commits?: CommitInfo[]) => {
  const {projectId, repoId} = useUrlState();
  const browserHistory = useHistory();

  const [searchValue, setSearchValue] = useState('');

  const isSearchValid = UUID_WITHOUT_DASHES_REGEX.test(searchValue);

  const clearSearch = () => {
    setSearchValue('');
  };

  const {commit: searchedCommit, loading: searchLoading} = useCommitSearch(
    {
      projectId,
      repoId,
      commitId: searchValue,
    },
    !isSearchValid,
  );

  const showNoSearchResults =
    !searchLoading && isSearchValid && !searchedCommit;

  const displayCommits = searchedCommit ? [searchedCommit] : commits;

  const updateSelectedCommit = (
    selectedCommitId: string,
    selectedBranchId?: string,
  ) => {
    browserHistory.push(
      fileBrowserRoute({
        projectId,
        repoId,
        branchId: selectedBranchId || '',
        commitId: selectedCommitId,
      }),
    );
  };

  return {
    updateSelectedCommit,
    displayCommits,
    searchValue,
    setSearchValue,
    clearSearch,
    showNoSearchResults,
    isSearchValid,
  };
};

export default useCommitList;
