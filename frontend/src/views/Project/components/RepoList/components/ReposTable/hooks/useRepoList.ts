import {useCallback, useEffect, useState} from 'react';

import {usePaginatedRepos} from '@dash-frontend/hooks/useRepos';
import useUrlState from '@dash-frontend/hooks/useUrlState';

export const REPOS_DEFAULT_PAGE_SIZE = 15;

const useRepositoriesList = () => {
  const {projectId} = useUrlState();
  const [pageIndex, setPageIndex] = useState(0);
  const [pageSize, setPageSize] = useState(REPOS_DEFAULT_PAGE_SIZE);

  const {repos, loading, error} = usePaginatedRepos(
    projectId,
    pageSize,
    pageIndex,
  );

  const updatePage = useCallback((page: number) => {
    setPageIndex(page - 1);
  }, []);

  useEffect(() => {
    setPageIndex(0);
  }, []);

  return {
    repos,
    loading,
    error,
    pageIndex,
    updatePage,
    pageSize,
    setPageSize,
    hasNextPage: repos?.length === pageSize,
  };
};

export default useRepositoriesList;
