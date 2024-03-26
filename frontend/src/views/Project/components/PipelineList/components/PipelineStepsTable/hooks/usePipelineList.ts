import {useCallback, useEffect, useState} from 'react';

import {ResourceType, useAuthorize} from '@dash-frontend/hooks/useAuthorize';
import {usePaginatedPipelines} from '@dash-frontend/hooks/usePipelines';
import useUrlState from '@dash-frontend/hooks/useUrlState';

export const PIPELINES_DEFAULT_PAGE_SIZE = 15;

const usePipelinesList = () => {
  const {projectId} = useUrlState();
  const [pageIndex, setPageIndex] = useState(0);
  const [pageSize, setPageSize] = useState(PIPELINES_DEFAULT_PAGE_SIZE);

  const {isAuthActive} = useAuthorize({
    permissions: [],
    resource: {
      type: ResourceType.PROJECT,
      name: projectId,
    },
  });

  const {pipelines, loading, error} = usePaginatedPipelines(
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
    pipelines,
    loading,
    error,
    pageIndex,
    updatePage,
    pageSize,
    setPageSize,
    hasNextPage: pipelines?.length === pageSize,
    isAuthActive,
  };
};

export default usePipelinesList;
